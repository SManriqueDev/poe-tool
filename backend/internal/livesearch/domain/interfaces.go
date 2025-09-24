package domain

import (
	"context"
	"errors"
)

// Errores del dominio
var (
	ErrNoActiveTradeLinks  = errors.New("no active trade links found")
	ErrLiveSearchRunning   = errors.New("live search is already running")
	ErrLiveSearchStopped   = errors.New("live search is not running")
	ErrTradeLink           = errors.New("trade link not found")
	ErrSettingNotFound     = errors.New("setting not found")
	ErrInvalidSettingValue = errors.New("invalid setting value")
)

// TradeLinkRepository define el contrato para el repositorio de trade links
type TradeLinkRepository interface {
	GetActiveTradeLinks(ctx context.Context) ([]TradeLink, error)
	GetByID(ctx context.Context, id int) (*TradeLink, error)
	Create(ctx context.Context, tradeLink *TradeLink) error
	Update(ctx context.Context, tradeLink *TradeLink) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]TradeLink, error)
	GetAll(ctx context.Context) ([]TradeLink, error) // Alias para List para compatibilidad
}

// LiveSearchRepository define el contrato para configuración de live search
type LiveSearchRepository interface {
	GetSetting(ctx context.Context, key string) (interface{}, error)
	SetSetting(ctx context.Context, key string, value interface{}) error
	GetHideoutSettings(ctx context.Context) (*HideoutSettings, error)
	UpdateHideoutSettings(ctx context.Context, settings *HideoutSettings) error
}

// WebSocketClient define el contrato para el cliente WebSocket
type WebSocketClient interface {
	Connect(ctx context.Context, url string) error
	Disconnect(ctx context.Context) error
	Subscribe(ctx context.Context, searchID string) error
	Unsubscribe(ctx context.Context, searchID string) error
	IsConnected() bool
	GetMessageChannel() <-chan ItemResult
}

// EventBus define el contrato para el bus de eventos
type EventBus interface {
	EmitNewItems(ctx context.Context, searchID string, items []ItemResult) error
	EmitLinkStatusChanged(ctx context.Context, linkID int, status string) error
	EmitLiveSearchStarted(ctx context.Context) error
	EmitLiveSearchStopped(ctx context.Context) error
}

// Logger define el contrato para logging
type Logger interface {
	Info(module, message string, metadata map[string]interface{}) error
	Error(module, message string, metadata map[string]interface{}) error
	Warning(module, message string, metadata map[string]interface{}) error
	Debug(module, message string, metadata map[string]interface{}) error
}

// WindowManager define el contrato para gestión de ventanas
type WindowManager interface {
	OpenLogsWindow(ctx context.Context) error
	CloseWindow(ctx context.Context, windowID string) error
	ShowWindow(ctx context.Context, windowID string) error
	HideWindow(ctx context.Context, windowID string) error
	GetActiveWindows(ctx context.Context) ([]WindowInfo, error)
	BringToFront(ctx context.Context, windowID string) error
}

// HideoutAutomation define el contrato para automatización de hideout
type HideoutAutomation interface {
	ProcessHideoutVisit(ctx context.Context, hideoutToken, itemID string) error
	QueueHideoutVisit(ctx context.Context, hideoutToken, itemID string) error
	GetQueueSize(ctx context.Context) (int, error)
	IsProcessing(ctx context.Context) (bool, error)
	StartProcessingQueue(ctx context.Context) error
	StopProcessingQueue(ctx context.Context) error
	ClearQueue(ctx context.Context) error
}

// SystemAPIClient define el contrato para APIs del sistema (PoE API)
type SystemAPIClient interface {
	SendHideoutRequest(ctx context.Context, hideoutToken, poeSessid string) error
	IsConnected(ctx context.Context) (bool, error)
	GetSystemInfo(ctx context.Context) (*SystemInfo, error)
	ValidatePoeSessid(ctx context.Context, poeSessid string) error
}
