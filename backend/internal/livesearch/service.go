package livesearch

import (
	"context"
	"log"
	"net/http"

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
	repo             *Repository
	wsClient         *WebSocketClient
	eventBus         EventBus
	liveSearchWG     *sync.WaitGroup
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
		repo:        NewRepository(),
		wsClient:    NewWebSocketClient(),
		eventBus:    &WailsEventBus{},
	}
}

func (s *Service) broadcastStatusUpdate(link TradeLink) {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "linkStatusChanged", link)
	}
}

func (s *Service) AddTradeLink(url string, description string) {
	link := TradeLink{
		URL:         url,
		Description: description,
		Status:      "idle",
	}
	s.links = append(s.links, link)
	_ = s.repo.AddTradeLink(link.URL, link.Description)
}

func (s *Service) ListTradeLinks() []TradeLink {
	links, err := s.repo.GetTradeLinks()
	if err != nil {
		return []TradeLink{}
	}
	var tradeLinks []TradeLink
	for _, l := range links {
		tradeLinks = append(tradeLinks, TradeLink{
			ID:          l.ID,
			URL:         l.URL,
			Description: l.Description,
			Selected:    l.Selected,
			Status:      "idle",
		})
	}
	return append([]TradeLink{}, tradeLinks...)
}

func (s *Service) StartLiveSearch() []TradeLink {
	cfg := s.settingsSvc.Get()
	poeSess := cfg.PoeSessid

	links, err := s.repo.GetTradeLinks()
	if err != nil {
		return []TradeLink{}
	}

	s.mu.Lock()
	if s.liveSearchCancel != nil {
		s.liveSearchCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.liveSearchCancel = cancel
	s.mu.Unlock()

	var selectedLinks []TradeLink
	for _, link := range links {
		if link.Selected {
			selectedLinks = append(selectedLinks, link)
		}
	}
	if len(selectedLinks) == 0 {
		return links
	}

	s.liveSearchWG = &sync.WaitGroup{}
	statusLinks := make([]TradeLink, len(links))
	copy(statusLinks, links)
	msgCh := make(chan WSMessage, 100)

	var workerWg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for msg := range msgCh {
				log.Printf("Processing message for %s: %s", msg.SearchID, string(msg.Message))
			}
		}()
	}

	for i, link := range links {
		if !link.Selected {
			continue
		}
		s.liveSearchWG.Add(1)
		go func(idx int, link TradeLink) {
			defer s.liveSearchWG.Done()
			conn, resp, err := s.wsClient.Connect(ctx, link, poeSess)
			if err != nil {
				log.Printf("WebSocket connection error for %s: %v", link.URL, err)
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
					case msgCh <- WSMessage{SearchID: link.SearchID(), Message: message}:
					default:
						log.Printf("msgCh full, dropping message for %s", link.SearchID)
					}
				}
			}
		}(i, link)
	}
	s.liveSearchWG.Wait()
	close(msgCh)
	workerWg.Wait()

	return statusLinks
}

func (s *Service) StopLiveSearch() {
	s.mu.Lock()
	if s.liveSearchCancel != nil {
		s.liveSearchCancel()
		s.liveSearchCancel = nil
	}
	links, err := s.repo.GetTradeLinks()
	if err == nil {
		for _, link := range links {
			link.Status = "idle"
			s.eventBus.EmitStatusUpdate(s.ctx, link)
		}
	}
	s.mu.Unlock()
}

func (s *Service) UpdateTradeLink(id int, url string, description string, selected bool) error {
	return s.repo.UpdateTradeLink(id, url, description, selected)
}

func (s *Service) SetGoToHideout(value bool) error {
	cfg := s.settingsSvc.Get()
	cfg.GoToHideout = value
	return s.settingsSvc.Save()
}
