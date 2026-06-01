package livesearch

import (
	"testing"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// MockTradeRepository implements RepositoryInterface for testing TradeLinkManager
type MockTradeRepository struct {
	links      []domain.TradeLink
	nextID     int
	shouldFail bool
	failOn     string // "add", "get", "update", "delete"
}

func NewMockTradeRepository() *MockTradeRepository {
	return &MockTradeRepository{
		links:  make([]domain.TradeLink, 0),
		nextID: 1,
	}
}

func (m *MockTradeRepository) SetShouldFail(operation string) {
	m.shouldFail = true
	m.failOn = operation
}

func (m *MockTradeRepository) AddTradeLink(url, description string) error {
	if m.shouldFail && m.failOn == "add" {
		return MockError("mock add error")
	}

	link := domain.TradeLink{
		ID:          m.nextID,
		URL:         url,
		Description: description,
		Selected:    false,
	}
	m.links = append(m.links, link)
	m.nextID++
	return nil
}

func (m *MockTradeRepository) GetTradeLinks() ([]domain.TradeLink, error) {
	if m.shouldFail && m.failOn == "get" {
		return nil, MockError("mock get error")
	}
	return append([]domain.TradeLink{}, m.links...), nil
}

func (m *MockTradeRepository) UpdateTradeLink(id int, url, description string, selected bool) error {
	if m.shouldFail && m.failOn == "update" {
		return MockError("mock update error")
	}

	for i, link := range m.links {
		if link.ID == id {
			m.links[i].URL = url
			m.links[i].Description = description
			m.links[i].Selected = selected
			return nil
		}
	}
	return MockError("trade link not found")
}

func (m *MockTradeRepository) DeleteTradeLink(id int) error {
	if m.shouldFail && m.failOn == "delete" {
		return MockError("mock delete error")
	}

	for i, link := range m.links {
		if link.ID == id {
			m.links = append(m.links[:i], m.links[i+1:]...)
			return nil
		}
	}
	return MockError("trade link not found")
}

// Additional mock methods to satisfy RepositoryInterface
func (m *MockTradeRepository) InitializeLiveSearchSetting(name string, enabled bool) error {
	return nil
}

func (m *MockTradeRepository) UpdateLiveSearchSetting(name string, enabled bool) error {
	return nil
}

func (m *MockTradeRepository) GetLiveSearchSetting(name string) (bool, error) {
	return false, nil
}

// MockError represents a mock error for testing
type MockError string

func (e MockError) Error() string {
	return string(e)
}

// Test TradeLinkManager functionality
func TestTradeLinkManager_Add(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		description string
		shouldFail  bool
		expectError bool
	}{
		{
			name:        "valid trade link",
			url:         "https://example.com/trade",
			description: "Test trade link",
			shouldFail:  false,
			expectError: false,
		},
		{
			name:        "empty URL",
			url:         "",
			description: "Test description",
			shouldFail:  false,
			expectError: true,
		},
		{
			name:        "empty description",
			url:         "https://example.com/trade",
			description: "",
			shouldFail:  false,
			expectError: true,
		},
		{
			name:        "repository failure",
			url:         "https://example.com/trade",
			description: "Test description",
			shouldFail:  true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := NewMockTradeRepository()
			if tt.shouldFail {
				mockRepo.SetShouldFail("add")
			}

			tm := NewTradeLinkManager(WithTradeLinkRepository(mockRepo))

			// Execute
			err := tm.Add(tt.url, tt.description)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Verify the link was added
				links, _ := tm.List()
				if len(links) != 1 {
					t.Errorf("expected 1 link, got %d", len(links))
				}
				if len(links) > 0 && links[0].URL != tt.url {
					t.Errorf("expected URL %s, got %s", tt.url, links[0].URL)
				}
			}
		})
	}
}

func TestTradeLinkManager_List(t *testing.T) {
	tests := []struct {
		name        string
		shouldFail  bool
		expectError bool
		setupLinks  int // Number of links to pre-create
	}{
		{
			name:        "successful list",
			shouldFail:  false,
			expectError: false,
			setupLinks:  3,
		},
		{
			name:        "empty list",
			shouldFail:  false,
			expectError: false,
			setupLinks:  0,
		},
		{
			name:        "repository failure",
			shouldFail:  true,
			expectError: true,
			setupLinks:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := NewMockTradeRepository()
			if tt.shouldFail {
				mockRepo.SetShouldFail("get")
			}

			tm := NewTradeLinkManager(WithTradeLinkRepository(mockRepo))

			// Pre-create some links
			for i := 0; i < tt.setupLinks; i++ {
				_ = tm.Add("https://example.com/trade", "Test description")
			}

			// Execute
			links, err := tm.List()

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(links) != tt.setupLinks {
					t.Errorf("expected %d links, got %d", tt.setupLinks, len(links))
				}
			}
		})
	}
}

