package adapters

import (
	"context"
	"errors"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// WailsEventListener conecta eventos del dominio a eventos de Wails
type WailsEventListener struct{}

// NewWailsEventListener crea un listener que emite eventos de Wails
func NewWailsEventListener() *WailsEventListener {
	return &WailsEventListener{}
}

// OnNewItems emite evento de nuevos items al frontend
func (l *WailsEventListener) OnNewItems(ctx context.Context, searchID string, items []domain.ItemResult) error {
	app := application.Get()
	if app == nil {
		return errors.New("application instance not available")
	}

	// Convertir items del dominio al formato que espera el frontend
	var frontendItems []map[string]interface{}
	for _, item := range items {
		itemMap := map[string]interface{}{
			"id":       item.ID,
			"item":     item.Item,
			"listing":  item.Listing,
			"itemName": item.Item.Name,
		}

		// Fall back to typeLine if name is empty
		if item.Item.Name == "" {
			itemMap["itemName"] = item.Item.TypeLine
		}

		// Extract price info if available (price is optional)
		if item.Listing.Price != nil {
			itemMap["priceAmount"] = item.Listing.Price.Amount
			itemMap["priceCurrency"] = item.Listing.Price.Currency
			itemMap["priceType"] = item.Listing.Price.Type
		}

		frontendItems = append(frontendItems, itemMap)
	}

	app.Event.Emit("livesearch:new-items", map[string]interface{}{
		"searchId": searchID,
		"items":    frontendItems,
	})

	return nil
}

// OnLinkStatusChanged emite evento de cambio de estado de link
func (l *WailsEventListener) OnLinkStatusChanged(ctx context.Context, linkID int, status string) error {
	app := application.Get()
	if app == nil {
		return errors.New("application instance not available")
	}

	app.Event.Emit("linkStatusChanged", map[string]interface{}{
		"linkID": linkID,
		"status": status,
	})

	return nil
}

// OnLiveSearchStarted emite evento de inicio de live search
func (l *WailsEventListener) OnLiveSearchStarted(ctx context.Context) error {
	app := application.Get()
	if app == nil {
		return errors.New("application instance not available")
	}

	app.Event.Emit("livesearch:started", map[string]interface{}{
		"timestamp": ctx.Value("timestamp"),
	})

	return nil
}

// OnLiveSearchStopped emite evento de parada de live search
func (l *WailsEventListener) OnLiveSearchStopped(ctx context.Context) error {
	app := application.Get()
	if app == nil {
		return errors.New("application instance not available")
	}

	app.Event.Emit("livesearch:stopped", nil)

	return nil
}
