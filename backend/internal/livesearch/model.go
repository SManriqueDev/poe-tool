package livesearch

import (
	"net/url"
	"strings"
)

type TradeLink struct {
	ID int `json:"id"`
	//League      string `json:"league"`
	//SearchID    string `json:"searchId"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Selected    bool   `json:"selected"`
	Status      string `json:"status"` // e.g., "connected", "auth", "error"
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

type TradeLinkDTO struct {
	ID          int    `json:"id"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Selected    bool   `json:"selected"`
	Status      string `json:"status"`
	League      string `json:"league"`
	SearchID    string `json:"searchId"`
}

func toDTO(link TradeLink) TradeLinkDTO {
	return TradeLinkDTO{
		ID:          link.ID,
		URL:         link.URL,
		Description: link.Description,
		Selected:    link.Selected,
		Status:      link.Status,
		League:      link.League(),
		SearchID:    link.SearchID(),
	}
}
