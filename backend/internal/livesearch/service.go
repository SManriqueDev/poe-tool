package livesearch

import (
	"github.com/SManriqueDev/poe-tool/backend/internal/settings"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
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

func (s *Service) StartLiveSearch() []TradeLink {
	cfg := s.settingsSvc.Get()
	poeSess := cfg.PoeSessid

	s.mu.Lock()
	links := append([]TradeLink{}, s.links...)
	s.mu.Unlock()

	var wg sync.WaitGroup
	statusLinks := make([]TradeLink, len(links))

	for i, link := range links {
		statusLinks[i] = link
		if !link.Selected {
			continue
		}
		wg.Add(1)
		go func(idx int, link TradeLink) {
			defer wg.Done()
			wsURL := url.URL{
				Scheme: "wss",
				Host:   "www.pathofexile.com",
				Path:   "/api/trade2/live/" + url.PathEscape(link.League) + "/" + link.SearchID,
			}
			header := http.Header{}
			header.Set("Cookie", "POESESSID="+poeSess)
			header.Set("Content-Type", "application/json")

			conn, resp, err := websocket.DefaultDialer.Dial(wsURL.String(), header)

			if err != nil {
				log.Printf("WebSocket dial error for %s: %v (HTTP %v)", wsURL.String(), err, resp)
				return
			}
			defer conn.Close()

			var authResp struct {
				Auth bool `json:"auth"`
			}
			if err := conn.ReadJSON(&authResp); err != nil || !authResp.Auth {
				statusLinks[idx].Status = "auth_error"
				return
			}
			statusLinks[idx].Status = "ok"
		}(i, link)
	}
	wg.Wait()
	return statusLinks
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
