package application

import (
	"context"
	"sync"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// LiveSearchApplicationService implementa los casos de uso de LiveSearch
type LiveSearchApplicationService struct {
	tradeLinkRepo   domain.TradeLinkRepository
	liveSearchRepo  domain.LiveSearchRepository
	webSocketClient domain.WebSocketClient
	eventBus        domain.EventBus
	hideoutAutomation domain.HideoutAutomation
	logger          domain.Logger

	// State management
	state      domain.LiveSearchState
	stateMu    sync.RWMutex
	cancelFunc context.CancelFunc

	// Link status tracking
	linkStatuses map[int]string
	statusMu     sync.RWMutex
}

// NewLiveSearchApplicationService crea una nueva instancia del servicio de aplicación
func NewLiveSearchApplicationService(
	tradeLinkRepo domain.TradeLinkRepository,
	liveSearchRepo domain.LiveSearchRepository,
	wsClient domain.WebSocketClient,
	eventBus domain.EventBus,
	logger domain.Logger,
	hideoutAutomation domain.HideoutAutomation,
) *LiveSearchApplicationService {
	return &LiveSearchApplicationService{
		tradeLinkRepo:   tradeLinkRepo,
		liveSearchRepo:  liveSearchRepo,
		webSocketClient: wsClient,
		eventBus:        eventBus,
		logger:          logger,
		hideoutAutomation: hideoutAutomation,
		state:           domain.LiveSearchStopped,
		linkStatuses:    make(map[int]string),
	}
}

// StartLiveSearch inicia la búsqueda en vivo
func (s *LiveSearchApplicationService) StartLiveSearch(ctx context.Context) error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()

	if s.state == domain.LiveSearchRunning {
		return nil // Ya está corriendo
	}

	// Obtener solo los trade links activos (seleccionados) para iniciar búsqueda
	tradeLinks, err := s.tradeLinkRepo.GetActiveTradeLinks(ctx)
	if err != nil {
		s.logger.Error("livesearch", "Failed to get active trade links", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	if len(tradeLinks) == 0 {
		return domain.ErrNoActiveTradeLinks
	}

	// Inicializar el estado de SOLO los links SELECCIONADOS a "connecting"
	// Los links NO seleccionados mantienen su estado actual sin cambios
	for _, link := range tradeLinks {
		s.SetLinkStatus(link.ID, "connecting")
	}

	// Crear contexto cancelable
	liveSearchCtx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel

	// Iniciar búsqueda
	go s.runLiveSearch(liveSearchCtx, tradeLinks)

	s.state = domain.LiveSearchRunning
	s.logger.Info("livesearch", "Live search started", map[string]interface{}{
		"active_links": len(tradeLinks),
	})

	return nil
}

// StopLiveSearch detiene la búsqueda en vivo
func (s *LiveSearchApplicationService) StopLiveSearch(ctx context.Context) error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()

	if s.state != domain.LiveSearchRunning {
		return nil // Ya está detenida
	}

	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}

	// Resetear solo los links que estaban siendo monitoreados (los que están en memoria)
	// Esto preserva el estado de los links no seleccionados
	s.ResetAllLinkStatuses("idle")

	s.state = domain.LiveSearchStopped
	s.logger.Info("livesearch", "Live search stopped", nil)

	return nil
}

// IsLiveSearchRunning verifica si la búsqueda está corriendo
func (s *LiveSearchApplicationService) IsLiveSearchRunning() bool {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	return s.state == domain.LiveSearchRunning
}

