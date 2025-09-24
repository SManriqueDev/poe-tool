package livesearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
	"github.com/SManriqueDev/poe-tool/backend/internal/logging"
	"github.com/SManriqueDev/poe-tool/backend/internal/settings"
	"github.com/corpix/uarand"
)

const workerCount = 10 // Number of concurrent workers

type Service struct {
	// Dependency managers
	tradeLinkMgr *TradeLinkManager

	// Fase 4: Domain components
	windowManager     domain.WindowManager
	hideoutAutomation domain.HideoutAutomation
	systemAPIClient   domain.SystemAPIClient

	// Legacy fields (serán eliminados en Fase 5)
	links             []TradeLink
	mu                sync.Mutex
	settingsSvc       *settings.Service
	loggingSvc        *logging.Service
	liveSearchCancel  context.CancelFunc
	ctx               context.Context
	repo              *Repository
	wsClient          *WebSocketClient
	eventBus          EventBus
	liveSearchWG      *sync.WaitGroup
	linkStatuses      map[int]string // In-memory storage for current link statuses
	statusMu          sync.RWMutex   // Separate mutex for status operations
	hideoutQueue      chan HideoutQueueItem
	hideoutProcessing bool
	hideoutMu         sync.Mutex
	processedItems    map[string]bool // Track processed items to avoid duplicates
	processedItemsMu  sync.RWMutex    // Mutex for processed items map
}

type WSMessage struct {
	SearchID string
	Message  []byte
}

// WebSocket message structure from PoE API
type LiveSearchMessage struct {
	New []string `json:"new"`
}

// Item fetch response from PoE API
type ItemFetchResponse struct {
	Result []ItemResult `json:"result"`
}

type ItemResult struct {
	ID      string          `json:"id"`
	Item    json.RawMessage `json:"item"`
	Listing json.RawMessage `json:"listing"`
}

// Listing structure from PoE API that contains hideout_token
type ListingData struct {
	Method       string      `json:"method"`
	Indexed      string      `json:"indexed"`
	Stash        StashInfo   `json:"stash"`
	HideoutToken string      `json:"hideout_token,omitempty"`
	Whisper      string      `json:"whisper,omitempty"`
	Account      AccountInfo `json:"account"`
	Price        PriceInfo   `json:"price"`
	Fee          int         `json:"fee,omitempty"`
}

