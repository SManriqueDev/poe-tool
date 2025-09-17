package livesearch

type TradeLink struct {
	ID          int    `json:"id"`
	League      string `json:"league"`
	SearchID    string `json:"searchId"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Selected    bool   `json:"selected"`
	Status      string `json:"status"` // e.g., "connected", "auth", "error"
}