// runLiveSearch ejecuta la lógica de búsqueda en vivo
func (s *LiveSearchApplicationService) runLiveSearch(ctx context.Context, tradeLinks []domain.TradeLink) {
	s.logger.Info("livesearch", "Starting live search process", map[string]interface{}{
		"trade_links_count": len(tradeLinks),
	})

	// Emitir evento de inicio
	if err := s.eventBus.EmitLiveSearchStarted(ctx); err != nil {
		s.logger.Error("livesearch", "Failed to emit live search started event", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Obtener POESESSID desde settings repository
	var poeSessID string
	poeSessIDSetting, err := s.liveSearchRepo.GetSetting(ctx, "poesessid")
	if err != nil {
		s.logger.Warning("livesearch", "POESESSID not found in settings, WebSocket may not authenticate", map[string]interface{}{
			"error": err.Error(),
		})
		poeSessID = ""
	} else if poeSessIDSetting == nil {
		s.logger.Warning("livesearch", "POESESSID not set in settings, WebSocket may not authenticate", nil)
		poeSessID = ""
	} else {
		poeSessID = poeSessIDSetting.(string)
	}

	// Configurar POESESSID en el WebSocket client
	if poeSessID != "" {
		s.webSocketClient.SetPOESESSID(poeSessID)
		s.logger.Info("livesearch", "POESESSID configured for WebSocket", map[string]interface{}{
			"poesessid_length": len(poeSessID),
		})
	} else {
		s.logger.Warning("livesearch", "No valid POESESSID found", nil)
	}

	// Conectar WebSocket (preparación)
	if err := s.webSocketClient.Connect(ctx, ""); err != nil {
		s.logger.Error("livesearch", "Failed to prepare WebSocket connections", map[string]interface{}{
			"error": err.Error(),
		})
		// Marcar todos los links como error si no se puede conectar
		for _, link := range tradeLinks {
			s.SetLinkStatus(link.ID, "error")
		}
		return
	}

	// Suscribirse a cada trade link
	for _, link := range tradeLinks {
		go s.processTradeLink(ctx, link)
	}

	// Mantener el bucle de monitoreo activo usando el canal del WebSocket
	s.monitorLiveSearch(ctx, tradeLinks)
}

// processTradeLink procesa un trade link individual
func (s *LiveSearchApplicationService) processTradeLink(ctx context.Context, link domain.TradeLink) {
	s.logger.Info("livesearch", "Processing trade link", map[string]interface{}{
		"link_id":     link.ID,
		"description": link.Description,
		"url":         link.URL,
	})

	// Marcar como conectando
	s.SetLinkStatus(link.ID, "connecting")

	// Extraer search ID de la URL usando lógica de dominio
	searchID, err := domain.ExtractSearchID(link.URL)
	if err != nil {
		s.logger.Error("livesearch", "Could not extract search ID from URL", map[string]interface{}{
			"link_id": link.ID,
			"url":     link.URL,
			"error":   err.Error(),
		})
		s.SetLinkStatus(link.ID, "error")
		return
	}

	// Determinar la liga: primero intentar desde el TradeLink, luego fallback a settings
	league := link.League
	if league == "" {
		if setting, err := s.liveSearchRepo.GetSetting(ctx, "league"); err == nil && setting != nil {
			if leagueStr, ok := setting.(string); ok && leagueStr != "" {
				league = leagueStr
				s.logger.Info("livesearch", "Using league from settings", map[string]interface{}{
					"link_id": link.ID,
					"league":  league,
				})
			}
		}
	}

	if league == "" {
		s.logger.Error("livesearch", "No league available for subscription", map[string]interface{}{
			"link_id": link.ID,
		})
		s.SetLinkStatus(link.ID, "error")
		return
	}

	// Suscribirse al search ID con la liga determinada
	if err := s.webSocketClient.Subscribe(ctx, searchID, league); err != nil {
		s.logger.Error("livesearch", "Failed to subscribe to search", map[string]interface{}{
			"link_id":   link.ID,
			"search_id": searchID,
			"league":    league,
			"error":     err.Error(),
		})
		s.SetLinkStatus(link.ID, "error")
		return
	}

	// Conexión exitosa, cambiar a estado conectado antes de monitoreo
	s.SetLinkStatus(link.ID, "connected")
	s.SetLinkStatus(link.ID, "monitoring")
	s.logger.Info("livesearch", "Trade link monitoring started", map[string]interface{}{
		"link_id":   link.ID,
		"search_id": searchID,
		"league":    league,
	})
}

// monitorLiveSearch mantiene el monitoreo activo de los trade links
func (s *LiveSearchApplicationService) monitorLiveSearch(ctx context.Context, tradeLinks []domain.TradeLink) {
	s.logger.Info("livesearch", "Starting continuous monitoring", map[string]interface{}{
		"active_links": len(tradeLinks),
	})

	// Obtener canal de mensajes del WebSocket
	msgChannel := s.webSocketClient.GetMessageChannel()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("livesearch", "Monitoring cancelled", nil)
			return
		case item, ok := <-msgChannel:
			if !ok {
				s.logger.Info("livesearch", "Message channel closed", nil)
				return
			}
			s.handleNewItem(ctx, item)
		}
	}
}