// StashInfo represents stash information from PoE API
type StashInfo struct {
	Name string `json:"name"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

// AccountInfo represents account information from PoE API
type AccountInfo struct {
	Name   string      `json:"name"`
	Online interface{} `json:"online"` // Can be null or other values
}

// PriceInfo represents price information from PoE API
type PriceInfo struct {
	Type     string  `json:"type"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

func (s *Service) SetContext(ctx context.Context) {
	s.ctx = ctx
	// También configurar el contexto en el servicio de logging
	s.loggingSvc.SetContext(ctx)
}

// fetchItemDetails fetches item details from PoE API for given item IDs
func (s *Service) fetchItemDetails(itemIDs []string, searchId string) (*ItemFetchResponse, error) {
	if len(itemIDs) == 0 {
		return &ItemFetchResponse{}, nil
	}

	// Limit the number of IDs per request (PoE API typically has limits)
	const maxItemsPerRequest = 10
	if len(itemIDs) > maxItemsPerRequest {
		itemIDs = itemIDs[:maxItemsPerRequest]
		log.Printf("Limiting request to %d items (max allowed)", maxItemsPerRequest)
	}

	// Join IDs with commas for the API call
	idsParam := strings.Join(itemIDs, ",")
	url := fmt.Sprintf("https://www.pathofexile.com/api/trade2/fetch/%s", idsParam)

	// Add searchId parameter
	url += "?query=" + searchId + "&realm=poe2"

	// Get PoE session from settings
	cfg := s.settingsSvc.Get()
	poeSess := cfg.PoeSessid

	if poeSess == "" {
		return nil, fmt.Errorf("PoE session ID is not configured")
	}

	// Create HTTP request with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add required headers
	req.Header.Set("Cookie", "POESESSID="+poeSess)
	req.Header.Set("User-Agent", uarand.GetRandom())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://www.pathofexile.com")

	// Make the request with timing
	startTime := time.Now()
	client := &http.Client{}
	resp, err := client.Do(req)
	responseTime := time.Since(startTime)

	// Log the API call
	var errorMessage string
	if err != nil {
		errorMessage = err.Error()
		s.loggingSvc.LogAPICall(url, "GET", 0, responseTime, errorMessage)
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode == http.StatusUnauthorized {
		errorMessage = "authentication failed - check PoE session ID"
		s.loggingSvc.LogAPICall(url, "GET", resp.StatusCode, responseTime, errorMessage)
		return nil, fmt.Errorf("%s", errorMessage)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errorMessage = fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body))
		s.loggingSvc.LogAPICall(url, "GET", resp.StatusCode, responseTime, errorMessage)
		return nil, fmt.Errorf("%s", errorMessage)
	}

	// Log successful API call
	s.loggingSvc.LogAPICall(url, "GET", resp.StatusCode, responseTime, "")

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON response
	var fetchResp ItemFetchResponse
	if err := json.Unmarshal(body, &fetchResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &fetchResp, nil
}

// logNewItem creates a log entry for a new item found during live search
func (s *Service) logNewItem(item ItemResult, searchID string, tradeLink *TradeLink) {
	// Extract item name from the item JSON
	var itemData map[string]interface{}
	if err := json.Unmarshal(item.Item, &itemData); err != nil {
		log.Printf("Failed to parse item JSON for logging: %v", err)
		return
	}

	// Extract listing data for price information
	var listingData map[string]interface{}
	if err := json.Unmarshal(item.Listing, &listingData); err != nil {
		log.Printf("Failed to parse listing JSON for logging: %v", err)
		listingData = nil
	}

	// Extract item name
	itemName := "Unknown Item"
	if name, ok := itemData["name"].(string); ok && name != "" {
		itemName = name
	} else if typeLine, ok := itemData["typeLine"].(string); ok {
		itemName = typeLine
	}

	// Extract price information
	var price *logging.Price
	if listingData != nil {
		if priceData, ok := listingData["price"].(map[string]interface{}); ok {
			if amount, ok := priceData["amount"].(float64); ok {
				if currency, ok := priceData["currency"].(string); ok {
					priceType := "exact"
					if pType, ok := priceData["type"].(string); ok {
						priceType = pType
					}
					price = &logging.Price{
						Amount:   amount,
						Currency: currency,
						Type:     priceType,
					}
				}
			}
		}
	}

	// Get league and search URL
	league := "Unknown"
	searchURL := ""
	if tradeLink != nil {
		league = tradeLink.League()
		searchURL = tradeLink.URL
	}

	// Create item details string
	itemDetails := fmt.Sprintf("Item ID: %s", item.ID)
	if corrupted, ok := itemData["corrupted"].(bool); ok && corrupted {
		itemDetails += " (Corrupted)"
	}
	if identified, ok := itemData["identified"].(bool); ok && !identified {
		itemDetails += " (Unidentified)"
	}

	// Log the item
	err := s.loggingSvc.LogItemFound(searchID, item.ID, itemName, league, searchURL, price, itemDetails)
	if err != nil {
		log.Printf("Failed to log new item: %v", err)
	}
}

func NewService(settingsSvc *settings.Service, loggingSvc *logging.Service) *Service {
	repo := NewRepository()

	s := &Service{
		// Initialize dependency managers
		tradeLinkMgr: NewTradeLinkManager(WithTradeLinkRepository(repo)),

		// Phase 4 - Domain components will be injected via SetDomainComponents
		windowManager:     nil,
		hideoutAutomation: nil,
		systemAPIClient:   nil,

		// Legacy fields (will be refactored in future phases)
		links:          make([]TradeLink, 0),
		settingsSvc:    settingsSvc,
		loggingSvc:     loggingSvc,
		repo:           repo,
		wsClient:       NewWebSocketClient(),
		eventBus:       &WailsEventBus{},
		linkStatuses:   make(map[int]string),
		hideoutQueue:   make(chan HideoutQueueItem, 100), // Buffer for 100 items
		processedItems: make(map[string]bool),            // Initialize processed items tracker
	}

	// Initialize go_to_hideout setting with default value false if it doesn't exist
	_ = s.repo.InitializeLiveSearchSetting("go_to_hideout", false)

	// Start hideout processing goroutine
	go s.processHideoutQueue()

	return s
}

// SetDomainComponents injects Phase 4 domain components
func (s *Service) SetDomainComponents(windowManager domain.WindowManager, hideoutAutomation domain.HideoutAutomation, systemAPIClient domain.SystemAPIClient) {
	s.windowManager = windowManager
	s.hideoutAutomation = hideoutAutomation
	s.systemAPIClient = systemAPIClient
}

// GetRepository retorna el repositorio para permitir crear adaptadores
// TODO: Este método será eliminado una vez completada la migración a Clean Architecture
func (s *Service) GetRepository() *Repository {
	return s.repo
}

// GetEventBus retorna el event bus para permitir crear adaptadores
// TODO: Este método será eliminado una vez completada la migración a Clean Architecture
func (s *Service) GetEventBus() EventBus {
	return s.eventBus
}

// SetupEventEmitter configures the event emitter for real-time log updates
func (s *Service) SetupEventEmitter(loggingSvc *logging.Service) {
	loggingSvc.SetEventEmitter(s.eventBus)
}

// TestLogEvent creates a test log to verify real-time events are working
func (s *Service) TestLogEvent() error {
	return s.loggingSvc.Log(logging.LogModuleLiveSearch, logging.LogLevelInfo, "Test log event for real-time updates", map[string]interface{}{
		"test":      true,
		"timestamp": time.Now(),
	})
}

func (s *Service) AddTradeLink(url string, description string) {
	// Delegate to TradeLinkManager
	if err := s.tradeLinkMgr.Add(url, description); err != nil {
		// Log error but maintain backward compatibility with original void return
		s.loggingSvc.Error("livesearch", "Failed to add trade link", map[string]interface{}{
			"url":         url,
			"description": description,
			"error":       err.Error(),
		})
	}
}

func (s *Service) ListTradeLinks() []TradeLink {
	// Delegate to TradeLinkManager
	links, err := s.tradeLinkMgr.List()
	if err != nil {
		// Log error and return empty slice for backward compatibility
		s.loggingSvc.Error("livesearch", "Failed to list trade links", map[string]interface{}{
			"error": err.Error(),
		})
		return []TradeLink{}
	}
	return links
}

func (s *Service) StartLiveSearch() []TradeLink {
	cfg := s.settingsSvc.Get()
	poeSess := cfg.PoeSessid

	links, err := s.repo.GetTradeLinks()
	if err != nil {
		return []TradeLink{}
	}

	// Cancelar búsqueda previa si estaba corriendo
	s.mu.Lock()
	if s.liveSearchCancel != nil {
		s.liveSearchCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.liveSearchCancel = cancel
	s.mu.Unlock()

	// Clear processed items for new search session
	s.processedItemsMu.Lock()
	s.processedItems = make(map[string]bool)
	s.processedItemsMu.Unlock()

	// Cleanup old processed items periodically
	s.cleanupOldProcessedItems()

	// Filtrar links seleccionados
	var selectedLinks []TradeLink
	for _, link := range links {
		if link.Selected {
			selectedLinks = append(selectedLinks, link)
		}
	}
	if len(selectedLinks) == 0 {
		return links
	}

	// Copia inicial de los links
	statusLinks := make([]TradeLink, len(links))
	copy(statusLinks, links)

	// Canal para mensajes de sockets
	msgCh := make(chan WSMessage, 500)
	// Canal para actualizaciones de estado
	statusCh := make(chan func(), 100)

	// Workers para mensajes
	var workerWg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for msg := range msgCh {
				// Parse the WebSocket message
				var liveMsg LiveSearchMessage
				if err := json.Unmarshal(msg.Message, &liveMsg); err != nil {
					log.Printf("Failed to parse WebSocket message for %s: %v", msg.SearchID, err)
					continue
				}

				// Check if there are new items
				if len(liveMsg.New) == 0 {
					log.Printf("No new items in message for %s", msg.SearchID)
					continue
				}

				log.Printf("Found %d new items for search %s: %v", len(liveMsg.New), msg.SearchID, liveMsg.New)

				// Fetch item details from PoE API
				itemResp, err := s.fetchItemDetails(liveMsg.New, msg.SearchID)
				if err != nil {
					log.Printf("Failed to fetch item details for search %s: %v", msg.SearchID, err)
					continue
				}

				log.Printf("Successfully fetched %d items for search %s", len(itemResp.Result), msg.SearchID)

				// Emit event with new items for frontend
				s.eventBus.EmitNewItems(s.ctx, msg.SearchID, itemResp.Result)

				// Find the trade link for more context
				var tradeLink *TradeLink
				for _, link := range links {
					if link.SearchID() == msg.SearchID {
						tradeLink = &link
						break
					}
				}

				// Process each item and log it
				for _, item := range itemResp.Result {
					// Check if item was already processed
					s.processedItemsMu.RLock()
					alreadyProcessed := s.processedItems[item.ID]
					s.processedItemsMu.RUnlock()

					if alreadyProcessed {
						s.loggingSvc.Debug(logging.LogModuleLiveSearch, "Skipping duplicate item", map[string]interface{}{
							"item_id":   item.ID,
							"search_id": msg.SearchID,
						})
						continue
					}

					// Mark item as processed
					s.processedItemsMu.Lock()
					s.processedItems[item.ID] = true
					s.processedItemsMu.Unlock()

					s.logNewItem(item, msg.SearchID, tradeLink)

					// Process hideout token if available
					s.processItemForHideout(item)
				}
			}
		}()
	}

	// Goroutine única para manejar actualizaciones de estado
	go func() {
		for update := range statusCh {
			update()
		}
	}()

	// WaitGroup de conexiones
	s.liveSearchWG = &sync.WaitGroup{}

	for i, link := range links {
		if !link.Selected {
			continue
		}
		s.liveSearchWG.Add(1)
		go func(idx int, link TradeLink) {
			defer s.liveSearchWG.Done()

			conn, resp, err := s.wsClient.Connect(ctx, link, poeSess)
			if err != nil {
				log.Printf("WebSocket connection error for %s: %v", link.URL, err)
				statusCh <- func() {
					if resp != nil && resp.StatusCode == http.StatusUnauthorized {
						statusLinks[idx].Status = "auth_error"
						s.UpdateLinkStatus(statusLinks[idx].ID, "auth_error")
					} else {
						statusLinks[idx].Status = "error"
						s.UpdateLinkStatus(statusLinks[idx].ID, "error")
					}
					s.eventBus.EmitStatusUpdate(s.ctx, statusLinks[idx])
				}
				return
			}

			defer func() {
				_ = conn.Close()
			}()

			// Autenticación inicial
			var authResp struct {
				Auth bool `json:"auth"`
			}
			if err := conn.ReadJSON(&authResp); err != nil || !authResp.Auth {
				statusCh <- func() {
					statusLinks[idx].Status = "auth_error"
					s.UpdateLinkStatus(statusLinks[idx].ID, "auth_error")
					s.eventBus.EmitStatusUpdate(s.ctx, statusLinks[idx])
				}
				return
			}

			// OK → actualizar estado
			statusCh <- func() {
				statusLinks[idx].Status = "ok"
				s.UpdateLinkStatus(statusLinks[idx].ID, "ok")
				s.eventBus.EmitStatusUpdate(s.ctx, statusLinks[idx])
			}

			// Bucle de lectura
			for {
				select {
				case <-ctx.Done():
					statusCh <- func() {
						statusLinks[idx].Status = "idle"
						s.UpdateLinkStatus(statusLinks[idx].ID, "idle")
						s.eventBus.EmitStatusUpdate(s.ctx, statusLinks[idx])
					}
					return
				default:
					_, message, err := conn.ReadMessage()
					if err != nil {
						log.Printf("WebSocket read error for %s: %v", link.URL, err)
						return
					}
					select {
					case msgCh <- WSMessage{SearchID: link.SearchID(), Message: message}:
					default:
						log.Printf("msgCh full, dropping message for %s", link.SearchID())
					}
				}
			}
		}(i, link)
	}

	// Goroutine para limpiar al terminar todas las conexiones
	go func() {
		s.liveSearchWG.Wait()
		close(msgCh) // cerrar workers
		workerWg.Wait()
		close(statusCh) // cerrar actualizador de estado
	}()

	// Retornar inmediatamente (queda corriendo en background)
	return statusLinks
}

