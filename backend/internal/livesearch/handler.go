package livesearch

import (
	"context"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/application"
)

type Handler struct {
	// Servicios de aplicación (nueva arquitectura)
	tradeLinkAppSvc  *application.TradeLinkApplicationService
	hideoutAppSvc    *application.HideoutApplicationService
	liveSearchAppSvc *application.LiveSearchApplicationService

	// Servicio legacy (para compatibilidad durante migración)
	svc *Service
}

// NewHandler crea un handler con ambos: servicios nuevos y legacy
// Los servicios de aplicación se inyectan desde app.go para evitar dependencias circulares
func NewHandler(svc *Service, tradeLinkAppSvc *application.TradeLinkApplicationService, hideoutAppSvc *application.HideoutApplicationService, liveSearchAppSvc *application.LiveSearchApplicationService) *Handler {
	return &Handler{
		tradeLinkAppSvc:  tradeLinkAppSvc,
		hideoutAppSvc:    hideoutAppSvc,
		liveSearchAppSvc: liveSearchAppSvc,
		svc:              svc, // Para funcionalidades no migradas aún
	}
}

// MIGRADO: Usar servicio de aplicación (mantener compatibilidad con frontend)
func (h *Handler) AddTradeLink(url string, description string) {
	ctx := context.Background()
	if err := h.tradeLinkAppSvc.AddTradeLink(ctx, url, description); err != nil {
		// Logear el error pero mantener la firma original para compatibilidad
		// En una futura iteración se puede cambiar la firma para devolver error
	}
}

// MIGRADO: Usar servicio de aplicación (mantener compatibilidad con frontend)
func (h *Handler) ListTradeLinks() []TradeLink {
	ctx := context.Background()
	domainTradeLinks, err := h.tradeLinkAppSvc.ListTradeLinks(ctx)
	if err != nil {
		// En caso de error, devolver slice vacío para mantener compatibilidad
		return []TradeLink{}
	}

	// Convertir a modelo actual para mantener compatibilidad con frontend
	var tradeLinks []TradeLink
	for _, dtl := range domainTradeLinks {
		tradeLinks = append(tradeLinks, TradeLink{
			ID:          dtl.ID,
			URL:         dtl.URL,
			Description: dtl.Description,
			Selected:    dtl.Selected,
		})
	}

	return tradeLinks
}

//func (h *Handler) UpdateTradeLinks(links []TradeLink) {
//	h.svc.UpdateTradeLinks(links)
//}

// MIGRADO: Usar servicio de aplicación
func (h *Handler) UpdateTradeLink(id int, url string, description string, selected bool) error {
	ctx := context.Background()
	return h.tradeLinkAppSvc.UpdateTradeLink(ctx, id, url, description, selected)
}

// StartLiveSearch inicia la búsqueda en vivo usando el servicio legacy funcional
// NOTA: Mantenemos el servicio legacy para WebSocket hasta completar migración completa
func (h *Handler) StartLiveSearch() []TradeLink {
	// Usar directamente el servicio legacy que tiene toda la funcionalidad WebSocket trabajando
	return h.svc.StartLiveSearch()
}

// StopLiveSearch detiene la búsqueda en vivo usando el servicio legacy funcional
func (h *Handler) StopLiveSearch() {
	// Usar directamente el servicio legacy que tiene toda la funcionalidad WebSocket trabajando
	h.svc.StopLiveSearch()
}

func (h *Handler) SetContext(ctx context.Context) {
	h.svc.SetContext(ctx)
}

// MIGRADO: Usar servicio de aplicación
func (h *Handler) DeleteTradeLink(id int) error {
	ctx := context.Background()
	return h.tradeLinkAppSvc.DeleteTradeLink(ctx, id)
}

// MIGRADO: Usar servicio de aplicación
func (h *Handler) SetGoToHideout(enabled bool) error {
	ctx := context.Background()
	return h.hideoutAppSvc.SetGoToHideout(ctx, enabled)
}

// MIGRADO: Usar servicio de aplicación
func (h *Handler) GetGoToHideout() (bool, error) {
	ctx := context.Background()
	return h.hideoutAppSvc.IsGoToHideoutEnabled(ctx)
}

// IsLiveSearchRunning verifica si la búsqueda está corriendo usando el servicio legacy funcional
func (h *Handler) IsLiveSearchRunning() bool {
	// Usar directamente el servicio legacy que tiene el estado real del WebSocket
	return h.svc.IsLiveSearchRunning()
}

// GetAllLinkStatuses retorna los estados actuales usando el servicio legacy funcional
func (h *Handler) GetAllLinkStatuses() map[int]string {
	// Usar directamente el servicio legacy que mantiene el estado real de las conexiones
	return h.svc.GetAllLinkStatuses()
}

// MIGRADO: Usar servicio de aplicación
func (h *Handler) GetHideoutQueueSize() int {
	ctx := context.Background()
	size, err := h.hideoutAppSvc.GetQueueSize(ctx)
	if err != nil {
		// Fallback al servicio legacy
		return h.svc.GetHideoutQueueSize()
	}
	return size
}

// MIGRADO: Usar servicio de aplicación
func (h *Handler) IsHideoutProcessing() bool {
	ctx := context.Background()
	processing, err := h.hideoutAppSvc.IsProcessing(ctx)
	if err != nil {
		// Fallback al servicio legacy
		return h.svc.IsHideoutProcessing()
	}
	return processing
}

func (h *Handler) OpenLogsWindow() error {
	return OpenLogsWindow()
}

func (h *Handler) TestLogEvent() error {
	// Crear un log de prueba para testing
	return h.svc.loggingSvc.LogItemFound(
		"test-search-id",
		"test-item-id",
		"Test Item for Event Testing",
		"Test League",
		"http://test-url",
		nil,
		"Test item details for event debugging",
	)
}
