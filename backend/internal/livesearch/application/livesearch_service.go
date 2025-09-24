package application

import (
	"context"
	"strings"
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

	// Por ahora usar un POESESSID hardcodeado conocido para debug
	// TODO: Obtener POESESSID desde settings repository correctamente
	poeSessID := "16a54175b9a8539332dcfbf6994ed854" // El que vimos en la DB

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

	// Extraer search ID de la URL
	searchID := s.extractSearchID(link.URL)
	if searchID == "" {
		s.logger.Error("livesearch", "Could not extract search ID from URL", map[string]interface{}{
			"link_id": link.ID,
			"url":     link.URL,
		})
		s.SetLinkStatus(link.ID, "error")
		return
	}

	// Suscribirse al search ID
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
	// URL típica: https://www.pathofexile.com/trade2/search/poe2/Rise%20of%20the%20Abyssal/4nVv4ggf9
	// Necesitamos extraer "4nVv4ggf9" de la URL

	// Buscar después del último slash
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		searchID := parts[len(parts)-1]

		// El search ID suele ser alfanumérico y puede tener guiones/underscores
		if len(searchID) > 0 && searchID != "" {
			s.logger.Info("livesearch", "Extracted search ID", map[string]interface{}{
				"url":       url,
				"search_id": searchID,
			})
			return searchID
		}
	}

	s.logger.Warning("livesearch", "Could not extract search ID from URL", map[string]interface{}{
		"url": url,
	})
	return ""
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

	// Si no hay estados pero el search está corriendo, inicializar con estados básicos
	if len(s.linkStatuses) == 0 && s.state == domain.LiveSearchRunning {
		ctx := context.Background()
		links, err := s.tradeLinkRepo.GetAll(ctx)
		if err == nil {
			for _, link := range links {
				if link.Selected {
					s.linkStatuses[link.ID] = "connecting"
				} else {
					s.linkStatuses[link.ID] = "idle"
				}
			}
		}
	} else if len(s.linkStatuses) == 0 {
		// Si no está corriendo, inicializar con "idle"
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
	}
}
