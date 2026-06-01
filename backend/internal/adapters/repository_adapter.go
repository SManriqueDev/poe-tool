package adapters

import (
	"context"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch"
	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// RepositoryAdapter adapta el Repository actual al contrato del dominio
type RepositoryAdapter struct {
	repo *livesearch.Repository
}

// NewRepositoryAdapter crea un nuevo adaptador para el repositorio
func NewRepositoryAdapter(repo *livesearch.Repository) *RepositoryAdapter {
	return &RepositoryAdapter{repo: repo}
}

// GetActiveTradeLinks obtiene los trade links activos
func (a *RepositoryAdapter) GetActiveTradeLinks(ctx context.Context) ([]domain.TradeLink, error) {
	tradeLinks, err := a.repo.GetTradeLinks()
	if err != nil {
		return nil, err
	}

	var active []domain.TradeLink
	for _, tl := range tradeLinks {
		if tl.Selected {
			active = append(active, tl)
		}
	}

	return active, nil
}

// GetByID obtiene un trade link por ID
func (a *RepositoryAdapter) GetByID(ctx context.Context, id int) (*domain.TradeLink, error) {
	tradeLinks, err := a.repo.GetTradeLinks()
	if err != nil {
		return nil, err
	}

	for _, tl := range tradeLinks {
		if tl.ID == id {
			return &tl, nil
		}
	}

	return nil, domain.ErrTradeLink
}

// Create crea un nuevo trade link
func (a *RepositoryAdapter) Create(ctx context.Context, tradeLink *domain.TradeLink) error {
	return a.repo.AddTradeLink(tradeLink.URL, tradeLink.Description)
}

// Update actualiza un trade link
func (a *RepositoryAdapter) Update(ctx context.Context, tradeLink *domain.TradeLink) error {
	return a.repo.UpdateTradeLink(tradeLink.ID, tradeLink.URL, tradeLink.Description, tradeLink.Selected)
}

// Delete elimina un trade link
func (a *RepositoryAdapter) Delete(ctx context.Context, id int) error {
	return a.repo.DeleteTradeLink(id)
}

// List obtiene todos los trade links
func (a *RepositoryAdapter) List(ctx context.Context) ([]domain.TradeLink, error) {
	return a.repo.GetTradeLinks()
}

// GetAll es un alias para List para compatibilidad
func (a *RepositoryAdapter) GetAll(ctx context.Context) ([]domain.TradeLink, error) {
	return a.List(ctx)
}
