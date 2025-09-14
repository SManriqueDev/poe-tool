package livesearch

import (
	"context"
	"github.com/SManriqueDev/poe-tool/backend/internal/settings"
	"github.com/gorilla/websocket"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Service struct {
	links            []TradeLink
	mu               sync.Mutex
	settingsSvc      *settings.Service
	liveSearchCancel context.CancelFunc
	ctx              context.Context
}

type WSMessage struct {
	SearchID string
	Message  []byte
}

func (s *Service) SetContext(ctx context.Context) {
	s.ctx = ctx
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

func (s *Service) broadcastStatusUpdate(link TradeLink) {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "linkStatusChanged", link)
	}
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
	// Cancel any previous live search
	if s.liveSearchCancel != nil {
		s.liveSearchCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.liveSearchCancel = cancel
	s.mu.Unlock()

	var wg sync.WaitGroup
	statusLinks := make([]TradeLink, len(links))
	msgCh := make(chan WSMessage, 100)

	// Processor goroutine: handle all incoming messages
	go func() {
		for msg := range msgCh {
			log.Printf("Processing message for %s: %s", msg.SearchID, string(msg.Message))
			// Extend: update state, notify frontend, etc.
		}
	}()

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
				Path:   "/api/trade2/live/poe2/" + link.League + "/" + link.SearchID,
			}
			header := http.Header{}
			header.Set("Cookie", "POESESSID="+poeSess)
			header.Set("Origin", "https://www.pathofexile.com")
			header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/)")
			header.Set("Content-Type", "application/json")

			conn, resp, err := websocket.DefaultDialer.Dial(wsURL.String(), header)
			if err != nil {
				if resp != nil && resp.StatusCode == http.StatusUnauthorized {
					statusLinks[idx].Status = "auth_error"
					s.broadcastStatusUpdate(statusLinks[idx])
				} else {
					statusLinks[idx].Status = "error"
				}
				return
			}
			defer conn.Close()

			var authResp struct {
				Auth bool `json:"auth"`
			}
			if err := conn.ReadJSON(&authResp); err != nil || !authResp.Auth {
				statusLinks[idx].Status = "auth_error"
				s.broadcastStatusUpdate(statusLinks[idx])
				return
			}
			statusLinks[idx].Status = "ok"
			s.broadcastStatusUpdate(statusLinks[idx])

			for {
				select {
				case <-ctx.Done():
					return
				default:
					_, message, err := conn.ReadMessage()
					if err != nil {
						log.Printf("WebSocket read error: %v", err)
						return
					}
					msgCh <- WSMessage{SearchID: link.SearchID, Message: message}
				}
			}
		}(i, link)
	}
	wg.Wait()
	close(msgCh)

	s.mu.Lock()
	s.links = statusLinks
	_ = s.SaveLinksToConfig()
	s.mu.Unlock()

	return statusLinks
}

func (s *Service) StopLiveSearch() {
	s.mu.Lock()
	if s.liveSearchCancel != nil {
		s.liveSearchCancel()
		s.liveSearchCancel = nil
	}
	s.mu.Unlock()
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
