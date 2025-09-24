package domain

import "time"

// TradeLink representa un enlace de búsqueda de trade
type TradeLink struct {
	ID          int       `json:"id"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Selected    bool      `json:"selected"`
	CreatedAt   time.Time `json:"created_at"`
}

// LiveSearchState representa el estado de la búsqueda en vivo
type LiveSearchState int

const (
	LiveSearchStopped LiveSearchState = iota
	LiveSearchRunning
	LiveSearchPaused
)

// HideoutSettings representa la configuración del hideout
type HideoutSettings struct {
	Enabled bool `json:"enabled"`
}

// ItemResult representa un resultado de item encontrado
type ItemResult struct {
	ID       string      `json:"id"`
	Item     interface{} `json:"item"`
	Listing  interface{} `json:"listing"`
	SearchID string      `json:"search_id"`
}
