package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/application"
	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// Mocks para las interfaces del dominio
type MockTradeLinkRepository struct {
	mock.Mock
}

func (m *MockTradeLinkRepository) GetActiveTradeLinks(ctx context.Context) ([]domain.TradeLink, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.TradeLink), args.Error(1)
}

func (m *MockTradeLinkRepository) GetByID(ctx context.Context, id int) (*domain.TradeLink, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TradeLink), args.Error(1)
}

func (m *MockTradeLinkRepository) Create(ctx context.Context, tradeLink *domain.TradeLink) error {
	args := m.Called(ctx, tradeLink)
	return args.Error(0)
}

func (m *MockTradeLinkRepository) Update(ctx context.Context, tradeLink *domain.TradeLink) error {
	args := m.Called(ctx, tradeLink)
	return args.Error(0)
}

func (m *MockTradeLinkRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTradeLinkRepository) List(ctx context.Context) ([]domain.TradeLink, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.TradeLink), args.Error(1)
}

func (m *MockTradeLinkRepository) GetAll(ctx context.Context) ([]domain.TradeLink, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.TradeLink), args.Error(1)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(module, message string, metadata map[string]interface{}) error {
	args := m.Called(module, message, metadata)
	return args.Error(0)
}

func (m *MockLogger) Error(module, message string, metadata map[string]interface{}) error {
	args := m.Called(module, message, metadata)
	return args.Error(0)
}

func (m *MockLogger) Warning(module, message string, metadata map[string]interface{}) error {
	args := m.Called(module, message, metadata)
	return args.Error(0)
}

func (m *MockLogger) Debug(module, message string, metadata map[string]interface{}) error {
	args := m.Called(module, message, metadata)
	return args.Error(0)
}

// Tests del TradeLinkApplicationService
func TestTradeLinkApplicationService_AddTradeLink(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockTradeLinkRepository)
	mockLogger := new(MockLogger)

	service := application.NewTradeLinkApplicationService(mockRepo, mockLogger)

	url := "https://www.pathofexile.com/trade/search/example"
	description := "Test trade link"

	// Configurar mocks
	mockRepo.On("Create", ctx, mock.MatchedBy(func(tl *domain.TradeLink) bool {
		return tl.URL == url && tl.Description == description && tl.Selected == true
	})).Return(nil)

	mockLogger.On("Info", "livesearch", "Trade link added successfully", mock.MatchedBy(func(metadata map[string]interface{}) bool {
		return metadata["url"] == url && metadata["description"] == description
	})).Return(nil)

	// Act
	err := service.AddTradeLink(ctx, url, description)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestTradeLinkApplicationService_AddTradeLink_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockTradeLinkRepository)
	mockLogger := new(MockLogger)

	service := application.NewTradeLinkApplicationService(mockRepo, mockLogger)

	url := "https://www.pathofexile.com/trade/search/example"
	description := "Test trade link"
	expectedError := assert.AnError

	// Configurar mocks
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.TradeLink")).Return(expectedError)
	mockLogger.On("Error", "livesearch", "Failed to create trade link", mock.MatchedBy(func(metadata map[string]interface{}) bool {
		return metadata["error"] == expectedError.Error()
	})).Return(nil)

	// Act
	err := service.AddTradeLink(ctx, url, description)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestTradeLinkApplicationService_ListTradeLinks(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockTradeLinkRepository)
	mockLogger := new(MockLogger)

	service := application.NewTradeLinkApplicationService(mockRepo, mockLogger)

	expectedTradeLinks := []domain.TradeLink{
		{
			ID:          1,
			URL:         "https://example1.com",
			Description: "First trade link",
			Selected:    true,
			CreatedAt:   time.Now(),
		},
		{
			ID:          2,
			URL:         "https://example2.com",
			Description: "Second trade link",
			Selected:    false,
			CreatedAt:   time.Now(),
		},
	}

	// Configurar mocks
	mockRepo.On("List", ctx).Return(expectedTradeLinks, nil)

	// Act
	result, err := service.ListTradeLinks(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedTradeLinks, result)
	mockRepo.AssertExpectations(t)
}

