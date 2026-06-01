package livesearch

import (
	"fmt"
	"sync"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// RepositoryInterface defines the contract for data access operations
type RepositoryInterface interface {
	AddTradeLink(url, description string) error
	GetTradeLinks() ([]domain.TradeLink, error)
	UpdateTradeLink(id int, url, description string, selected bool) error
	DeleteTradeLink(id int) error
	InitializeLiveSearchSetting(name string, enabled bool) error
	UpdateLiveSearchSetting(name string, enabled bool) error
	GetLiveSearchSetting(name string) (bool, error)
}

// TradeLinkManagerInterface defines the contract for managing trade links
type TradeLinkManagerInterface interface {
	Add(url, description string) error
	List() ([]domain.TradeLink, error)
	Update(id int, url, description string, selected bool) error
	Delete(id int) error
	GetByID(id int) (*domain.TradeLink, error)
}

// TradeLinkManager handles all trade link related operations
type TradeLinkManager struct {
	repo RepositoryInterface
	mu   sync.RWMutex // Read-write mutex for concurrent access
}

// TradeLinkManagerOption defines functional options for TradeLinkManager
type TradeLinkManagerOption func(*TradeLinkManager)

// WithTradeLinkRepository sets a custom repository
func WithTradeLinkRepository(repo RepositoryInterface) TradeLinkManagerOption {
	return func(tm *TradeLinkManager) {
		tm.repo = repo
	}
}

// NewTradeLinkManager creates a new trade link manager with options
func NewTradeLinkManager(opts ...TradeLinkManagerOption) *TradeLinkManager {
	tm := &TradeLinkManager{
		repo: NewRepository(), // Default repository - Repository implements RepositoryInterface
	}

	// Apply options
	for _, opt := range opts {
		opt(tm)
	}

	return tm
}

// Add creates a new trade link
func (tm *TradeLinkManager) Add(url, description string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Validate input
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	if description == "" {
		return fmt.Errorf("description cannot be empty")
	}

	if err := tm.repo.AddTradeLink(url, description); err != nil {
		return fmt.Errorf("failed to add trade link: %w", err)
	}

	return nil
}

// List retrieves all trade links
func (tm *TradeLinkManager) List() ([]domain.TradeLink, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	links, err := tm.repo.GetTradeLinks()
	if err != nil {
		return nil, fmt.Errorf("failed to get trade links: %w", err)
	}

	return links, nil
}

// Update modifies an existing trade link
func (tm *TradeLinkManager) Update(id int, url, description string, selected bool) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Validate input
	if id <= 0 {
		return fmt.Errorf("invalid ID: %d", id)
	}
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	if description == "" {
		return fmt.Errorf("description cannot be empty")
	}

	// Update in repository
	if err := tm.repo.UpdateTradeLink(id, url, description, selected); err != nil {
		return fmt.Errorf("failed to update trade link: %w", err)
	}

	return nil
}

// Delete removes a trade link
func (tm *TradeLinkManager) Delete(id int) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Validate input
	if id <= 0 {
		return fmt.Errorf("invalid ID: %d", id)
	}

	// Delete from repository
	if err := tm.repo.DeleteTradeLink(id); err != nil {
		return fmt.Errorf("failed to delete trade link: %w", err)
	}

	return nil
}

// GetByID retrieves a specific trade link by ID
func (tm *TradeLinkManager) GetByID(id int) (*domain.TradeLink, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// Validate input
	if id <= 0 {
		return nil, fmt.Errorf("invalid ID: %d", id)
	}

	links, err := tm.repo.GetTradeLinks()
	if err != nil {
		return nil, fmt.Errorf("failed to get trade links: %w", err)
	}

	for _, l := range links {
		if l.ID == id {
			return &l, nil
		}
	}

	return nil, fmt.Errorf("trade link with ID %d not found", id)
}

// Count returns the total number of trade links
func (tm *TradeLinkManager) Count() (int, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	links, err := tm.repo.GetTradeLinks()
	if err != nil {
		return 0, fmt.Errorf("failed to count trade links: %w", err)
	}

	return len(links), nil
}

// GetSelected returns only the selected trade links
func (tm *TradeLinkManager) GetSelected() ([]domain.TradeLink, error) {
	links, err := tm.List()
	if err != nil {
		return nil, err
	}

	var selected []domain.TradeLink
	for _, link := range links {
		if link.Selected {
			selected = append(selected, link)
		}
	}

	return selected, nil
}
