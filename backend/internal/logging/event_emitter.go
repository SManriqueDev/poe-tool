package logging

import "context"

// EventEmitter defines the interface for emitting events
// This allows logging service to emit events without depending on livesearch package
type EventEmitter interface {
	EmitNewLog(ctx context.Context, logEntry interface{})
}
