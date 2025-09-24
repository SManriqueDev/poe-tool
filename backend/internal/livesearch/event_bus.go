package livesearch

import (
	"context"
	"errors"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type EventBus interface {
	EmitStatusUpdate(ctx context.Context, link TradeLink)
	EmitNewItems(ctx context.Context, searchID string, items []ItemResult) error
	EmitNewLog(ctx context.Context, logEntry interface{})
	EmitLinkStatusChanged(ctx context.Context, linkID int, status string) error
	EmitLiveSearchStarted(ctx context.Context) error
	EmitLiveSearchStopped(ctx context.Context) error
}

type WailsEventBus struct {
}

func (b *WailsEventBus) EmitStatusUpdate(ctx context.Context, link TradeLink) {
	app := application.Get()
	if app != nil {
		app.Event.Emit("linkStatusChanged", link)
	}
}

func (b *WailsEventBus) EmitLinkStatusChanged(ctx context.Context, linkID int, status string) error {
	app := application.Get()
	if app == nil {
		return errors.New("could not get application instance")
	}

	app.Event.Emit("linkStatusChanged", map[string]interface{}{
		"linkID": linkID,
		"status": status,
	})

	return nil
}

func (b *WailsEventBus) EmitLiveSearchStarted(ctx context.Context) error {
	app := application.Get()
	if app == nil {
		return errors.New("could not get application instance")
	}

	app.Event.Emit("livesearch:started", map[string]interface{}{
		"timestamp": ctx.Value("timestamp"),
	})

	return nil
}

func (b *WailsEventBus) EmitLiveSearchStopped(ctx context.Context) error {
	app := application.Get()
	if app == nil {
		return errors.New("could not get application instance")
	}

	app.Event.Emit("livesearch:stopped", map[string]interface{}{
		"timestamp": ctx.Value("timestamp"),
	})

	return nil
}

func (b *WailsEventBus) EmitNewItems(ctx context.Context, searchID string, items []ItemResult) error {
	app := application.Get()
	if app == nil {
		return errors.New("could not get application instance")
	}

	app.Event.Emit("livesearch:newItemsFound", map[string]interface{}{
		"searchID": searchID,
		"items":    items,
		"count":    len(items),
	})

	return nil
}

func (b *WailsEventBus) EmitNewLog(ctx context.Context, logEntry interface{}) {
	app := application.Get()
	if app == nil {
		return
	}

	// Emitir evento global
	app.Event.Emit("livesearch:newLog", logEntry)
}
