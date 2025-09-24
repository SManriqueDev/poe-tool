package adapters

import (
	"context"
	"time"

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

// convertLegacyToDomain convierte un TradeLink legacy al formato domain
func (a *RepositoryAdapter) convertLegacyToDomain(tl *livesearch.TradeLink) domain.TradeLink {
	// Por ahora usamos time.Now() ya que el modelo legacy no tiene CreatedAt
	// En el futuro podríamos obtener este valor de la base de datos directamente
	return domain.TradeLink{
		ID:          tl.ID,
		URL:         tl.URL,
		Description: tl.Description,
		Selected:    tl.Selected,
		CreatedAt:   time.Now(), // Temporal hasta que migremos el modelo legacy
	}
}

// GetActiveTradeLinks obtiene los trade links activos
func (a *RepositoryAdapter) GetActiveTradeLinks(ctx context.Context) ([]domain.TradeLink, error) {
	// Usar el método existente
	oldTradeLinks, err := a.repo.GetTradeLinks()
	if err != nil {
		return nil, err
	}

	// Convertir al formato del dominio
	var tradeLinks []domain.TradeLink
	for _, tl := range oldTradeLinks {
		if tl.Selected { // Solo los seleccionados
			tradeLinks = append(tradeLinks, a.convertLegacyToDomain(&tl))
		}
	}

	return tradeLinks, nil
}

// GetByID obtiene un trade link por ID
func (a *RepositoryAdapter) GetByID(ctx context.Context, id int) (*domain.TradeLink, error) {
	tradeLinks, err := a.repo.GetTradeLinks()
	if err != nil {
		return nil, err
	}

	for _, tl := range tradeLinks {
		if tl.ID == id {
			converted := a.convertLegacyToDomain(&tl)
			return &converted, nil
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
	oldTradeLinks, err := a.repo.GetTradeLinks()
	if err != nil {
		return nil, err
	}

	// Convertir al formato del dominio
	var tradeLinks []domain.TradeLink
	for _, tl := range oldTradeLinks {
		tradeLinks = append(tradeLinks, a.convertLegacyToDomain(&tl))
	}

	return tradeLinks, nil
}

// GetAll es un alias para List para compatibilidad
func (a *RepositoryAdapter) GetAll(ctx context.Context) ([]domain.TradeLink, error) {
	return a.List(ctx)
}
