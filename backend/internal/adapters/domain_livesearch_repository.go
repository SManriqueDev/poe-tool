package adapters

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/SManriqueDev/poe-tool/backend/db"
	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// DomainLiveSearchRepository implementa LiveSearchRepository usando acceso directo a la base de datos
type DomainLiveSearchRepository struct {
	db *sql.DB
}

// NewDomainLiveSearchRepository crea un nuevo repository domain puro
func NewDomainLiveSearchRepository() *DomainLiveSearchRepository {
	return &DomainLiveSearchRepository{db: db.GetDB()}
}

// GetSetting obtiene una configuración por clave
func (r *DomainLiveSearchRepository) GetSetting(ctx context.Context, key string) (interface{}, error) {
	query := `SELECT enabled FROM live_search_settings WHERE name = ? LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, key)

	var enabledInt int
	err := row.Scan(&enabledInt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrSettingNotFound
		}
		return nil, err
	}

	// Convertir int a bool para settings tipo go_to_hideout
	return enabledInt == 1, nil
}

// SetSetting establece una configuración
func (r *DomainLiveSearchRepository) SetSetting(ctx context.Context, key string, value interface{}) error {
	// Convertir el valor a int (assuming boolean settings)
	var intValue int
	switch v := value.(type) {
	case bool:
		if v {
			intValue = 1
		} else {
			intValue = 0
		}
	case int:
		intValue = v
	default:
		return domain.ErrInvalidSettingValue
	}

	// Usar patrón UPDATE-then-INSERT para evitar duplicados
	updateQuery := `UPDATE live_search_settings SET enabled = ? WHERE name = ?`
	result, err := r.db.ExecContext(ctx, updateQuery, intValue, key)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// Si no se actualizó ninguna fila, insertar una nueva
	if rowsAffected == 0 {
		insertQuery := `INSERT INTO live_search_settings (name, enabled) VALUES (?, ?)`
		_, err = r.db.ExecContext(ctx, insertQuery, key, intValue)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetHideoutSettings obtiene la configuración del hideout
func (r *DomainLiveSearchRepository) GetHideoutSettings(ctx context.Context) (*domain.HideoutSettings, error) {
	// Inicializar setting si no existe
	err := r.initializeSettingIfNotExists(ctx, "go_to_hideout", false)
	if err != nil {
		return nil, err
	}

	enabled, err := r.GetSetting(ctx, "go_to_hideout")
	if err != nil {
		return nil, err
	}

	enabledBool, ok := enabled.(bool)
	if !ok {
		return nil, fmt.Errorf("invalid setting value type for go_to_hideout")
	}

	return &domain.HideoutSettings{
		Enabled: enabledBool,
	}, nil
}

// UpdateHideoutSettings actualiza la configuración del hideout
func (r *DomainLiveSearchRepository) UpdateHideoutSettings(ctx context.Context, settings *domain.HideoutSettings) error {
	return r.SetSetting(ctx, "go_to_hideout", settings.Enabled)
}

// initializeSettingIfNotExists inicializa un setting si no existe
func (r *DomainLiveSearchRepository) initializeSettingIfNotExists(ctx context.Context, key string, defaultValue bool) error {
	// Verificar si existe
	query := `SELECT COUNT(*) FROM live_search_settings WHERE name = ?`
	row := r.db.QueryRowContext(ctx, query, key)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	// Si no existe, crear con valor por defecto
	if count == 0 {
		insertQuery := `INSERT INTO live_search_settings (name, enabled) VALUES (?, ?)`
		intValue := 0
		if defaultValue {
			intValue = 1
		}
		_, err = r.db.ExecContext(ctx, insertQuery, key, intValue)
		if err != nil {
			return err
		}
	}

	return nil
}

// Verificar que implementa la interfaz
var _ domain.LiveSearchRepository = (*DomainLiveSearchRepository)(nil)
