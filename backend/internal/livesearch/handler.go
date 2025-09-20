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

func (h *Handler) SetGoToHideout(value bool) error {
	return h.svc.SetGoToHideout(value)
}

func (h *Handler) GetGoToHideout() bool {
	cfg := h.svc.settingsSvc.Get()
	return cfg.GoToHideout
}

func (h *Handler) DeleteTradeLink(id int) error {
	return h.svc.DeleteTradeLink(id)
}
