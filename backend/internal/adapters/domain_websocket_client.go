package adapters

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
	"github.com/gorilla/websocket"
)

// DomainWebSocketClient implementa domain.WebSocketClient de forma pura
type DomainWebSocketClient struct {
	logger domain.Logger

	// Conexiones WebSocket
	connections map[string]*websocket.Conn
	connMu      sync.RWMutex

	// Canal de mensajes para los items encontrados
	messageChannel chan domain.ItemResult
	channelClosed  bool
	channelMu      sync.RWMutex

	// Estado de conexión
	connected bool
	connState sync.RWMutex

	// Configuración
	maxRetries   int
	retryDelay   time.Duration
	pingInterval time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
}

// NewDomainWebSocketClient crea una nueva instancia del cliente WebSocket
func NewDomainWebSocketClient(logger domain.Logger) *DomainWebSocketClient {
	return &DomainWebSocketClient{
		logger:         logger,
		connections:    make(map[string]*websocket.Conn),
		messageChannel: make(chan domain.ItemResult, 100), // Buffer de 100 mensajes
		maxRetries:     3,
		retryDelay:     5 * time.Second,
		pingInterval:   30 * time.Second,
		readTimeout:    60 * time.Second,
		writeTimeout:   10 * time.Second,
	}
}

// Connect establece conexión con el WebSocket de Path of Exile
func (c *DomainWebSocketClient) Connect(ctx context.Context, tradeURL string) error {
	c.connState.Lock()
	defer c.connState.Unlock()

	if c.connected {
		return nil // Ya conectado
	}

	// Construir URL de WebSocket de PoE
	wsURL, err := c.buildWebSocketURL(tradeURL)
	if err != nil {
		return fmt.Errorf("failed to build websocket URL: %w", err)
	}

	c.logger.Info("websocket", "Connecting to Path of Exile WebSocket", map[string]interface{}{
		"url": wsURL,
	})

	// Configurar dialer con timeouts
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 30 * time.Second

	// Conectar al WebSocket
	conn, _, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		c.logger.Error("websocket", "Failed to connect to WebSocket", map[string]interface{}{
			"url":   wsURL,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to connect to websocket: %w", err)
	}

	// Configurar timeouts
	conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))

	// Almacenar conexión
	c.connMu.Lock()
	c.connections[wsURL] = conn
	c.connMu.Unlock()

	c.connected = true

	c.logger.Info("websocket", "WebSocket connected successfully", map[string]interface{}{
		"url": wsURL,
	})

	// Iniciar goroutines para manejo de mensajes
	go c.handleMessages(ctx, conn, wsURL)
	go c.pingHandler(ctx, conn, wsURL)

	return nil
}

// Disconnect cierra la conexión WebSocket
func (c *DomainWebSocketClient) Disconnect(ctx context.Context) error {
	c.connState.Lock()
	defer c.connState.Unlock()

	if !c.connected {
		return nil // Ya desconectado
	}

	c.connMu.Lock()
	defer c.connMu.Unlock()

	// Cerrar todas las conexiones
	for url, conn := range c.connections {
		if conn != nil {
			conn.Close()
			c.logger.Info("websocket", "WebSocket connection closed", map[string]interface{}{
				"url": url,
			})
		}
	}

	// Limpiar conexiones
	c.connections = make(map[string]*websocket.Conn)
	c.connected = false

	// Cerrar canal de mensajes
	c.channelMu.Lock()
	if !c.channelClosed {
		close(c.messageChannel)
		c.channelClosed = true
	}
	c.channelMu.Unlock()

	c.logger.Info("websocket", "All WebSocket connections disconnected", nil)
	return nil
}

// Subscribe se suscribe a actualizaciones de un search ID específico
func (c *DomainWebSocketClient) Subscribe(ctx context.Context, searchID string) error {
	if !c.connected {
		return fmt.Errorf("not connected to websocket")
	}

	c.logger.Info("websocket", "Subscribing to search updates", map[string]interface{}{
		"search_id": searchID,
	})

	// Mensaje de suscripción para PoE API
	subscribeMsg := map[string]interface{}{
		"type":      "subscribe",
		"search_id": searchID,
	}

	// Enviar mensaje de suscripción a todas las conexiones activas
	c.connMu.RLock()
	defer c.connMu.RUnlock()

	for url, conn := range c.connections {
		if conn != nil {
			conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
			if err := conn.WriteJSON(subscribeMsg); err != nil {
				c.logger.Error("websocket", "Failed to send subscribe message", map[string]interface{}{
					"url":       url,
					"search_id": searchID,
					"error":     err.Error(),
				})
				return err
			}
		}
	}

	c.logger.Info("websocket", "Successfully subscribed to search", map[string]interface{}{
		"search_id": searchID,
	})

	return nil
}

