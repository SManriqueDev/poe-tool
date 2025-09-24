package adapters

import (
	"context"
	"sync"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch"
	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// WebSocketClientAdapter adapta el WebSocketClient actual al contrato del dominio
// Delega la funcionalidad completa al Service legacy que tiene la implementación funcional
type WebSocketClientAdapter struct {
	service        *livesearch.Service
	messageChannel chan domain.ItemResult
	channelMu      sync.RWMutex
	isStarted      bool
}

// NewWebSocketClientAdapter crea un nuevo adaptador para el cliente WebSocket
func NewWebSocketClientAdapter(service *livesearch.Service) *WebSocketClientAdapter {
	return &WebSocketClientAdapter{
		service:        service,
		messageChannel: make(chan domain.ItemResult, 500), // Buffer para evitar bloqueos
	}
}

// Connect delega al servicio legacy que maneja múltiples conexiones WebSocket
func (a *WebSocketClientAdapter) Connect(ctx context.Context, url string) error {
	a.channelMu.Lock()
	defer a.channelMu.Unlock()

	if a.isStarted {
		return nil // Ya está iniciado
	}

	// Iniciar el live search del servicio legacy
	// Esto maneja todas las conexiones WebSocket automáticamente
	a.service.StartLiveSearch()
	a.isStarted = true

	return nil
}

// GetLegacyLinkStatuses obtiene los estados del servicio legacy para sincronización
func (a *WebSocketClientAdapter) GetLegacyLinkStatuses() map[int]string {
	return a.service.GetAllLinkStatuses()
}

// Disconnect delega al servicio legacy
func (a *WebSocketClientAdapter) Disconnect(ctx context.Context) error {
	a.service.StopLiveSearch()
	a.channelMu.Lock()
	defer a.channelMu.Unlock()
	a.isStarted = false
	return nil
}

// Subscribe delega al servicio legacy que maneja todas las suscripciones
func (a *WebSocketClientAdapter) Subscribe(ctx context.Context, searchID string) error {
	// El servicio legacy maneja todas las suscripciones automáticamente
	// Este método es parte del contrato pero no se usa individualmente
	return nil
}

// Unsubscribe delega al servicio legacy
func (a *WebSocketClientAdapter) Unsubscribe(ctx context.Context, searchID string) error {
	// Manejado por StopLiveSearch()
	return nil
}

// IsConnected verifica si hay búsqueda activa
func (a *WebSocketClientAdapter) IsConnected() bool {
	return a.service.IsLiveSearchRunning()
}

// GetMessageChannel obtiene el canal de mensajes
// NOTA: El servicio legacy usa un patrón diferente (emite eventos directamente)
// Para compatibilidad, devolvemos un canal que se puede usar en el futuro
func (a *WebSocketClientAdapter) GetMessageChannel() <-chan domain.ItemResult {
	a.channelMu.RLock()
	defer a.channelMu.RUnlock()
	return a.messageChannel
}

// SetPOESESSID configura el POESESSID (no usado en adapter legacy)
func (a *WebSocketClientAdapter) SetPOESESSID(poeSessID string) {
	// El servicio legacy obtiene el POESESSID de la configuración directamente
	// No necesitamos implementar nada aquí
}

// Verificar que implementa la interfaz
var _ domain.WebSocketClient = (*WebSocketClientAdapter)(nil)