func (s *Service) IsLiveSearchRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.liveSearchCancel != nil
}

func (s *Service) StopLiveSearch() {
	s.mu.Lock()
	if s.liveSearchCancel != nil {
		s.liveSearchCancel()
		s.liveSearchCancel = nil
	}
	links, err := s.repo.GetTradeLinks()
	if err == nil {
		for _, link := range links {
			link.Status = "idle"
			s.UpdateLinkStatus(link.ID, "idle")
			s.eventBus.EmitStatusUpdate(s.ctx, link)
		}
	}
	s.mu.Unlock()
}

func (s *Service) UpdateTradeLink(id int, url string, description string, selected bool) error {
	// Delegate to TradeLinkManager
	return s.tradeLinkMgr.Update(id, url, description, selected)
}

func (s *Service) DeleteTradeLink(id int) error {
	// Delegate to TradeLinkManager
	return s.tradeLinkMgr.Delete(id)
}

func (s *Service) SetGoToHideout(enabled bool) error {
	return s.repo.UpdateLiveSearchSetting("go_to_hideout", enabled)
}

func (s *Service) GetGoToHideout() (bool, error) {
	return s.repo.GetLiveSearchSetting("go_to_hideout")
}

// UpdateLinkStatus updates the status of a link in memory
func (s *Service) UpdateLinkStatus(linkID int, status string) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	s.linkStatuses[linkID] = status
}

