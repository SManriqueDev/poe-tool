package livesearch

import (
	"database/sql"

	"github.com/SManriqueDev/poe-tool/backend/db"
)

type LiveSearchSetting struct {
	ID      int
	Name    string
	Enabled bool
}

type Repository struct {
	db *sql.DB
}

func NewRepository() *Repository {
	return &Repository{db: db.GetDB()}
}

func (r *Repository) AddTradeLink(url, description string) error {
	_, err := r.db.Exec("INSERT INTO trade_links (url, description, selected) VALUES (?, ?, ?)", url, description, boolToInt(false))
	return err
}

func (r *Repository) GetTradeLinks() ([]TradeLink, error) {
	rows, err := r.db.Query("SELECT id, url, description, selected FROM trade_links")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []TradeLink
	for rows.Next() {
		var id int
		var url, description string
		var selected bool

		if err := rows.Scan(&id, &url, &description, &selected); err != nil {
			return nil, err
		}

		tl := NewTradeLink(
			WithID(id),
			WithURL(url),
			WithDescription(description),
			WithSelected(selected),
			WithStatus("idle"),
		)
		links = append(links, *tl)
	}
	return links, nil
}

func (r *Repository) AddLiveSearchSetting(name string, enabled bool) error {
	_, err := r.db.Exec("INSERT INTO live_search_settings (name, enabled) VALUES (?, ?)", name, boolToInt(enabled))
	return err
}

func (r *Repository) GetLiveSearchSettings() ([]LiveSearchSetting, error) {
	rows, err := r.db.Query("SELECT id, name, enabled FROM live_search_settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var settings []LiveSearchSetting
	for rows.Next() {
		var s LiveSearchSetting
		var enabledInt int
		if err := rows.Scan(&s.ID, &s.Name, &enabledInt); err != nil {
			return nil, err
		}
		s.Enabled = enabledInt == 1
		settings = append(settings, s)
	}
	return settings, nil
}

func (r *Repository) UpdateTradeLink(id int, url string, description string, selected bool) error {
	_, err := r.db.Exec(
		"UPDATE trade_links SET url = ?, description = ?, selected = ? WHERE id = ?",
		url, description, boolToInt(selected), id,
	)
	return err
}

func (r *Repository) DeleteTradeLink(id int) error {
	_, err := r.db.Exec("DELETE FROM trade_links WHERE id = ?",
		id,
	)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (r *Repository) SetLiveSearchSetting(name string, enabled bool) error {
	// First try to update existing record
	result, err := r.db.Exec("UPDATE live_search_settings SET enabled = ? WHERE name = ?", boolToInt(enabled), name)
	if err != nil {
		return err
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were affected, insert new record
	if rowsAffected == 0 {
		_, err = r.db.Exec("INSERT INTO live_search_settings (name, enabled) VALUES (?, ?)", name, boolToInt(enabled))
		return err
	}

	return nil
}

func (r *Repository) UpdateLiveSearchSetting(name string, enabled bool) error {
	_, err := r.db.Exec("UPDATE live_search_settings SET enabled = ? WHERE name = ?", boolToInt(enabled), name)
	return err
}

func (r *Repository) LiveSearchSettingExists(name string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM live_search_settings WHERE name = ?", name).Scan(&count)
	return count > 0, err
}

func (r *Repository) InitializeLiveSearchSetting(name string, defaultValue bool) error {
	exists, err := r.LiveSearchSettingExists(name)
	if err != nil {
		return err
	}

	if !exists {
		return r.SetLiveSearchSetting(name, defaultValue)
	}

	return nil
}

func (r *Repository) GetLiveSearchSetting(name string) (bool, error) {
	var enabledInt int
	err := r.db.QueryRow("SELECT enabled FROM live_search_settings WHERE name = ?", name).Scan(&enabledInt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Default to false if setting doesn't exist
		}
		return false, err
	}
	return enabledInt == 1, nil
}
