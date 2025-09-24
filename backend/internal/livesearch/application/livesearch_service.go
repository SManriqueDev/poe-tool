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
) *LiveSearchApplicationService {
	return &LiveSearchApplicationService{
		tradeLinkRepo:   tradeLinkRepo,
		liveSearchRepo:  liveSearchRepo,
		webSocketClient: wsClient,
		eventBus:        eventBus,
		logger:          logger,
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

	// Obtener trade links activos
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

	// Inicializar el estado de todos los links
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

	// Reiniciar el estado de todos los links a "idle"
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

	// Delegar al servicio WebSocket (que ya implementa toda la lógica funcional)
	// El adapter internamente usa el servicio legacy que maneja todas las conexiones
	if err := s.webSocketClient.Connect(ctx, ""); err != nil {
		s.logger.Error("livesearch", "Failed to start WebSocket connections", map[string]interface{}{
			"error": err.Error(),
		})
		// Continuar con el estado de monitoreo aunque haya error de conexión
	}

	// Mantener el bucle de monitoreo activo usando el canal del WebSocket
	s.monitorLiveSearch(ctx, tradeLinks)
}

// processTradeLink procesa un trade link individual
func (s *LiveSearchApplicationService) processTradeLink(ctx context.Context, link domain.TradeLink) {
	s.logger.Info("livesearch", "Processing trade link", map[string]interface{}{
		"link_id":     link.ID,
		"description": link.Description,
	})

	// Conectar WebSocket
	s.SetLinkStatus(link.ID, "connecting")

	if err := s.webSocketClient.Connect(ctx, link.URL); err != nil {
		s.logger.Error("livesearch", "Failed to connect WebSocket", map[string]interface{}{
			"link_id": link.ID,
			"error":   err.Error(),
		})
		s.SetLinkStatus(link.ID, "error")
		return
	}

	// Suscribirse a actualizaciones
	searchID := s.extractSearchID(link.URL)
	if err := s.webSocketClient.Subscribe(ctx, searchID); err != nil {
		s.logger.Error("livesearch", "Failed to subscribe to search", map[string]interface{}{
			"link_id":   link.ID,
			"search_id": searchID,
			"error":     err.Error(),
		})
		s.SetLinkStatus(link.ID, "error")
		return
	}

	s.SetLinkStatus(link.ID, "monitoring")
	s.logger.Info("livesearch", "Trade link monitoring started", map[string]interface{}{
		"link_id":   link.ID,
		"search_id": searchID,
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
		"item_id": item.ID,
	})

	// Emitir evento de nuevo item (por ahora con slice de un item)
	items := []domain.ItemResult{item}
	if err := s.eventBus.EmitNewItems(ctx, "live-search", items); err != nil {
		s.logger.Error("livesearch", "Failed to emit new items", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// extractSearchID extrae el ID de búsqueda de la URL
func (s *LiveSearchApplicationService) extractSearchID(url string) string {
	// Implementación simplificada - en el futuro usar regex para extraer el ID real
	return "extracted-search-id"
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

	// Si el live search está corriendo, sincronizar con el servicio legacy
	if s.state == domain.LiveSearchRunning {
		if wsAdapter, ok := s.webSocketClient.(interface{ GetLegacyLinkStatuses() map[int]string }); ok {
			legacyStatuses := wsAdapter.GetLegacyLinkStatuses()

			// Sincronizar estados del servicio legacy
			for id, status := range legacyStatuses {
				s.linkStatuses[id] = status
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
	}
}