// GetLinkStatus gets the status of a specific link
func (s *Service) GetLinkStatus(linkID int) string {
	s.statusMu.RLock()
	defer s.statusMu.RUnlock()
	status, exists := s.linkStatuses[linkID]
	if !exists {
		return "idle"
	}
	return status
}

// GetAllLinkStatuses returns a map of all current link statuses
func (s *Service) GetAllLinkStatuses() map[int]string {
	s.statusMu.RLock()
	defer s.statusMu.RUnlock()

	// Create a copy to avoid race conditions
	statuses := make(map[int]string)
	for id, status := range s.linkStatuses {
		statuses[id] = status
	}
	return statuses
}

// ClearLinkStatuses resets all link statuses to idle
func (s *Service) ClearLinkStatuses() {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	s.linkStatuses = make(map[int]string)
}

// OpenLogsWindow abre una nueva ventana mostrando los logs de LiveSearch
func (s *Service) OpenLogsWindow() error {
	// Use Phase 4 domain window manager if available
	if s.windowManager != nil {
		return s.windowManager.OpenLogsWindow(s.ctx)
	}
	// Legacy fallback
	return fmt.Errorf("window manager not available - use handler method instead")
}

// HideoutQueueItem represents an item waiting to go to hideout
type HideoutQueueItem struct {
	Token     string
	ItemID    string
	Timestamp time.Time
	Priority  bool // High priority items skip delay
}

