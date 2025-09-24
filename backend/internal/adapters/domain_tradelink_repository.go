package adapters

import (
	"context"
	"database/sql"
	"time"

	"github.com/SManriqueDev/poe-tool/backend/db"
	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// DomainTradeLinkRepository implementa TradeLinkRepository usando acceso directo a la base de datos
type DomainTradeLinkRepository struct {
	db *sql.DB
}

// NewDomainTradeLinkRepository crea un nuevo repository domain puro
func NewDomainTradeLinkRepository() *DomainTradeLinkRepository {
	return &DomainTradeLinkRepository{db: db.GetDB()}
}

// GetActiveTradeLinks obtiene los trade links activos
func (r *DomainTradeLinkRepository) GetActiveTradeLinks(ctx context.Context) ([]domain.TradeLink, error) {
	query := `
		SELECT id, url, description, selected, COALESCE(created_at, datetime('now'))
		FROM trade_links
		WHERE selected = 1
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []domain.TradeLink
	for rows.Next() {
		var link domain.TradeLink
		var createdAtStr string

		err := rows.Scan(&link.ID, &link.URL, &link.Description, &link.Selected, &createdAtStr)
		if err != nil {
			return nil, err
		}

		// Parsear el timestamp
		if createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
			link.CreatedAt = createdAt
		} else {
			link.CreatedAt = time.Now()
		}

		links = append(links, link)
	}

	return links, rows.Err()
}

// GetByID obtiene un trade link por ID
func (r *DomainTradeLinkRepository) GetByID(ctx context.Context, id int) (*domain.TradeLink, error) {
	query := `
		SELECT id, url, description, selected, COALESCE(created_at, datetime('now'))
		FROM trade_links
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var link domain.TradeLink
	var createdAtStr string

	err := row.Scan(&link.ID, &link.URL, &link.Description, &link.Selected, &createdAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTradeLink
		}
		return nil, err
	}

	// Parsear el timestamp
	if createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
		link.CreatedAt = createdAt
	} else {
		link.CreatedAt = time.Now()
	}

	return &link, nil
}

// Create crea un nuevo trade link
func (r *DomainTradeLinkRepository) Create(ctx context.Context, tradeLink *domain.TradeLink) error {
	query := `
		INSERT INTO trade_links (url, description, selected, created_at)
		VALUES (?, ?, ?, datetime('now'))
	`
	result, err := r.db.ExecContext(ctx, query, tradeLink.URL, tradeLink.Description, r.boolToInt(tradeLink.Selected))
	if err != nil {
		return err
	}

	// Obtener el ID generado
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	tradeLink.ID = int(id)
	tradeLink.CreatedAt = time.Now()
	return nil
}

// Update actualiza un trade link
func (r *DomainTradeLinkRepository) Update(ctx context.Context, tradeLink *domain.TradeLink) error {
	query := `
		UPDATE trade_links
		SET url = ?, description = ?, selected = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query, tradeLink.URL, tradeLink.Description, r.boolToInt(tradeLink.Selected), tradeLink.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrTradeLink
	}

	return nil
}

// Delete elimina un trade link
func (r *DomainTradeLinkRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM trade_links WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrTradeLink
	}

	return nil
}

// List obtiene todos los trade links
func (r *DomainTradeLinkRepository) List(ctx context.Context) ([]domain.TradeLink, error) {
	query := `
		SELECT id, url, description, selected, COALESCE(created_at, datetime('now'))
		FROM trade_links
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []domain.TradeLink
	for rows.Next() {
		var link domain.TradeLink
		var createdAtStr string

		err := rows.Scan(&link.ID, &link.URL, &link.Description, &link.Selected, &createdAtStr)
		if err != nil {
			return nil, err
		}

		// Parsear el timestamp
		if createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
			link.CreatedAt = createdAt
		} else {
			link.CreatedAt = time.Now()
		}

		links = append(links, link)
	}

	return links, rows.Err()
}

// GetAll es un alias para List para compatibilidad
func (r *DomainTradeLinkRepository) GetAll(ctx context.Context) ([]domain.TradeLink, error) {
	return r.List(ctx)
}

// boolToInt convierte bool a int para SQLite
func (r *DomainTradeLinkRepository) boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Verificar que implementa la interfaz
var _ domain.TradeLinkRepository = (*DomainTradeLinkRepository)(nil)
