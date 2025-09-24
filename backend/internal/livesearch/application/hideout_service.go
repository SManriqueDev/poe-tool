package application

import (
	"context"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// HideoutApplicationService implementa los casos de uso para gestión de hideout
type HideoutApplicationService struct {
	repo   domain.LiveSearchRepository
	logger domain.Logger
}

// NewHideoutApplicationService crea una nueva instancia del servicio
func NewHideoutApplicationService(repo domain.LiveSearchRepository, logger domain.Logger) *HideoutApplicationService {
	return &HideoutApplicationService{
		repo:   repo,
		logger: logger,
	}
}

// EnableGoToHideout habilita la funcionalidad de ir al hideout automáticamente
func (s *HideoutApplicationService) EnableGoToHideout(ctx context.Context) error {
	settings := &domain.HideoutSettings{
		Enabled: true,
	}

	if err := s.repo.UpdateHideoutSettings(ctx, settings); err != nil {
		s.logger.Error("livesearch", "Failed to enable go to hideout", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	s.logger.Info("livesearch", "Go to hideout enabled successfully", nil)
	return nil
}

// DisableGoToHideout deshabilita la funcionalidad de ir al hideout automáticamente
func (s *HideoutApplicationService) DisableGoToHideout(ctx context.Context) error {
	settings := &domain.HideoutSettings{
		Enabled: false,
	}

	if err := s.repo.UpdateHideoutSettings(ctx, settings); err != nil {
		s.logger.Error("livesearch", "Failed to disable go to hideout", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	s.logger.Info("livesearch", "Go to hideout disabled successfully", nil)
	return nil
}

// IsGoToHideoutEnabled verifica si la funcionalidad está habilitada
func (s *HideoutApplicationService) IsGoToHideoutEnabled(ctx context.Context) (bool, error) {
	settings, err := s.repo.GetHideoutSettings(ctx)
	if err != nil {
		s.logger.Error("livesearch", "Failed to get hideout settings", map[string]interface{}{
			"error": err.Error(),
		})
		return false, err
	}

	return settings.Enabled, nil
}

// SetGoToHideout establece el estado de la funcionalidad
func (s *HideoutApplicationService) SetGoToHideout(ctx context.Context, enabled bool) error {
	if enabled {
		return s.EnableGoToHideout(ctx)
	}
	return s.DisableGoToHideout(ctx)
}