// GoToHideout makes a POST request to Path of Exile whisper API to teleport to hideout
func (s *Service) GoToHideout(hideoutToken string) error {
	// Get POESESSID from settings service
	config := s.settingsSvc.Get()
	if config == nil || config.PoeSessid == "" {
		return fmt.Errorf("POESESSID not configured")
	}

	// Use Phase 4 domain system API client if available
	if s.systemAPIClient != nil {
		return s.systemAPIClient.SendHideoutRequest(s.ctx, hideoutToken, config.PoeSessid)
	}

	// Legacy fallback implementation
	// Prepare request body
	requestBody := map[string]interface{}{
		"token":    hideoutToken,
		"continue": true,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://www.pathofexile.com/api/trade2/whisper", bytes.NewBuffer(bodyBytes))

	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers similar to the browser request you provided
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", uarand.GetRandom())
	req.Header.Set("Cookie", "POESESSID="+config.PoeSessid)

	// Make request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("failed to make hideout request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read hideout response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		// Try to parse error response
		var errorResp struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}

		errorMessage := fmt.Sprintf("HTTP %d", resp.StatusCode)
		if json.Unmarshal(respBody, &errorResp) == nil && errorResp.Error.Message != "" {
			errorMessage = errorResp.Error.Message
		}

		return fmt.Errorf("hideout request failed (status %d): %s", resp.StatusCode, errorMessage)
	}

	s.loggingSvc.Debug(logging.LogModuleLiveSearch, "Successfully sent hideout teleport request", map[string]interface{}{
		"token": hideoutToken[:20] + "...", // Log only first 20 chars for security
	})

	return nil
}

