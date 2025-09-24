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
	// Esta función contendrá la lógica actual de búsqueda que está en el Service
	// Se moverá gradualmente desde service.go
}
