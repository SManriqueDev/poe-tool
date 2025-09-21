package livesearch

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type EventBus interface {
	EmitStatusUpdate(ctx context.Context, link TradeLink)
	EmitNewItems(ctx context.Context, searchID string, items []ItemResult)
}

type WailsEventBus struct {
}

func (b *WailsEventBus) EmitStatusUpdate(ctx context.Context, link TradeLink) {
	app := application.Get()
	if app != nil {
		app.Event.Emit("linkStatusChanged", link)
	} else {
		fmt.Printf("❌ Could not get application instance for EmitStatusUpdate\n")
	}
}

func (b *WailsEventBus) EmitNewItems(ctx context.Context, searchID string, items []ItemResult) {
	app := application.Get()
	if app == nil {
		fmt.Printf("❌ Could not get application instance for EmitNewItems\n")
		return
	}

	app.Event.Emit("livesearch:newItemsFound", map[string]interface{}{
		"searchID": searchID,
		"items":    items,
		"count":    len(items),
	})
}