func TestTradeLinkApplicationService_ListTradeLinks_Empty(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockTradeLinkRepository)
	mockLogger := new(MockLogger)

	service := application.NewTradeLinkApplicationService(mockRepo, mockLogger)

	mockRepo.On("List", ctx).Return([]domain.TradeLink{}, nil)

	result, err := service.ListTradeLinks(ctx)

	assert.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestTradeLinkApplicationService_ListTradeLinks_RepositoryError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockTradeLinkRepository)
	mockLogger := new(MockLogger)

	service := application.NewTradeLinkApplicationService(mockRepo, mockLogger)

	expectedError := assert.AnError
	mockRepo.On("List", ctx).Return([]domain.TradeLink{}, expectedError)
	mockLogger.On("Error", "livesearch", "Failed to list trade links", mock.MatchedBy(func(metadata map[string]interface{}) bool {
		return metadata["error"] == expectedError.Error()
	})).Return(nil)

	result, err := service.ListTradeLinks(ctx)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestTradeLinkApplicationService_UpdateTradeLink(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockTradeLinkRepository)
	mockLogger := new(MockLogger)

	service := application.NewTradeLinkApplicationService(mockRepo, mockLogger)

	id := 1
	url := "https://updated-example.com"
	description := "Updated description"
	selected := false

	existingTradeLink := &domain.TradeLink{
		ID:          id,
		URL:         "https://original-example.com",
		Description: "Original description",
		Selected:    true,
		CreatedAt:   time.Now(),
	}

	// Configurar mocks
	mockRepo.On("GetByID", ctx, id).Return(existingTradeLink, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(tl *domain.TradeLink) bool {
		return tl.ID == id && tl.URL == url && tl.Description == description && tl.Selected == selected
	})).Return(nil)
	mockLogger.On("Info", "livesearch", "Trade link updated successfully", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Act
	err := service.UpdateTradeLink(ctx, id, url, description, selected)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestTradeLinkApplicationService_UpdateTradeLink_NotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockTradeLinkRepository)
	mockLogger := new(MockLogger)

	service := application.NewTradeLinkApplicationService(mockRepo, mockLogger)

	id := 999
	url := "https://example.com"
	description := "Description"
	selected := true

	// Configurar mocks
	mockRepo.On("GetByID", ctx, id).Return(nil, domain.ErrTradeLink)
	mockLogger.On("Error", "livesearch", "Trade link not found for update", mock.MatchedBy(func(metadata map[string]interface{}) bool {
		return metadata["id"] == id
	})).Return(nil)

	// Act
	err := service.UpdateTradeLink(ctx, id, url, description, selected)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrTradeLink, err)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestTradeLinkApplicationService_UpdateTradeLink_RepositoryError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockTradeLinkRepository)
	mockLogger := new(MockLogger)

	service := application.NewTradeLinkApplicationService(mockRepo, mockLogger)

	id := 1
	url := "https://example.com"
	description := "Description"
	selected := true

	existingTradeLink := &domain.TradeLink{
		ID:          id,
		URL:         "https://original.com",
		Description: "Original",
		Selected:    false,
		CreatedAt:   time.Now(),
	}

	expectedError := assert.AnError
	mockRepo.On("GetByID", ctx, id).Return(existingTradeLink, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*domain.TradeLink")).Return(expectedError)
	mockLogger.On("Error", "livesearch", "Failed to update trade link", mock.MatchedBy(func(metadata map[string]interface{}) bool {
		return metadata["id"] == id && metadata["error"] == expectedError.Error()
	})).Return(nil)

	err := service.UpdateTradeLink(ctx, id, url, description, selected)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestTradeLinkApplicationService_DeleteTradeLink(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockTradeLinkRepository)
	mockLogger := new(MockLogger)

	service := application.NewTradeLinkApplicationService(mockRepo, mockLogger)

	id := 1

	// Configurar mocks
	mockRepo.On("Delete", ctx, id).Return(nil)
	mockLogger.On("Info", "livesearch", "Trade link deleted successfully", mock.MatchedBy(func(metadata map[string]interface{}) bool {
		return metadata["id"] == id
	})).Return(nil)

	// Act
	err := service.DeleteTradeLink(ctx, id)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestTradeLinkApplicationService_DeleteTradeLink_RepositoryError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockTradeLinkRepository)
	mockLogger := new(MockLogger)

	service := application.NewTradeLinkApplicationService(mockRepo, mockLogger)

	id := 1
	expectedError := assert.AnError

	mockRepo.On("Delete", ctx, id).Return(expectedError)
	mockLogger.On("Error", "livesearch", "Failed to delete trade link", mock.MatchedBy(func(metadata map[string]interface{}) bool {
		return metadata["id"] == id && metadata["error"] == expectedError.Error()
	})).Return(nil)

	err := service.DeleteTradeLink(ctx, id)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// Benchmark para comparar performance
func BenchmarkTradeLinkApplicationService_AddTradeLink(b *testing.B) {
	ctx := context.Background()
	mockRepo := new(MockTradeLinkRepository)
	mockLogger := new(MockLogger)

	service := application.NewTradeLinkApplicationService(mockRepo, mockLogger)

	// Configurar mocks para múltiples llamadas
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.TradeLink")).Return(nil)
	mockLogger.On("Info", "livesearch", "Trade link added successfully", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.AddTradeLink(ctx, "https://example.com", "description")
	}
}

// Test de integración (usando repositorio real pero base de datos en memoria)
func TestTradeLinkApplicationService_Integration(t *testing.T) {
	// Este test mostraría cómo testear con dependencias reales pero controladas
	t.Skip("Integration test - requires database setup")

	// Ejemplo de cómo se vería:
	// 1. Crear base de datos en memoria
	// 2. Crear repositorio real con esa DB
	// 3. Crear logger real o mock
	// 4. Probar flujo completo
}
