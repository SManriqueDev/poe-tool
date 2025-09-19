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

	// Cancelar búsqueda previa si estaba corriendo
	s.mu.Lock()
	if s.liveSearchCancel != nil {
		s.liveSearchCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.liveSearchCancel = cancel
	s.mu.Unlock()

	// Filtrar links seleccionados
	var selectedLinks []TradeLink
	for _, link := range links {
		if link.Selected {
			selectedLinks = append(selectedLinks, link)
		}
	}
	if len(selectedLinks) == 0 {
		return links
	}

	// Copia inicial de los links
	statusLinks := make([]TradeLink, len(links))
	copy(statusLinks, links)

	// Canal para mensajes de sockets
	msgCh := make(chan WSMessage, 500)
	// Canal para actualizaciones de estado
	statusCh := make(chan func(), 100)

	// Workers para mensajes
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

	// Goroutine única para manejar actualizaciones de estado
	go func() {
		for update := range statusCh {
			update()
		}
	}()

	// WaitGroup de conexiones
	s.liveSearchWG = &sync.WaitGroup{}

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
				statusCh <- func() {
					if resp != nil && resp.StatusCode == http.StatusUnauthorized {
						statusLinks[idx].Status = "auth_error"
					} else {
						statusLinks[idx].Status = "error"
					}
					s.eventBus.EmitStatusUpdate(s.ctx, statusLinks[idx])
				}
				return
			}

			defer func() {
				_ = conn.Close()
			}()

			// Autenticación inicial
			var authResp struct {
				Auth bool `json:"auth"`
			}
			if err := conn.ReadJSON(&authResp); err != nil || !authResp.Auth {
				statusCh <- func() {
					statusLinks[idx].Status = "auth_error"
					s.eventBus.EmitStatusUpdate(s.ctx, statusLinks[idx])
				}
				return
			}

			// OK → actualizar estado
			statusCh <- func() {
				statusLinks[idx].Status = "ok"
				s.eventBus.EmitStatusUpdate(s.ctx, statusLinks[idx])
			}

			// Bucle de lectura
			for {
				select {
				case <-ctx.Done():
					statusCh <- func() {
						statusLinks[idx].Status = "idle"
						s.eventBus.EmitStatusUpdate(s.ctx, statusLinks[idx])
					}
					return
				default:
					_, message, err := conn.ReadMessage()
					if err != nil {
						log.Printf("WebSocket read error for %s: %v", link.URL, err)
						return
					}
					select {
					case msgCh <- WSMessage{SearchID: link.SearchID(), Message: message}:
					default:
						log.Printf("msgCh full, dropping message for %s", link.SearchID())
					}
				}
			}
		}(i, link)
	}

	// Goroutine para limpiar al terminar todas las conexiones
	go func() {
		s.liveSearchWG.Wait()
		close(msgCh) // cerrar workers
		workerWg.Wait()
		close(statusCh) // cerrar actualizador de estado
	}()

	// Retornar inmediatamente (queda corriendo en background)
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
