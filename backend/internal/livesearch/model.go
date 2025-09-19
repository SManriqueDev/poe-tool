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