func TestTradeLinkManager_Update(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		url         string
		description string
		selected    bool
		shouldFail  bool
		expectError bool
	}{
		{
			name:        "valid update",
			id:          1,
			url:         "https://updated.com/trade",
			description: "Updated description",
			selected:    true,
			shouldFail:  false,
			expectError: false,
		},
		{
			name:        "invalid ID",
			id:          0,
			url:         "https://example.com/trade",
			description: "Test description",
			selected:    false,
			shouldFail:  false,
			expectError: true,
		},
		{
			name:        "empty URL",
			id:          1,
			url:         "",
			description: "Test description",
			selected:    false,
			shouldFail:  false,
			expectError: true,
		},
		{
			name:        "repository failure",
			id:          1,
			url:         "https://example.com/trade",
			description: "Test description",
			selected:    false,
			shouldFail:  true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := NewMockTradeRepository()
			if tt.shouldFail {
				mockRepo.SetShouldFail("update")
			}

			tm := NewTradeLinkManager(WithTradeLinkRepository(mockRepo))

			// Pre-create a link
			_ = tm.Add("https://example.com/trade", "Original description")

			// Execute
			err := tm.Update(tt.id, tt.url, tt.description, tt.selected)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestTradeLinkManager_Delete(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		shouldFail  bool
		expectError bool
	}{
		{
			name:        "valid delete",
			id:          1,
			shouldFail:  false,
			expectError: false,
		},
		{
			name:        "invalid ID",
			id:          0,
			shouldFail:  false,
			expectError: true,
		},
		{
			name:        "repository failure",
			id:          1,
			shouldFail:  true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := NewMockTradeRepository()
			if tt.shouldFail {
				mockRepo.SetShouldFail("delete")
			}

			tm := NewTradeLinkManager(WithTradeLinkRepository(mockRepo))

			// Pre-create a link
			_ = tm.Add("https://example.com/trade", "Test description")

			// Execute
			err := tm.Delete(tt.id)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Verify the link was deleted
				links, _ := tm.List()
				if len(links) != 0 {
					t.Errorf("expected 0 links after deletion, got %d", len(links))
				}
			}
		})
	}
}

func TestTradeLinkManager_GetByID(t *testing.T) {
	// Setup
	mockRepo := NewMockTradeRepository()
	tm := NewTradeLinkManager(WithTradeLinkRepository(mockRepo))

	// Pre-create a link
	_ = tm.Add("https://example.com/trade", "Test description")

	tests := []struct {
		name        string
		id          int
		expectError bool
	}{
		{
			name:        "existing ID",
			id:          1,
			expectError: false,
		},
		{
			name:        "non-existing ID",
			id:          999,
			expectError: true,
		},
		{
			name:        "invalid ID",
			id:          0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			link, err := tm.GetByID(tt.id)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if link != nil {
					t.Errorf("expected nil link but got %+v", link)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if link == nil {
					t.Errorf("expected link but got nil")
				}
			}
		})
	}
}

func TestTradeLinkManager_GetSelected(t *testing.T) {
	// Setup
	mockRepo := NewMockTradeRepository()
	tm := NewTradeLinkManager(WithTradeLinkRepository(mockRepo))

	// Pre-create links with different selection states
	_ = tm.Add("https://example1.com/trade", "Description 1")
	_ = tm.Add("https://example2.com/trade", "Description 2")
	_ = tm.Add("https://example3.com/trade", "Description 3")

	// Select some links
	_ = tm.Update(1, "https://example1.com/trade", "Description 1", true)
	_ = tm.Update(3, "https://example3.com/trade", "Description 3", true)

	// Execute
	selected, err := tm.GetSelected()

	// Assert
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(selected) != 2 {
		t.Errorf("expected 2 selected links, got %d", len(selected))
	}

	// Verify that the selected links are correct
	for _, link := range selected {
		if !link.Selected {
			t.Errorf("expected selected link but got unselected: %+v", link)
		}
	}
}
