package adapters

import (
	"context"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch"
)

// LiveSearchRepositoryAdapter adapta las operaciones de configuración del livesearch
type LiveSearchRepositoryAdapter struct {
	repo *livesearch.Repository
}

// NewLiveSearchRepositoryAdapter crea un nuevo adaptador
func NewLiveSearchRepositoryAdapter(repo *livesearch.Repository) *LiveSearchRepositoryAdapter {
	return &LiveSearchRepositoryAdapter{repo: repo}
}

// GetSetting obtiene una configuración por clave
func (a *LiveSearchRepositoryAdapter) GetSetting(ctx context.Context, key string) (interface{}, error) {
	// Mapear las configuraciones conocidas
	switch key {
	case "go_to_hideout":
		return a.repo.GetLiveSearchSetting(key)
	default:
		return nil, domain.ErrSettingNotFound
	}
}

// SetSetting establece una configuración
func (a *LiveSearchRepositoryAdapter) SetSetting(ctx context.Context, key string, value interface{}) error {
	switch key {
	case "go_to_hideout":
		if boolVal, ok := value.(bool); ok {
			return a.repo.SetLiveSearchSetting(key, boolVal)
		}
		return domain.ErrInvalidSettingValue
	default:
		return domain.ErrSettingNotFound
	}
}

// GetHideoutSettings obtiene la configuración del hideout
func (a *LiveSearchRepositoryAdapter) GetHideoutSettings(ctx context.Context) (*domain.HideoutSettings, error) {
	enabled, err := a.repo.GetLiveSearchSetting("go_to_hideout")
	if err != nil {
		return nil, err
	}
	
	return &domain.HideoutSettings{
		Enabled: enabled,
	}, nil
}

// UpdateHideoutSettings actualiza la configuración del hideout
func (a *LiveSearchRepositoryAdapter) UpdateHideoutSettings(ctx context.Context, settings *domain.HideoutSettings) error {
	return a.repo.SetLiveSearchSetting("go_to_hideout", settings.Enabled)
}