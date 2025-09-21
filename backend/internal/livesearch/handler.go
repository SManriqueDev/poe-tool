package livesearch

import "context"

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) AddTradeLink(url string, description string) {
	h.svc.AddTradeLink(url, description)
}

func (h *Handler) ListTradeLinks() []TradeLink {
	return h.svc.ListTradeLinks()
}

//func (h *Handler) UpdateTradeLinks(links []TradeLink) {
//	h.svc.UpdateTradeLinks(links)
//}

func (h *Handler) UpdateTradeLink(id int, url string, description string, selected bool) error {
	return h.svc.UpdateTradeLink(id, url, description, selected)
}

func (h *Handler) StartLiveSearch() []TradeLink {
	return h.svc.StartLiveSearch()
}

func (h *Handler) StopLiveSearch() {
	h.svc.StopLiveSearch()
}

func (h *Handler) SetContext(ctx context.Context) {
	h.svc.SetContext(ctx)
}

func (h *Handler) DeleteTradeLink(id int) error {
	return h.svc.DeleteTradeLink(id)
}

func (h *Handler) SetGoToHideout(enabled bool) error {
	return h.svc.SetGoToHideout(enabled)
}

func (h *Handler) GetGoToHideout() (bool, error) {
	return h.svc.GetGoToHideout()
}

func (h *Handler) IsLiveSearchRunning() bool {
	return h.svc.IsLiveSearchRunning()
}

func (h *Handler) GetAllLinkStatuses() map[int]string {
	return h.svc.GetAllLinkStatuses()
}

func (h *Handler) OpenLogsWindow() error {
	return OpenLogsWindow()
}
