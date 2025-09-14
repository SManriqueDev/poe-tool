// backend/internal/livesearch/repository.go
package livesearch

import "github.com/SManriqueDev/poe-tool/backend/internal/settings"

type TradeLinkRepository struct {
	settingsSvc *settings.Service
}

func NewTradeLinkRepository(settingsSvc *settings.Service) *TradeLinkRepository {
	return &TradeLinkRepository{settingsSvc: settingsSvc}
}

func (r *TradeLinkRepository) Load() []TradeLink {
	cfg := r.settingsSvc.Get()
	links := make([]TradeLink, 0)
	for _, d := range cfg.DefaultTradeLinks {
		league, searchId := ParseTradeLink(d.URL)
		links = append(links, TradeLink{
			League:      league,
			SearchID:    searchId,
			URL:         d.URL,
			Description: d.Description,
			Selected:    d.Selected,
			Status:      "idle",
		})
	}
	return links
}

func (r *TradeLinkRepository) Save(links []TradeLink) error {
	cfg := r.settingsSvc.Get()
	cfg.DefaultTradeLinks = make([]settings.DefaultTradeLink, 0)
	for _, l := range links {
		cfg.DefaultTradeLinks = append(cfg.DefaultTradeLinks, settings.DefaultTradeLink{
			URL:         l.URL,
			Description: l.Description,
			Selected:    l.Selected,
		})
	}
	return r.settingsSvc.Save()
}
