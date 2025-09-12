package settings

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) LoadConfig() *Config {
	return h.svc.Get()
}

func (h *Handler) SaveConfig(cfg *Config) error {
	return h.svc.Update(*cfg)
}
