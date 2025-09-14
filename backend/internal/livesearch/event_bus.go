package livesearch

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type EventBus interface {
	EmitStatusUpdate(ctx context.Context, link TradeLink)
}

type WailsEventBus struct{}

func (b *WailsEventBus) EmitStatusUpdate(ctx context.Context, link TradeLink) {
	if ctx != nil {
		runtime.EventsEmit(ctx, "linkStatusChanged", link)
	}
}
