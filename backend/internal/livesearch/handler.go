package livesearch

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

func (h *Handler) UpdateTradeLinks(links []TradeLink) {
	h.svc.UpdateTradeLinks(links)
}
