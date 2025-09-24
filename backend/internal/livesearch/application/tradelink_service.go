package application

import (
	"context"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// TradeLinkApplicationService implementa los casos de uso para gestión de trade links
type TradeLinkApplicationService struct {
	repo   domain.TradeLinkRepository
	logger domain.Logger
}

// NewTradeLinkApplicationService crea una nueva instancia del servicio
func NewTradeLinkApplicationService(repo domain.TradeLinkRepository, logger domain.Logger) *TradeLinkApplicationService {
	return &TradeLinkApplicationService{
		repo:   repo,
		logger: logger,
	}
}

// AddTradeLink añade un nuevo trade link
func (s *TradeLinkApplicationService) AddTradeLink(ctx context.Context, url, description string) error {
	tradeLink := &domain.TradeLink{
		URL:         url,
		Description: description,
		Selected:    true, // Por defecto seleccionado
	}

	if err := s.repo.Create(ctx, tradeLink); err != nil {
		s.logger.Error("livesearch", "Failed to create trade link", map[string]interface{}{
			"url":         url,
			"description": description,
			"error":       err.Error(),
		})
		return err
	}

	s.logger.Info("livesearch", "Trade link added successfully", map[string]interface{}{
		"url":         url,
		"description": description,
	})

	return nil
}

// ListTradeLinks obtiene todos los trade links
func (s *TradeLinkApplicationService) ListTradeLinks(ctx context.Context) ([]domain.TradeLink, error) {
	tradeLinks, err := s.repo.List(ctx)
	if err != nil {
		s.logger.Error("livesearch", "Failed to list trade links", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	return tradeLinks, nil
}

// UpdateTradeLink actualiza un trade link existente
func (s *TradeLinkApplicationService) UpdateTradeLink(ctx context.Context, id int, url, description string, selected bool) error {
	// Primero obtener el trade link existente
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("livesearch", "Trade link not found for update", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return err
	}

	// Actualizar campos
	existing.URL = url
	existing.Description = description
	existing.Selected = selected

	if err := s.repo.Update(ctx, existing); err != nil {
		s.logger.Error("livesearch", "Failed to update trade link", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return err
	}

	s.logger.Info("livesearch", "Trade link updated successfully", map[string]interface{}{
		"id":          id,
		"url":         url,
		"description": description,
		"selected":    selected,
	})

	return nil
}

// DeleteTradeLink elimina un trade link
func (s *TradeLinkApplicationService) DeleteTradeLink(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("livesearch", "Failed to delete trade link", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return err
	}

	s.logger.Info("livesearch", "Trade link deleted successfully", map[string]interface{}{
		"id": id,
	})

	return nil
}
