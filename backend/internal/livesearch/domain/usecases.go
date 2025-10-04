package domain

import (
	"context"
	"errors"
	"strings"
)

// LiveSearchUseCase define los casos de uso de negocio para búsqueda en vivo
type LiveSearchUseCase interface {
	StartLiveSearch(ctx context.Context) error
	StopLiveSearch(ctx context.Context) error
	IsLiveSearchRunning() bool
}

// TradeLinkUseCase define los casos de uso para gestión de enlaces de trade
type TradeLinkUseCase interface {
	AddTradeLink(ctx context.Context, url, description string) error
	ListTradeLinks(ctx context.Context) ([]TradeLink, error)
	UpdateTradeLink(ctx context.Context, id int, url, description string, selected bool) error
	DeleteTradeLink(ctx context.Context, id int) error
}

// HideoutUseCase define los casos de uso para gestión de hideout
type HideoutUseCase interface {
	EnableGoToHideout(ctx context.Context) error
	DisableGoToHideout(ctx context.Context) error
	IsGoToHideoutEnabled(ctx context.Context) (bool, error)
}

// ExtractSearchID extrae el ID de búsqueda de una URL de Path of Exile trade
// Retorna error si la URL no es válida o no se puede extraer el ID
func ExtractSearchID(url string) (string, error) {
	// Validar que sea una URL de PoE trade
	if !strings.Contains(url, "pathofexile.com/trade") {
		return "", errors.New("invalid Path of Exile trade URL")
	}

	// URL típica: https://www.pathofexile.com/trade2/search/poe2/Rise%20of%20the%20Abyssal/4nVv4ggf9
	// Necesitamos extraer "4nVv4ggf9" del último segmento

	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return "", errors.New("invalid URL format")
	}

	searchID := parts[len(parts)-1]

	// Validar que el searchID no esté vacío y tenga formato básico (alphanumérico con posibles guiones)
	if len(searchID) == 0 || searchID == "" {
		return "", errors.New("search ID not found in URL")
	}

	// Validación básica: debe contener solo caracteres válidos para search ID
	for _, r := range searchID {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return "", errors.New("invalid search ID format")
		}
	}

	return searchID, nil
}