// handleNewItem procesa un nuevo item encontrado
func (s *LiveSearchApplicationService) handleNewItem(ctx context.Context, item domain.ItemResult) {
	s.logger.Info("livesearch", "New item found", map[string]interface{}{
		"item_id":   item.ID,
		"search_id": item.SearchID,
	})

	// Extract hideout_token from listing and queue hideout visit if available
	if listingMap, ok := item.Listing.(map[string]interface{}); ok {
		if hideoutToken, ok := listingMap["hideout_token"].(string); ok && hideoutToken != "" {
			// Skip if a hideout visit is already in progress or queued
			isProcessing, _ := s.hideoutAutomation.IsProcessing(ctx)
			queueSize, _ := s.hideoutAutomation.GetQueueSize(ctx)
			if isProcessing || queueSize > 0 {
				s.logger.Info("livesearch", "Skipping hideout - another visit in progress", map[string]interface{}{
					"item_id":    item.ID,
					"processing": isProcessing,
					"queue_size": queueSize,
				})
			} else {
				s.logger.Info("livesearch", "Hideout token found, queuing visit", map[string]interface{}{
					"item_id": item.ID,
				})
				if err := s.hideoutAutomation.QueueHideoutVisit(ctx, hideoutToken, item.ID); err != nil {
					s.logger.Error("livesearch", "Failed to queue hideout visit", map[string]interface{}{
						"item_id": item.ID,
						"error":   err.Error(),
					})
				}
			}
		}
	}

	// Emitir evento de nuevo item
	items := []domain.ItemResult{item}
	if err := s.eventBus.EmitNewItems(ctx, item.SearchID, items); err != nil {
		s.logger.Error("livesearch", "Failed to emit new items", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// GetActiveTradeLinksCount retorna el número de trade links activos
func (s *LiveSearchApplicationService) GetActiveTradeLinksCount(ctx context.Context) (int, error) {
	tradeLinks, err := s.tradeLinkRepo.GetActiveTradeLinks(ctx)
	if err != nil {
		return 0, err
	}
	return len(tradeLinks), nil
}

// GetAllTradeLinks retorna todos los trade links para mantener compatibilidad con handler
func (s *LiveSearchApplicationService) GetAllTradeLinks(ctx context.Context) ([]domain.TradeLink, error) {
	return s.tradeLinkRepo.GetAll(ctx)
}

// GetAllLinkStatuses retorna el estado actual de todos los trade links
func (s *LiveSearchApplicationService) GetAllLinkStatuses() map[int]string {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()

	// Si el mapa está vacío, inicializarlo una sola vez con "idle" para todos
	if len(s.linkStatuses) == 0 {
		ctx := context.Background()
		links, err := s.tradeLinkRepo.GetAll(ctx)
		if err == nil {
			for _, link := range links {
				s.linkStatuses[link.ID] = "idle"
			}
		}
	}

	// Crear una copia del mapa para evitar race conditions
	result := make(map[int]string)
	for id, status := range s.linkStatuses {
		result[id] = status
	}

	return result
}

// SetLinkStatus actualiza el estado de un trade link específico
func (s *LiveSearchApplicationService) SetLinkStatus(linkID int, status string) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()

	s.linkStatuses[linkID] = status

	// Emitir evento de cambio de estado
	if err := s.eventBus.EmitLinkStatusChanged(context.Background(), linkID, status); err != nil {
		s.logger.Error("livesearch", "Failed to emit link status changed", map[string]interface{}{
			"link_id": linkID,
			"status":  status,
			"error":   err.Error(),
		})
	}
}

// ResetAllLinkStatuses reinicia el estado de todos los links
func (s *LiveSearchApplicationService) ResetAllLinkStatuses(status string) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()

	for id := range s.linkStatuses {
		s.linkStatuses[id] = status

		// Emitir evento de cambio de estado para cada link
		if err := s.eventBus.EmitLinkStatusChanged(context.Background(), id, status); err != nil {
			s.logger.Error("livesearch", "Failed to emit link status changed in ResetAllLinkStatuses", map[string]interface{}{
				"link_id": id,
				"status":  status,
				"error":   err.Error(),
			})
		}
	}
}
