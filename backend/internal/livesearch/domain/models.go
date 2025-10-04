package domain

import "time"

// TradeLink representa un enlace de búsqueda de trade
type TradeLink struct {
	ID          int       `json:"id"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Selected    bool      `json:"selected"`
	Status      string    `json:"status"` // e.g., "connected", "auth-error", "error"
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

// WindowInfo representa información de una ventana del sistema
type WindowInfo struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Visible  bool   `json:"visible"`
	Active   bool   `json:"active"`
	Position struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"position"`
	Size struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"size"`
}

// SystemInfo representa información del sistema
type SystemInfo struct {
	OS           string            `json:"os"`
	Version      string            `json:"version"`
	Architecture string            `json:"architecture"`
	Environment  map[string]string `json:"environment"`
	Connected    bool              `json:"connected"`
}

// HideoutQueueItem representa un item en la cola de hideout
type HideoutQueueItem struct {
	Token     string    `json:"token"`
	ItemID    string    `json:"item_id"`
	Timestamp time.Time `json:"timestamp"`
	Priority  bool      `json:"priority"`
	Retries   int       `json:"retries"`
}

// HideoutQueue representa la cola completa de hideout
type HideoutQueue struct {
	Items       []HideoutQueueItem `json:"items"`
	Processing  bool               `json:"processing"`
	CurrentItem *HideoutQueueItem  `json:"current_item,omitempty"`
	MaxRetries  int                `json:"max_retries"`
	Delay       time.Duration      `json:"delay"`
}
