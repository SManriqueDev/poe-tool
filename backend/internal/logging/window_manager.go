package logging

import "context"

// WindowManager defines the interface for opening the logs window
// This allows the logging handler to open windows without depending on Wails or livesearch packages
type WindowManager interface {
	OpenLogsWindow(ctx context.Context) error
}
