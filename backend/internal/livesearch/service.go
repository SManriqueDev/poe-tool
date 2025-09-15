package livesearch

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/SManriqueDev/poe-tool/backend/internal/settings"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const workerCount = 10 // Number of concurrent workers

type Service struct {
	links            []TradeLink
	mu               sync.Mutex
	settingsSvc      *settings.Service
	liveSearchCancel context.CancelFunc
	ctx              context.Context
	repo             *TradeLinkRepository
	wsClient         *WebSocketClient
	eventBus         EventBus
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
		repo:        NewTradeLinkRepository(settingsSvc),
		wsClient:    NewWebSocketClient(),
		eventBus:    &WailsEventBus{},
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
			Selected:    d.Selected,
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
			Selected:    l.Selected,
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
	if s.liveSearchCancel != nil {
		s.liveSearchCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.liveSearchCancel = cancel
	s.mu.Unlock()

	// Filter selected links
	var selectedLinks []TradeLink
	for _, link := range links {
		if link.Selected {
			selectedLinks = append(selectedLinks, link)
		}
	}
	if len(selectedLinks) == 0 {
		return links
	}

	var wg sync.WaitGroup
	statusLinks := make([]TradeLink, len(links))
	copy(statusLinks, links)
	msgCh := make(chan WSMessage, 100)

	// Worker pool
	var workerWg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for msg := range msgCh {
				log.Printf("Processing message for %s: %s", msg.SearchID, string(msg.Message))
				// Extend as needed
			}
		}()
	}

	// Launch goroutines only for selected links
	for i, link := range links {
		if !link.Selected {
			continue
		}
		wg.Add(1)
		go func(idx int, link TradeLink) {
			defer wg.Done()
			conn, resp, err := s.wsClient.Connect(ctx, link, poeSess)
			if err != nil {
				if resp != nil && resp.StatusCode == http.StatusUnauthorized {
					statusLinks[idx].Status = "auth_error"
					s.eventBus.EmitStatusUpdate(s.ctx, statusLinks[idx])
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
				s.eventBus.EmitStatusUpdate(s.ctx, statusLinks[idx])
				return
			}

			select {
			case <-ctx.Done():
				return
			default:
				statusLinks[idx].Status = "ok"
				s.eventBus.EmitStatusUpdate(s.ctx, statusLinks[idx])
			}

			for {
				select {
				case <-ctx.Done():
					statusLinks[idx].Status = "idle"
					s.eventBus.EmitStatusUpdate(s.ctx, statusLinks[idx])
					return
				default:
					_, message, err := conn.ReadMessage()
					if err != nil {
						log.Printf("WebSocket read error: %v", err)
						return
					}
					select {
					case msgCh <- WSMessage{SearchID: link.SearchID, Message: message}:
					default:
						log.Printf("msgCh full, dropping message for %s", link.SearchID)
					}
				}
			}
		}(i, link)
	}
	wg.Wait()
	close(msgCh)
	workerWg.Wait()

	s.mu.Lock()
	s.links = statusLinks
	_ = s.repo.Save(statusLinks)
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
	_ = s.repo.Save(links)
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
