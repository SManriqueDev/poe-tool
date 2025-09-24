package adapters

import (
	"context"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch"
	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// WebSocketClientAdapter adapta el WebSocketClient actual al contrato del dominio
// Este es un adaptador simplificado que delega al Service actual
type WebSocketClientAdapter struct {
	service *livesearch.Service
}

// NewWebSocketClientAdapter crea un nuevo adaptador para el cliente WebSocket
func NewWebSocketClientAdapter(service *livesearch.Service) *WebSocketClientAdapter {
	return &WebSocketClientAdapter{service: service}
}

// Connect se conecta usando la lógica actual del service
func (a *WebSocketClientAdapter) Connect(ctx context.Context, url string) error {
	// La conexión se maneja internamente por el Service actual
	return nil
}

// Disconnect se desconecta usando la lógica actual del service
func (a *WebSocketClientAdapter) Disconnect(ctx context.Context) error {
	// La desconexión se maneja internamente por el Service actual
	return nil
}

// Subscribe se suscribe usando la lógica actual del service
func (a *WebSocketClientAdapter) Subscribe(ctx context.Context, searchID string) error {
	// La suscripción se maneja internamente por el Service actual
	return nil
}

// Unsubscribe se desuscribe usando la lógica actual del service
func (a *WebSocketClientAdapter) Unsubscribe(ctx context.Context, searchID string) error {
	// La desuscripción se maneja internamente por el Service actual
	return nil
}

// IsConnected verifica si hay búsqueda activa
func (a *WebSocketClientAdapter) IsConnected() bool {
	return a.service.IsLiveSearchRunning()
}

// GetMessageChannel obtiene el canal de mensajes (simplificado)
func (a *WebSocketClientAdapter) GetMessageChannel() <-chan domain.ItemResult {
	// Por ahora devolvemos un canal vacío, ya que la lógica actual no expone un canal
	// En una futura iteración se puede mejorar para exponer el canal real
	ch := make(chan domain.ItemResult)
	close(ch) // Canal cerrado para evitar bloqueos
	return ch
}

// Verificar que implementa la interfaz
var _ domain.WebSocketClient = (*WebSocketClientAdapter)(nil)
