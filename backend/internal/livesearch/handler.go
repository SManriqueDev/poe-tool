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

	logger        domain.Logger
	windowManager domain.WindowManager
}

// NewHandler crea un handler con ambos: servicios nuevos y legacy
// Los servicios de aplicación se inyectan desde app.go para evitar dependencias circulares
func NewHandler(tradeLinkAppSvc *application.TradeLinkApplicationService, hideoutAppSvc *application.HideoutApplicationService, liveSearchAppSvc *application.LiveSearchApplicationService, logger domain.Logger,
	windowManager domain.WindowManager) *Handler {
	return &Handler{
		tradeLinkAppSvc:  tradeLinkAppSvc,
		hideoutAppSvc:    hideoutAppSvc,
		liveSearchAppSvc: liveSearchAppSvc,
		logger:           logger,
		windowManager:    windowManager,
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

// MIGRADO: Usar servicio de aplicación (sin conversión, devuelve domain.TradeLink directamente)
func (h *Handler) ListTradeLinks() []domain.TradeLink {
	ctx := context.Background()
	domainTradeLinks, err := h.tradeLinkAppSvc.ListTradeLinks(ctx)
	if err != nil {
		// Log error y devolver slice vacío
		h.logger.Error("livesearch", "Failed to list trade links", map[string]interface{}{
			"error": err.Error(),
		})
		return []domain.TradeLink{}
	}
	return domainTradeLinks
}

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
		h.logger.Error("livesearch", "Failed to start live search", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Obtener todos los trade links desde el application service
	domainLinks, err := h.liveSearchAppSvc.GetAllTradeLinks(ctx)
	if err != nil {
		h.logger.Error("livesearch", "Failed to get trade links", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Convertir domain links a model links para compatibilidad con frontend
	return h.convertDomainToModelLinks(domainLinks)
}

// Helper para logging de errores
// func (h *Handler) logError(message string, err error) {
// 	if h.svc != nil && h.svc.loggingSvc != nil {
// 		h.svc.loggingSvc.Error("livesearch", message, map[string]interface{}{
// 			"error": err.Error(),
// 		})
// 	}
// }

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
		h.logger.Error("livesearch", "Failed to stop live search", map[string]interface{}{
			"error": err.Error(),
		})
	}
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
		h.logger.Error("livesearch", "Failed to get hideout queue size", map[string]interface{}{
			"error": err.Error(),
		})
		return 0
	}
	return size
}

// MIGRADO: Usar servicio de aplicación
func (h *Handler) IsHideoutProcessing() bool {
	ctx := context.Background()
	processing, err := h.hideoutAppSvc.IsProcessing(ctx)
	if err != nil {
		h.logger.Error("livesearch", "Failed to check if hideout is processing", map[string]interface{}{
			"error": err.Error(),
		})
		return false
	}
	return processing
}

func (h *Handler) OpenLogsWindow() error {
	ctx := context.Background()
	return h.windowManager.OpenLogsWindow(ctx)
}
