package settings

type DefaultTradeLink struct {
	URL         string `json:"url"`
	Description string `json:"description"`
	Selected    bool   `json:"selected"`
}
type Config struct {
	PoeSessid         string             `json:"poesessid"`
	AccountName       string             `json:"accountName"`
	League            string             `json:"league"`
	AutomationEnabled bool               `json:"automationEnabled"`
	Delay             int                `json:"delay"`
	DefaultTradeLinks []DefaultTradeLink `json:"defaultTradeLinks"`
	GoToHideout       bool               `json:"goToHideout"`
}
