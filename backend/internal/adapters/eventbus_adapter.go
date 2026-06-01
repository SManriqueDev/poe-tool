package adapters

import (
	"context"
	"encoding/json"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch"
	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// EventBusAdapter adapta el EventBus actual al contrato del dominio
type EventBusAdapter struct {
	eventBus livesearch.EventBus
}

// NewEventBusAdapter crea un nuevo adaptador para el bus de eventos
func NewEventBusAdapter(eventBus livesearch.EventBus) *EventBusAdapter {
	return &EventBusAdapter{eventBus: eventBus}
}

// EmitNewItems emite evento de nuevos items encontrados
func (a *EventBusAdapter) EmitNewItems(ctx context.Context, searchID string, items []domain.ItemResult) error {
	// Convertir items del dominio al formato actual
	var currentItems []livesearch.ItemResult
	for _, item := range items {
		// Convertir interface{} a json.RawMessage
		itemBytes, _ := json.Marshal(item.Item)
		listingBytes, _ := json.Marshal(item.Listing)

		currentItems = append(currentItems, livesearch.ItemResult{
			ID:      item.ID,
			Item:    json.RawMessage(itemBytes),
			Listing: json.RawMessage(listingBytes),
		})
	}

	a.eventBus.EmitNewItems(ctx, searchID, currentItems)
	return nil
}

// EmitLinkStatusChanged emite evento de cambio de estado de link
func (a *EventBusAdapter) EmitLinkStatusChanged(ctx context.Context, linkID int, status string) error {
	// Crear un TradeLink básico para el evento
	// En el futuro se puede mejorar para obtener el TradeLink completo
	link := domain.TradeLink{
		ID:          linkID,
		URL:         "",
		Description: "",
		Selected:    true,
	}

	a.eventBus.EmitStatusUpdate(ctx, link)
	return nil
}

// EmitLiveSearchStarted emite evento de inicio de búsqueda
func (a *EventBusAdapter) EmitLiveSearchStarted(ctx context.Context) error {
	// Emitir evento genérico de cambio de estado
	return a.EmitLinkStatusChanged(ctx, 0, "started")
}

// EmitLiveSearchStopped emite evento de parada de búsqueda
func (a *EventBusAdapter) EmitLiveSearchStopped(ctx context.Context) error {
	// Emitir evento genérico de cambio de estado
	return a.EmitLinkStatusChanged(ctx, 0, "stopped")
}

// Verificar que implementa la interfaz
var _ domain.EventBus = (*EventBusAdapter)(nil)
