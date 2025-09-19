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

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
