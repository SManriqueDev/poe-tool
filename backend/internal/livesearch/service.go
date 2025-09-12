package livesearch

import (
	"net/url"
	"strings"
	"sync"

	"github.com/SManriqueDev/poe-tool/backend/internal/settings"
)

type Service struct {
	links       []TradeLink
	mu          sync.Mutex
	settingsSvc *settings.Service
}

func NewService(settingsSvc *settings.Service) *Service {
	return &Service{
		links:       make([]TradeLink, 0),
		settingsSvc: settingsSvc,
	}
}

func (s *Service) LoadLinksFromConfig() {
	cfg := s.settingsSvc.Get()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links = make([]TradeLink, 0)
	for _, d := range cfg.DefaultTradeLinks {
		league, searchId := ParseTradeLink(d.URL)
		s.links = append(s.links, TradeLink{
			League:      league,
			SearchID:    searchId,
			URL:         d.URL,
			Description: d.Description,
			Selected:    false,
			Status:      "idle",
		})
	}
}

func (s *Service) SaveLinksToConfig() error {
	cfg := s.settingsSvc.Get()
	cfg.DefaultTradeLinks = make([]settings.DefaultTradeLink, 0)
	for _, l := range s.links {
		cfg.DefaultTradeLinks = append(cfg.DefaultTradeLinks, settings.DefaultTradeLink{
			URL:         l.URL,
			Description: l.Description,
		})
	}
	return s.settingsSvc.Save()
}

func (s *Service) AddTradeLink(url string, description string) {
	league, searchId := ParseTradeLink(url)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links = append(s.links, TradeLink{
		League:      league,
		SearchID:    searchId,
		URL:         url,
		Description: description,
		Selected:    false,
		Status:      "idle",
	})

	_ = s.SaveLinksToConfig()
}

func (s *Service) ListTradeLinks() []TradeLink {
	s.LoadLinksFromConfig()
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]TradeLink{}, s.links...)
}

func (s *Service) UpdateTradeLinks(links []TradeLink) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links = links
	_ = s.SaveLinksToConfig()
}

func ParseTradeLink(tradeURL string) (string, string) {
	u, err := url.Parse(tradeURL)
	if err != nil {
		return "", ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", ""
	}
	return parts[len(parts)-2], parts[len(parts)-1]
}
