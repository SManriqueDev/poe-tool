package livesearch

import (
	"context"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/application"
	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
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

// StartLiveSearch inicia la búsqueda en vivo usando Clean Architecture
func (h *Handler) StartLiveSearch() []TradeLink {
	ctx := context.Background()

	// Usar LiveSearchApplicationService nativo (ya no legacy)
	err := h.liveSearchAppSvc.StartLiveSearch(ctx)
	if err != nil {
		h.logError("Failed to start live search", err)
		// Fallback al servicio legacy solo en caso de error crítico
		return h.svc.StartLiveSearch()
	}

	// Obtener todos los trade links desde el application service
	domainLinks, err := h.liveSearchAppSvc.GetAllTradeLinks(ctx)
	if err != nil {
		h.logError("Failed to get trade links", err)
		// Fallback al servicio legacy
		return h.svc.StartLiveSearch()
	}

	// Convertir domain links a model links para compatibilidad con frontend
	return h.convertDomainToModelLinks(domainLinks)
}

// Helper para logging de errores
func (h *Handler) logError(message string, err error) {
	if h.svc != nil && h.svc.loggingSvc != nil {
		h.svc.loggingSvc.Error("livesearch", message, map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// Helper para conversión de domain a model
func (h *Handler) convertDomainToModelLinks(domainLinks []domain.TradeLink) []TradeLink {
	var modelLinks []TradeLink
	for _, dl := range domainLinks {
		modelLinks = append(modelLinks, TradeLink{
			ID:          dl.ID,
			URL:         dl.URL,
			Description: dl.Description,
			Selected:    dl.Selected,
		})
	}
	return modelLinks
}

// StopLiveSearch detiene la búsqueda en vivo usando Clean Architecture
func (h *Handler) StopLiveSearch() {
	ctx := context.Background()
	err := h.liveSearchAppSvc.StopLiveSearch(ctx)
	if err != nil {
		h.logError("Failed to stop live search", err)
		// Fallback al servicio legacy en caso de error
		h.svc.StopLiveSearch()
	}
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

// IsLiveSearchRunning verifica si la búsqueda está corriendo usando Clean Architecture
func (h *Handler) IsLiveSearchRunning() bool {
	// Usar LiveSearchApplicationService nativo
	return h.liveSearchAppSvc.IsLiveSearchRunning()
}

// GetAllLinkStatuses retorna los estados actuales usando Clean Architecture
func (h *Handler) GetAllLinkStatuses() map[int]string {
	// Usar LiveSearchApplicationService nativo
	return h.liveSearchAppSvc.GetAllLinkStatuses()
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
