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

	// Por ahora, implementación básica que simula el proceso
	// En futuras iteraciones se puede mover la lógica compleja desde service.go
	for _, link := range tradeLinks {
		select {
		case <-ctx.Done():
			s.logger.Info("livesearch", "Live search cancelled", nil)
			return
		default:
			// Procesar cada trade link
			s.logger.Info("livesearch", "Processing trade link", map[string]interface{}{
				"link_id":     link.ID,
				"description": link.Description,
			})

			// Emitir evento de cambio de estado
			if err := s.eventBus.EmitLinkStatusChanged(ctx, link.ID, "processing"); err != nil {
				s.logger.Error("livesearch", "Failed to emit link status changed", map[string]interface{}{
					"link_id": link.ID,
					"error":   err.Error(),
				})
			}
		}
	}

	s.logger.Info("livesearch", "Live search process completed", nil)
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
