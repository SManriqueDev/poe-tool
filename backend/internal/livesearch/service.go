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

	"github.com/SManriqueDev/poe-tool/backend/internal/logging"
	"github.com/SManriqueDev/poe-tool/backend/internal/settings"
	"github.com/corpix/uarand"
)

const workerCount = 10 // Number of concurrent workers

type Service struct {
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
	s := &Service{
		links:        make([]TradeLink, 0),
		settingsSvc:  settingsSvc,
		loggingSvc:   loggingSvc,
		repo:         NewRepository(),
		wsClient:     NewWebSocketClient(),
		eventBus:     &WailsEventBus{},
		linkStatuses: make(map[int]string),
		hideoutQueue: make(chan HideoutQueueItem, 100), // Buffer for 100 items
	}

	// Initialize go_to_hideout setting with default value false if it doesn't exist
	_ = s.repo.InitializeLiveSearchSetting("go_to_hideout", false)

	// Start hideout processing goroutine
	go s.processHideoutQueue()

	return s
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
	link := NewIdleTradeLink(url, description)
	s.links = append(s.links, *link)
	_ = s.repo.AddTradeLink(link.URL, link.Description)
}

func (s *Service) ListTradeLinks() []TradeLink {
	links, err := s.repo.GetTradeLinks()
	if err != nil {
		return []TradeLink{}
	}

	var tradeLinks []TradeLink
	for _, l := range links {
		tl := NewTradeLink(
			WithID(l.ID),
			WithURL(l.URL),
			WithDescription(l.Description),
			WithSelected(l.Selected),
			WithStatus("idle"),
		)
		tradeLinks = append(tradeLinks, *tl)
	}
	return append([]TradeLink{}, tradeLinks...)
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
	return s.repo.UpdateTradeLink(id, url, description, selected)
}

func (s *Service) DeleteTradeLink(id int) error {
	return s.repo.DeleteTradeLink(id)
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
	// Esta función necesita acceso a la aplicación global
	// La implementaremos desde el handler que puede importar main
	return fmt.Errorf("use handler method instead")
}

// HideoutQueueItem represents an item waiting to go to hideout
type HideoutQueueItem struct {
	Token     string
	ItemID    string
	Timestamp time.Time
}

// GoToHideout makes a POST request to Path of Exile whisper API to teleport to hideout
func (s *Service) GoToHideout(hideoutToken string) error {
	// Get POESESSID from settings service
	config := s.settingsSvc.Get()
	if config == nil || config.PoeSessid == "" {
		return fmt.Errorf("POESESSID not configured")
	}

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

// processHideoutQueue processes hideout requests sequentially with proper timing
func (s *Service) processHideoutQueue() {
	for item := range s.hideoutQueue {
		s.hideoutMu.Lock()
		s.hideoutProcessing = true
		s.hideoutMu.Unlock()

		// Log hideout attempt
		s.loggingSvc.Info(logging.LogModuleLiveSearch, "Processing hideout teleport request", map[string]interface{}{
			"item_id":   item.ItemID,
			"timestamp": item.Timestamp,
		})

		// Make hideout request
		err := s.GoToHideout(item.Token)
		if err != nil {
			s.loggingSvc.Error(logging.LogModuleLiveSearch, "Failed to teleport to hideout", map[string]interface{}{
				"item_id": item.ItemID,
				"error":   err.Error(),
				"token":   item.Token[:20] + "...",
			})
		} else {
			s.loggingSvc.Success(logging.LogModuleLiveSearch, "Successfully teleported to hideout", map[string]interface{}{
				"item_id": item.ItemID,
			})
		}

		// Wait before processing next item (configurable delay)
		// This gives time to buy the item before going to the next hideout
		config := s.settingsSvc.Get()
		delay := 5 * time.Second // Default 5 seconds
		if config != nil && config.Delay > 0 {
			delay = time.Duration(config.Delay) * time.Millisecond
		}
		time.Sleep(delay)

		s.hideoutMu.Lock()
		s.hideoutProcessing = false
		s.hideoutMu.Unlock()
	}
}

// QueueHideoutVisit adds a hideout visit to the queue if go_to_hideout is enabled
func (s *Service) QueueHideoutVisit(hideoutToken, itemID string) error {
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

	// Create queue item
	item := HideoutQueueItem{
		Token:     hideoutToken,
		ItemID:    itemID,
		Timestamp: time.Now(),
	}

	// Add to queue (non-blocking)
	select {
	case s.hideoutQueue <- item:
		s.loggingSvc.Debug(logging.LogModuleLiveSearch, "Added hideout visit to queue", map[string]interface{}{
			"item_id":    itemID,
			"queue_size": len(s.hideoutQueue),
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
	return len(s.hideoutQueue)
}

// IsHideoutProcessing returns whether a hideout request is currently being processed
func (s *Service) IsHideoutProcessing() bool {
	s.hideoutMu.Lock()
	defer s.hideoutMu.Unlock()
	return s.hideoutProcessing
}
