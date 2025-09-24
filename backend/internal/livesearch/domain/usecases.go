package domain

import (
	"context"
)

// LiveSearchUseCase define los casos de uso de negocio para búsqueda en vivo
type LiveSearchUseCase interface {
	StartLiveSearch(ctx context.Context) error
	StopLiveSearch(ctx context.Context) error
	IsLiveSearchRunning() bool
}

// TradeLinkUseCase define los casos de uso para gestión de enlaces de trade
type TradeLinkUseCase interface {
	AddTradeLink(ctx context.Context, url, description string) error
	ListTradeLinks(ctx context.Context) ([]TradeLink, error)
	UpdateTradeLink(ctx context.Context, id int, url, description string, selected bool) error
	DeleteTradeLink(ctx context.Context, id int) error
}

// HideoutUseCase define los casos de uso para gestión de hideout
type HideoutUseCase interface {
	EnableGoToHideout(ctx context.Context) error
	DisableGoToHideout(ctx context.Context) error
	IsGoToHideoutEnabled(ctx context.Context) (bool, error)
}
