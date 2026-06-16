package application

import (
	"context"
	"testing"
	"time"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// Mock implementations for testing

type mockTradeLinkRepository struct {
	tradeLinks []domain.TradeLink
	activeOnly bool
}

func (m *mockTradeLinkRepository) GetActiveTradeLinks(ctx context.Context) ([]domain.TradeLink, error) {
	if m.activeOnly {
		var active []domain.TradeLink
		for _, link := range m.tradeLinks {
			if link.Selected {
				active = append(active, link)
			}
		}
		return active, nil
	}
	return m.tradeLinks, nil
}

func (m *mockTradeLinkRepository) GetByID(ctx context.Context, id int) (*domain.TradeLink, error) {
	for _, link := range m.tradeLinks {
		if link.ID == id {
			return &link, nil
		}
	}
	return nil, domain.ErrTradeLink
}

func (m *mockTradeLinkRepository) Create(ctx context.Context, tradeLink *domain.TradeLink) error {
	tradeLink.ID = len(m.tradeLinks) + 1
	m.tradeLinks = append(m.tradeLinks, *tradeLink)
	return nil
}

func (m *mockTradeLinkRepository) Update(ctx context.Context, tradeLink *domain.TradeLink) error {
	for i, link := range m.tradeLinks {
		if link.ID == tradeLink.ID {
			m.tradeLinks[i] = *tradeLink
			return nil
		}
	}
	return domain.ErrTradeLink
}

func (m *mockTradeLinkRepository) Delete(ctx context.Context, id int) error {
	for i, link := range m.tradeLinks {
		if link.ID == id {
			m.tradeLinks = append(m.tradeLinks[:i], m.tradeLinks[i+1:]...)
			return nil
		}
	}
	return domain.ErrTradeLink
}

func (m *mockTradeLinkRepository) List(ctx context.Context) ([]domain.TradeLink, error) {
	return m.tradeLinks, nil
}

func (m *mockTradeLinkRepository) GetAll(ctx context.Context) ([]domain.TradeLink, error) {
	return m.List(ctx)
}

type mockLiveSearchRepository struct{}

func (m *mockLiveSearchRepository) GetSetting(ctx context.Context, key string) (interface{}, error) {
	return nil, nil
}

func (m *mockLiveSearchRepository) SetSetting(ctx context.Context, key string, value interface{}) error {
	return nil
}

func (m *mockLiveSearchRepository) GetHideoutSettings(ctx context.Context) (*domain.HideoutSettings, error) {
	return &domain.HideoutSettings{Enabled: true}, nil
}

func (m *mockLiveSearchRepository) UpdateHideoutSettings(ctx context.Context, settings *domain.HideoutSettings) error {
	return nil
}

type mockWebSocketClient struct {
	connected bool
}

func (m *mockWebSocketClient) Connect(ctx context.Context, url string) error {
	m.connected = true
	return nil
}

func (m *mockWebSocketClient) Disconnect(ctx context.Context) error {
	m.connected = false
	return nil
}

func (m *mockWebSocketClient) Subscribe(ctx context.Context, searchID, league string) error {
	return nil
}

func (m *mockWebSocketClient) Unsubscribe(ctx context.Context, searchID string) error {
	return nil
}

func (m *mockWebSocketClient) IsConnected() bool {
	return m.connected
}

func (m *mockWebSocketClient) GetMessageChannel() <-chan domain.ItemResult {
	ch := make(chan domain.ItemResult)
	close(ch) // Cerrar inmediatamente para evitar bloqueos en tests
	return ch
}

func (m *mockWebSocketClient) SetPOESESSID(poeSessID string) {
	// Mock implementation - no-op for testing
}

type mockEventBus struct {
	events []string
}

func (m *mockEventBus) EmitNewItems(ctx context.Context, searchID string, items []domain.ItemResult) error {
	m.events = append(m.events, "new_items")
	return nil
}

func (m *mockEventBus) EmitLinkStatusChanged(ctx context.Context, linkID int, status string) error {
	m.events = append(m.events, "status_changed")
	return nil
}

func (m *mockEventBus) EmitLiveSearchStarted(ctx context.Context) error {
	m.events = append(m.events, "live_search_started")
	return nil
}

func (m *mockEventBus) EmitLiveSearchStopped(ctx context.Context) error {
	m.events = append(m.events, "live_search_stopped")
	return nil
}

type mockLogger struct{}

func (m *mockLogger) Info(module, message string, metadata map[string]interface{}) error  { return nil }
func (m *mockLogger) Error(module, message string, metadata map[string]interface{}) error { return nil }
func (m *mockLogger) Warning(module, message string, metadata map[string]interface{}) error {
	return nil
}
func (m *mockLogger) Debug(module, message string, metadata map[string]interface{}) error { return nil }

type mockHideoutAutomation struct{}

func (m *mockHideoutAutomation) ProcessHideoutVisit(ctx context.Context, hideoutToken, itemID string) error {
	return nil
}
func (m *mockHideoutAutomation) QueueHideoutVisit(ctx context.Context, hideoutToken, itemID string) error {
	return nil
}
func (m *mockHideoutAutomation) GetQueueSize(ctx context.Context) (int, error)                     { return 0, nil }
func (m *mockHideoutAutomation) IsProcessing(ctx context.Context) (bool, error)                      { return false, nil }
func (m *mockHideoutAutomation) StartProcessingQueue(ctx context.Context) error                      { return nil }
func (m *mockHideoutAutomation) StopProcessingQueue(ctx context.Context) error                       { return nil }
func (m *mockHideoutAutomation) ClearQueue(ctx context.Context) error                                { return nil }

// Test functions

func TestLiveSearchApplicationService_StartLiveSearch(t *testing.T) {
	// Setup
	tradeLinkRepo := &mockTradeLinkRepository{
		tradeLinks: []domain.TradeLink{
			{ID: 1, URL: "test-url-1", Description: "Test Link 1", Selected: true},
			{ID: 2, URL: "test-url-2", Description: "Test Link 2", Selected: false},
		},
		activeOnly: true,
	}

	liveSearchRepo := &mockLiveSearchRepository{}
	wsClient := &mockWebSocketClient{}
	eventBus := &mockEventBus{}
	logger := &mockLogger{}

	service := NewLiveSearchApplicationService(
		tradeLinkRepo,
		liveSearchRepo,
		wsClient,
		eventBus,
		logger,
		&mockHideoutAutomation{},
	)

	// Test
	ctx := context.Background()
	err := service.StartLiveSearch(ctx)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !service.IsLiveSearchRunning() {
		t.Error("Expected live search to be running")
	}

	// Verificar que se emitieron eventos
	if len(eventBus.events) == 0 {
		t.Error("Expected events to be emitted")
	}

	// Limpiar
	service.StopLiveSearch(ctx)
}

func TestLiveSearchApplicationService_StopLiveSearch(t *testing.T) {
	// Setup
	tradeLinkRepo := &mockTradeLinkRepository{
		tradeLinks: []domain.TradeLink{
			{ID: 1, URL: "test-url", Description: "Test Link", Selected: true},
		},
		activeOnly: true,
	}

	service := NewLiveSearchApplicationService(
		tradeLinkRepo,
		&mockLiveSearchRepository{},
		&mockWebSocketClient{},
		&mockEventBus{},
		&mockLogger{},
		&mockHideoutAutomation{},
	)

	// Start first
	ctx := context.Background()
	service.StartLiveSearch(ctx)

	// Test stop
	err := service.StopLiveSearch(ctx)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if service.IsLiveSearchRunning() {
		t.Error("Expected live search to be stopped")
	}
}

func TestLiveSearchApplicationService_GetAllLinkStatuses(t *testing.T) {
	// Setup
	service := NewLiveSearchApplicationService(
		&mockTradeLinkRepository{},
		&mockLiveSearchRepository{},
		&mockWebSocketClient{},
		&mockEventBus{},
		&mockLogger{},
		&mockHideoutAutomation{},
	)

	// Test initial state
	statuses := service.GetAllLinkStatuses()
	if len(statuses) != 0 {
		t.Errorf("Expected empty statuses, got %v", statuses)
	}

	// Set some statuses
	service.SetLinkStatus(1, "monitoring")
	service.SetLinkStatus(2, "error")

	// Test after setting statuses
	statuses = service.GetAllLinkStatuses()
	if len(statuses) != 2 {
		t.Errorf("Expected 2 statuses, got %d", len(statuses))
	}

	if statuses[1] != "monitoring" {
		t.Errorf("Expected status 'monitoring', got %s", statuses[1])
	}

	if statuses[2] != "error" {
		t.Errorf("Expected status 'error', got %s", statuses[2])
	}
}

func TestLiveSearchApplicationService_NoActiveLinks(t *testing.T) {
	// Setup with no active links
	tradeLinkRepo := &mockTradeLinkRepository{
		tradeLinks: []domain.TradeLink{
			{ID: 1, URL: "test-url", Description: "Test Link", Selected: false},
		},
		activeOnly: true,
	}

	service := NewLiveSearchApplicationService(
		tradeLinkRepo,
		&mockLiveSearchRepository{},
		&mockWebSocketClient{},
		&mockEventBus{},
		&mockLogger{},
		&mockHideoutAutomation{},
	)

	// Test
	ctx := context.Background()
	err := service.StartLiveSearch(ctx)

	// Assert
	if err != domain.ErrNoActiveTradeLinks {
		t.Errorf("Expected ErrNoActiveTradeLinks, got %v", err)
	}

	if service.IsLiveSearchRunning() {
		t.Error("Expected live search to not be running")
	}
}

func TestLiveSearchApplicationService_ConcurrentAccess(t *testing.T) {
	// Setup
	service := NewLiveSearchApplicationService(
		&mockTradeLinkRepository{
			tradeLinks: []domain.TradeLink{
				{ID: 1, URL: "test-url", Description: "Test Link", Selected: true},
			},
			activeOnly: true,
		},
		&mockLiveSearchRepository{},
		&mockWebSocketClient{},
		&mockEventBus{},
		&mockLogger{},
		&mockHideoutAutomation{},
	)

	// Test concurrent access to status methods
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			service.SetLinkStatus(1, "status")
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			service.GetAllLinkStatuses()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// If we reach here without panicking, the test passes
}
