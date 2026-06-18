package adapters

import (
	"context"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// WailsLogEmitter implementa logging.EventEmitter para emitir eventos Wails
type WailsLogEmitter struct{}

// NewWailsLogEmitter crea un nuevo emitter de logs para Wails
func NewWailsLogEmitter() *WailsLogEmitter {
	return &WailsLogEmitter{}
}

// EmitNewLog emite un evento de nuevo log al frontend
func (e *WailsLogEmitter) EmitNewLog(ctx context.Context, logEntry interface{}) {
	app := application.Get()
	if app == nil {
		return
	}

	app.Event.Emit("logs:newEntry", logEntry)
}