// isTemporaryHideoutError determines if a hideout error is temporary and shouldn't affect LiveSearch
func (s *Service) isTemporaryHideoutError(err error) bool {
	if err == nil {
		return false
	}

	errorMsg := err.Error()

	// Temporary/service errors that shouldn't stop LiveSearch
	temporaryErrors := []string{
		"status 503",              // Service Temporarily Unavailable
		"status 429",              // Too Many Requests
		"status 500",              // Internal Server Error
		"Temporarily Unavailable", // PoE API message
		"timeout",                 // Network timeouts
		"connection refused",      // Connection issues
		"no such host",            // DNS issues
	}

	for _, tempError := range temporaryErrors {
		if strings.Contains(errorMsg, tempError) {
			return true
		}
	}

	return false
}

// cleanupOldProcessedItems removes old entries from processed items map to prevent memory leaks
func (s *Service) cleanupOldProcessedItems() {
	s.processedItemsMu.Lock()
	defer s.processedItemsMu.Unlock()

	// If map gets too large, clear it (items are usually processed quickly)
	if len(s.processedItems) > 10000 {
		s.processedItems = make(map[string]bool)
		s.loggingSvc.Debug(logging.LogModuleLiveSearch, "Cleaned up processed items cache", map[string]interface{}{
			"reason": "Cache size exceeded 10000 items",
		})
	}
}

// processHideoutQueue processes hideout requests sequentially with optimal competitive timing
func (s *Service) processHideoutQueue() {
	for item := range s.hideoutQueue {
		s.hideoutMu.Lock()
		s.hideoutProcessing = true
		s.hideoutMu.Unlock()

		// Log hideout attempt
		s.loggingSvc.Info(logging.LogModuleLiveSearch, "Processing hideout teleport request", map[string]interface{}{
			"item_id":          item.ItemID,
			"timestamp":        item.Timestamp,
			"priority":         item.Priority,
			"competitive_mode": true,
		})

		// Make hideout request
		err := s.GoToHideout(item.Token)
		if err != nil {
			// Determine if this is a temporary error that shouldn't affect LiveSearch status
			isTemporaryError := s.isTemporaryHideoutError(err)

			logLevel := "Error"
			if isTemporaryError {
				logLevel = "Warning" // Temporary errors are warnings, not critical errors
			}

			s.loggingSvc.Error(logging.LogModuleLiveSearch, fmt.Sprintf("Failed to teleport to hideout (%s)", logLevel), map[string]interface{}{
				"item_id":          item.ItemID,
				"error":            err.Error(),
				"token":            item.Token[:20] + "...",
				"temporary_error":  isTemporaryError,
				"continues_search": isTemporaryError,
			})

			// If it's a temporary error, log that LiveSearch continues normally
			if isTemporaryError {
				s.loggingSvc.Info(logging.LogModuleLiveSearch, "LiveSearch continues despite hideout error", map[string]interface{}{
					"item_id": item.ItemID,
					"reason":  "Temporary hideout service issue",
				})
			}
		} else {
			s.loggingSvc.Success(logging.LogModuleLiveSearch, "Successfully teleported to hideout", map[string]interface{}{
				"item_id": item.ItemID,
			})
		}

		// Realistic delay considering game loading screens and trading time
		if item.Priority {
			// Realistic delay for competitive teleportation considering:
			// - Game loading screen: 1-3 seconds
			// - Finding player in hideout: 1-2 seconds
			// - Trade negotiation/completion: 3-5 seconds
			// - Buffer for next teleport: 1-2 seconds
			// Total: ~8-10 seconds minimum for realistic trading
			time.Sleep(8 * time.Second)
		} else {
			// Fallback delay (should not occur in competitive mode)
			config := s.settingsSvc.Get()
			delay := 1 * time.Second
			if config != nil && config.Delay > 0 {
				delay = time.Duration(config.Delay) * time.Millisecond
			}
			time.Sleep(delay)
		}

		s.hideoutMu.Lock()
		s.hideoutProcessing = false
		s.hideoutMu.Unlock()
	}
} // QueueHideoutVisit adds a hideout visit to the queue if go_to_hideout is enabled
func (s *Service) QueueHideoutVisit(hideoutToken, itemID string) error {
	// Use Phase 4 domain hideout automation if available
	if s.hideoutAutomation != nil {
		return s.hideoutAutomation.QueueHideoutVisit(s.ctx, hideoutToken, itemID)
	}

	// Legacy implementation fallback
	// Check if go_to_hideout is enabled
	enabled, err := s.GetGoToHideout()
	if err != nil {
		return fmt.Errorf("failed to check go_to_hideout setting: %w", err)
	}

	if !enabled {
		return nil // Feature is disabled, ignore request
	}

	if hideoutToken == "" {
		return fmt.Errorf("hideout token is empty")
	}

	// Create queue item with priority for ALL items (competitive scenario)
	item := HideoutQueueItem{
		Token:     hideoutToken,
		ItemID:    itemID,
		Timestamp: time.Now(),
		Priority:  true, // ALL items get priority for maximum competitiveness
	}

	// Add to queue (non-blocking)
	select {
	case s.hideoutQueue <- item:
		s.loggingSvc.Info(logging.LogModuleLiveSearch, "Added hideout visit to queue", map[string]interface{}{
			"item_id":          itemID,
			"queue_size":       len(s.hideoutQueue),
			"priority":         true,
			"competitive_mode": true,
			"delay_seconds":    8, // Realistic delay considering game loading and trading time
		})
		return nil
	default:
		return fmt.Errorf("hideout queue is full")
	}
}

