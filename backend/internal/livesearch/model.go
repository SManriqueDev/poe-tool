package livesearch

import (
	"encoding/json"
	"net/url"
	"strings"
)

type TradeLink struct {
    ID          int    `json:"id"`
    URL         string `json:"url"`
    Description string `json:"description"`
    Selected    bool   `json:"selected"`
    Status      string `json:"status"` // e.g., "connected", "auth-error", "error"
}

// TradeLinkOption define una función de configuración para TradeLink
type TradeLinkOption func(*TradeLink)

// NewTradeLink crea un nuevo TradeLink con las opciones proporcionadas
func NewTradeLink(options ...TradeLinkOption) *TradeLink {
    tl := &TradeLink{
        Status: "idle", // valor por defecto
    }
    
    for _, option := range options {
        option(tl)
    }
    
    return tl
}

// Opciones para configurar TradeLink

// WithID establece el ID del TradeLink
func WithID(id int) TradeLinkOption {
    return func(tl *TradeLink) {
        tl.ID = id
    }
}

// WithURL establece la URL del TradeLink
func WithURL(url string) TradeLinkOption {
    return func(tl *TradeLink) {
        tl.URL = url
    }
}

// WithDescription establece la descripción del TradeLink
func WithDescription(description string) TradeLinkOption {
    return func(tl *TradeLink) {
        tl.Description = description
    }
}

// WithSelected establece si el TradeLink está seleccionado
func WithSelected(selected bool) TradeLinkOption {
    return func(tl *TradeLink) {
        tl.Selected = selected
    }
}

// WithStatus establece el estado del TradeLink
func WithStatus(status string) TradeLinkOption {
    return func(tl *TradeLink) {
        tl.Status = status
    }
}

// Métodos de conveniencia para crear TradeLinks comunes

// NewIdleTradeLink crea un TradeLink en estado idle
func NewIdleTradeLink(url, description string) *TradeLink {
    return NewTradeLink(
        WithURL(url),
        WithDescription(description),
        WithStatus("idle"),
        WithSelected(false),
    )
}

// NewSelectedTradeLink crea un TradeLink seleccionado
func NewSelectedTradeLink(url, description string) *TradeLink {
    return NewTradeLink(
        WithURL(url),
        WithDescription(description),
        WithSelected(true),
        WithStatus("idle"),
    )
}

// ...existing code...
func (t *TradeLink) League() string {
    u, err := url.Parse(t.URL)
    if err != nil {
        return ""
    }
    parts := strings.Split(strings.Trim(u.Path, "/"), "/")
    if len(parts) < 2 {
        return ""
    }
    return parts[len(parts)-2]
}

func (t *TradeLink) SearchID() string {
    u, err := url.Parse(t.URL)
    if err != nil {
        return ""
    }
    parts := strings.Split(strings.Trim(u.Path, "/"), "/")
    if len(parts) < 2 {
        return ""
    }
    return parts[len(parts)-1]
}

func (t *TradeLink) MarshalJSON() ([]byte, error) {
    type Alias TradeLink
    return json.Marshal(&struct {
        Alias
        League   string `json:"league"`
        SearchID string `json:"searchId"`
    }{
        Alias:    (Alias)(*t),
        League:   t.League(),
        SearchID: t.SearchID(),
    })
}