// Unsubscribe cancela la suscripción a un search ID
func (c *DomainWebSocketClient) Unsubscribe(ctx context.Context, searchID string) error {
	if !c.connected {
		return fmt.Errorf("not connected to websocket")
	}

	c.logger.Info("websocket", "Unsubscribing from search updates", map[string]interface{}{
		"search_id": searchID,
	})

	// Mensaje de desuscripción para PoE API
	unsubscribeMsg := map[string]interface{}{
		"type":      "unsubscribe",
		"search_id": searchID,
	}

	// Enviar mensaje de desuscripción a todas las conexiones activas
	c.connMu.RLock()
	defer c.connMu.RUnlock()

	for url, conn := range c.connections {
		if conn != nil {
			conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
			if err := conn.WriteJSON(unsubscribeMsg); err != nil {
				c.logger.Error("websocket", "Failed to send unsubscribe message", map[string]interface{}{
					"url":       url,
					"search_id": searchID,
					"error":     err.Error(),
				})
				return err
			}
		}
	}

	c.logger.Info("websocket", "Successfully unsubscribed from search", map[string]interface{}{
		"search_id": searchID,
	})

	return nil
}

// IsConnected verifica si hay conexiones activas
func (c *DomainWebSocketClient) IsConnected() bool {
	c.connState.RLock()
	defer c.connState.RUnlock()
	return c.connected
}

// GetMessageChannel retorna el canal de mensajes de items
func (c *DomainWebSocketClient) GetMessageChannel() <-chan domain.ItemResult {
	return c.messageChannel
}

// buildWebSocketURL construye la URL del WebSocket a partir de una URL de trade
func (c *DomainWebSocketClient) buildWebSocketURL(tradeURL string) (string, error) {
	if tradeURL == "" {
		// URL por defecto para PoE2
		return "wss://www.pathofexile.com/api/trade2/live/poe2", nil
	}

	// Determinar realm (poe1 vs poe2) basado en la URL
	realm := "poe2"
	if strings.Contains(tradeURL, "/trade/search/") {
		realm = "poe1"
	}

	// Construir URL de WebSocket
	wsURL := fmt.Sprintf("wss://www.pathofexile.com/api/trade2/live/%s", realm)

	return wsURL, nil
}

// extractSearchID extrae el search ID de una URL de trade
func (c *DomainWebSocketClient) extractSearchID(tradeURL string) string {
	// Regex para extraer search ID de URLs de PoE
	re := regexp.MustCompile(`/search/[^/]+/([a-zA-Z0-9_-]+)`)
	matches := re.FindStringSubmatch(tradeURL)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// handleMessages maneja los mensajes entrantes del WebSocket
func (c *DomainWebSocketClient) handleMessages(ctx context.Context, conn *websocket.Conn, wsURL string) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error("websocket", "Message handler panic", map[string]interface{}{
				"url":   wsURL,
				"error": r,
			})
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Configurar timeout de lectura
			conn.SetReadDeadline(time.Now().Add(c.readTimeout))

			var message map[string]interface{}
			if err := conn.ReadJSON(&message); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Error("websocket", "Unexpected close error", map[string]interface{}{
						"url":   wsURL,
						"error": err.Error(),
					})
				}
				return
			}

			c.processMessage(message)
		}
	}
}

// processMessage procesa un mensaje del WebSocket y lo convierte a ItemResult
func (c *DomainWebSocketClient) processMessage(message map[string]interface{}) {
	// Verificar si es un mensaje de nuevo item
	if msgType, ok := message["type"].(string); ok && msgType == "new_item" {
		// Convertir mensaje a ItemResult
		itemResult := c.messageToItemResult(message)
		if itemResult != nil {
			// Enviar al canal si está abierto
			c.channelMu.RLock()
			if !c.channelClosed {
				select {
				case c.messageChannel <- *itemResult:
					// Enviado exitosamente
				default:
					// Canal lleno, loggear warning
					c.logger.Warning("websocket", "Message channel full, dropping item", map[string]interface{}{
						"item_id": itemResult.ID,
					})
				}
			}
			c.channelMu.RUnlock()
		}
	}
}

// messageToItemResult convierte un mensaje del WebSocket a un ItemResult del dominio
func (c *DomainWebSocketClient) messageToItemResult(message map[string]interface{}) *domain.ItemResult {
	// Extraer campos del mensaje de PoE API
	itemID, _ := message["item_id"].(string)
	if itemID == "" {
		return nil
	}

	searchID, _ := message["search_id"].(string)
	item := message["item"]
	listing := message["listing"]

	// Construir ItemResult según el modelo del dominio
	return &domain.ItemResult{
		ID:       itemID,
		SearchID: searchID,
		Item:     item,
		Listing:  listing,
	}
}

// pingHandler envía pings periódicos para mantener la conexión viva
func (c *DomainWebSocketClient) pingHandler(ctx context.Context, conn *websocket.Conn, wsURL string) {
	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				c.logger.Error("websocket", "Failed to send ping", map[string]interface{}{
					"url":   wsURL,
					"error": err.Error(),
				})
				return
			}
		}
	}
}
