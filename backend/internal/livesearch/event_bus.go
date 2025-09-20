package livesearch

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type EventBus interface {
	EmitStatusUpdate(ctx context.Context, link TradeLink)
	EmitNewItems(ctx context.Context, searchID string, items []ItemResult)
}

type WailsEventBus struct{}

func (b *WailsEventBus) EmitStatusUpdate(ctx context.Context, link TradeLink) {
	if ctx != nil {
		runtime.EventsEmit(ctx, "linkStatusChanged", link)
	}
}

func (b *WailsEventBus) EmitNewItems(ctx context.Context, searchID string, items []ItemResult) {
	if ctx != nil {
		runtime.EventsEmit(ctx, "newItemsFound", map[string]interface{}{
			"searchID": searchID,
			"items":    items,
			"count":    len(items),
		})
	}
}