// processItemForHideout extracts hideout_token from item listing and queues hideout visit
func (s *Service) processItemForHideout(item ItemResult) {
	// Check if go_to_hideout is enabled first
	enabled, err := s.GetGoToHideout()
	if err != nil {
		return
	}

	if !enabled {
		return
	}

	// Parse the listing JSON to extract hideout_token
	var listing ListingData
	err = json.Unmarshal(item.Listing, &listing)
	if err != nil {
		s.loggingSvc.Debug(logging.LogModuleLiveSearch, "Failed to parse listing data", map[string]interface{}{
			"item_id": item.ID,
			"error":   err.Error(),
		})
		return
	}

	// If hideout_token is available, queue the hideout visit
	if listing.HideoutToken != "" {
		err := s.QueueHideoutVisit(listing.HideoutToken, item.ID)
		if err != nil {
			s.loggingSvc.Warning(logging.LogModuleLiveSearch, "Failed to queue hideout visit", map[string]interface{}{
				"item_id": item.ID,
				"error":   err.Error(),
			})
		}
	}
}

// GetHideoutQueueSize returns the current number of items in the hideout queue
func (s *Service) GetHideoutQueueSize() int {
	// Use Phase 4 domain hideout automation if available
	if s.hideoutAutomation != nil {
		size, err := s.hideoutAutomation.GetQueueSize(s.ctx)
		if err != nil {
			s.loggingSvc.Warning(logging.LogModuleLiveSearch, "Failed to get hideout queue size", map[string]interface{}{
				"error": err.Error(),
			})
			return 0
		}
		return size
	}
	// Legacy fallback
	return len(s.hideoutQueue)
}

// IsHideoutProcessing returns whether a hideout request is currently being processed
func (s *Service) IsHideoutProcessing() bool {
	// Use Phase 4 domain hideout automation if available
	if s.hideoutAutomation != nil {
		processing, err := s.hideoutAutomation.IsProcessing(s.ctx)
		if err != nil {
			s.loggingSvc.Warning(logging.LogModuleLiveSearch, "Failed to check hideout processing status", map[string]interface{}{
				"error": err.Error(),
			})
			return false
		}
		return processing
	}
	// Legacy fallback
	s.hideoutMu.Lock()
	defer s.hideoutMu.Unlock()
	return s.hideoutProcessing
}
