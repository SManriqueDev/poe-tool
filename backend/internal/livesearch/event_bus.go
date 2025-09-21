package livesearch

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

type EventBus interface {
	EmitStatusUpdate(link TradeLink)
	EmitNewItems(searchID string, items []ItemResult)
}

type WailsEventBus struct {
	app *application.App
}

func NewWailsEventBus(app *application.App) *WailsEventBus {
	return &WailsEventBus{app: app}
}

func (b *WailsEventBus) EmitStatusUpdate(link TradeLink) {
	if b.app != nil {
		b.app.Event.Emit("linkStatusChanged", link)
	}
}

func (b *WailsEventBus) EmitNewItems(searchID string, items []ItemResult) {
	if b.app != nil {
		b.app.Event.Emit("newItemsFound", map[string]interface{}{
			"searchID": searchID,
			"items":    items,
			"count":    len(items),
		})
	}
}
