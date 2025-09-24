package adapters

import (
	"context"
	"sync"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// DomainEventBus implementa domain.EventBus de forma pura
type DomainEventBus struct {
	logger domain.Logger

	// Event listeners
	newItemListeners         []NewItemListener
	linkStatusListeners      []LinkStatusListener
	liveSearchStatusListener []LiveSearchStatusListener
	listenersMu              sync.RWMutex

	// Event emission tracking
	eventsEmitted map[string]int
	eventsMu      sync.RWMutex
}

// Event listener interfaces
type NewItemListener interface {
	OnNewItems(ctx context.Context, searchID string, items []domain.ItemResult) error
}

type LinkStatusListener interface {
	OnLinkStatusChanged(ctx context.Context, linkID int, status string) error
}

type LiveSearchStatusListener interface {
	OnLiveSearchStarted(ctx context.Context) error
	OnLiveSearchStopped(ctx context.Context) error
}

// NewDomainEventBus crea una nueva instancia del bus de eventos
func NewDomainEventBus(logger domain.Logger) *DomainEventBus {
	return &DomainEventBus{
		logger:        logger,
		eventsEmitted: make(map[string]int),
	}
}

// EmitNewItems emite evento de nuevos items encontrados
func (e *DomainEventBus) EmitNewItems(ctx context.Context, searchID string, items []domain.ItemResult) error {
	e.trackEvent("new_items")

	e.logger.Info("eventbus", "Emitting new items event", map[string]interface{}{
		"search_id":   searchID,
		"items_count": len(items),
	})

	// Notificar a todos los listeners registrados
	e.listenersMu.RLock()
	listeners := make([]NewItemListener, len(e.newItemListeners))
	copy(listeners, e.newItemListeners)
	e.listenersMu.RUnlock()

	for _, listener := range listeners {
		if err := listener.OnNewItems(ctx, searchID, items); err != nil {
			e.logger.Error("eventbus", "Failed to notify new items listener", map[string]interface{}{
				"search_id": searchID,
				"error":     err.Error(),
			})
			// Continuar notificando otros listeners incluso si uno falla
		}
	}

	// Para compatibilidad con el sistema existente, también loggear cada item
	for _, item := range items {
		e.logger.Info("livesearch", "New item found", map[string]interface{}{
			"search_id": searchID,
			"item_id":   item.ID,
		})
	}

	e.logger.Debug("eventbus", "New items event emitted successfully", map[string]interface{}{
		"search_id":       searchID,
		"items_count":     len(items),
		"listeners_count": len(listeners),
	})

	return nil
}

// EmitLinkStatusChanged emite evento de cambio de estado de link
func (e *DomainEventBus) EmitLinkStatusChanged(ctx context.Context, linkID int, status string) error {
	e.trackEvent("link_status_changed")

	e.logger.Debug("eventbus", "Emitting link status changed event", map[string]interface{}{
		"link_id": linkID,
		"status":  status,
	})

	// Notificar a todos los listeners registrados
	e.listenersMu.RLock()
	listeners := make([]LinkStatusListener, len(e.linkStatusListeners))
	copy(listeners, e.linkStatusListeners)
	e.listenersMu.RUnlock()

	for _, listener := range listeners {
		if err := listener.OnLinkStatusChanged(ctx, linkID, status); err != nil {
			e.logger.Error("eventbus", "Failed to notify link status listener", map[string]interface{}{
				"link_id": linkID,
				"status":  status,
				"error":   err.Error(),
			})
			// Continuar notificando otros listeners incluso si uno falla
		}
	}

	return nil
}

// EmitLiveSearchStarted emite evento de inicio de live search
func (e *DomainEventBus) EmitLiveSearchStarted(ctx context.Context) error {
	e.trackEvent("live_search_started")

	e.logger.Info("eventbus", "Emitting live search started event", nil)

	// Notificar a todos los listeners registrados
	e.listenersMu.RLock()
	listeners := make([]LiveSearchStatusListener, len(e.liveSearchStatusListener))
	copy(listeners, e.liveSearchStatusListener)
	e.listenersMu.RUnlock()

	for _, listener := range listeners {
		if err := listener.OnLiveSearchStarted(ctx); err != nil {
			e.logger.Error("eventbus", "Failed to notify live search started listener", map[string]interface{}{
				"error": err.Error(),
			})
			// Continuar notificando otros listeners incluso si uno falla
		}
	}

	return nil
}

// EmitLiveSearchStopped emite evento de parada de live search
func (e *DomainEventBus) EmitLiveSearchStopped(ctx context.Context) error {
	e.trackEvent("live_search_stopped")

	e.logger.Info("eventbus", "Emitting live search stopped event", nil)

	// Notificar a todos los listeners registrados
	e.listenersMu.RLock()
	listeners := make([]LiveSearchStatusListener, len(e.liveSearchStatusListener))
	copy(listeners, e.liveSearchStatusListener)
	e.listenersMu.RUnlock()

	for _, listener := range listeners {
		if err := listener.OnLiveSearchStopped(ctx); err != nil {
			e.logger.Error("eventbus", "Failed to notify live search stopped listener", map[string]interface{}{
				"error": err.Error(),
			})
			// Continuar notificando otros listeners incluso si uno falla
		}
	}

	return nil
}

// RegisterNewItemListener registra un listener para eventos de nuevos items
func (e *DomainEventBus) RegisterNewItemListener(listener NewItemListener) {
	e.listenersMu.Lock()
	defer e.listenersMu.Unlock()

	e.newItemListeners = append(e.newItemListeners, listener)

	e.logger.Debug("eventbus", "New item listener registered", map[string]interface{}{
		"total_listeners": len(e.newItemListeners),
	})
}

// RegisterLinkStatusListener registra un listener para eventos de cambio de estado de link
func (e *DomainEventBus) RegisterLinkStatusListener(listener LinkStatusListener) {
	e.listenersMu.Lock()
	defer e.listenersMu.Unlock()

	e.linkStatusListeners = append(e.linkStatusListeners, listener)

	e.logger.Debug("eventbus", "Link status listener registered", map[string]interface{}{
		"total_listeners": len(e.linkStatusListeners),
	})
}

// RegisterLiveSearchStatusListener registra un listener para eventos de estado de live search
func (e *DomainEventBus) RegisterLiveSearchStatusListener(listener LiveSearchStatusListener) {
	e.listenersMu.Lock()
	defer e.listenersMu.Unlock()

	e.liveSearchStatusListener = append(e.liveSearchStatusListener, listener)

	e.logger.Debug("eventbus", "Live search status listener registered", map[string]interface{}{
		"total_listeners": len(e.liveSearchStatusListener),
	})
}

// GetEventStats retorna estadísticas de eventos emitidos
func (e *DomainEventBus) GetEventStats() map[string]int {
	e.eventsMu.RLock()
	defer e.eventsMu.RUnlock()

	// Crear copia para evitar race conditions
	stats := make(map[string]int)
	for eventType, count := range e.eventsEmitted {
		stats[eventType] = count
	}

	return stats
}

// trackEvent incrementa el contador para un tipo de evento
func (e *DomainEventBus) trackEvent(eventType string) {
	e.eventsMu.Lock()
	defer e.eventsMu.Unlock()

	e.eventsEmitted[eventType]++
}
