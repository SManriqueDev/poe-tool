package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
	"github.com/corpix/uarand"
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

	// Autenticación
	poeSessID string
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

// SetPOESESSID configura el cookie de sesión de Path of Exile
func (c *DomainWebSocketClient) SetPOESESSID(poeSessID string) {
	c.poeSessID = poeSessID
}

// Connect establece conexión con el WebSocket de Path of Exile para un trade link específico
func (c *DomainWebSocketClient) Connect(ctx context.Context, tradeURL string) error {
	// Si ya estamos conectados, no hacer nada
	if c.IsConnected() {
		return nil
	}

	c.logger.Info("websocket", "WebSocket connection requested", map[string]interface{}{
		"trade_url": tradeURL,
	})

	// La conexión real se hará en Subscribe() para cada search ID específico
	return nil
}

// connectToSearchID establece conexión específica para un search ID
func (c *DomainWebSocketClient) connectToSearchID(ctx context.Context, searchID, league string) error {
	// Construir URL exactamente como el servicio legacy
	wsURL := url.URL{
		Scheme: "wss",
		Host:   "www.pathofexile.com",
		Path:   "/api/trade2/live/poe2/" + league + "/" + searchID,
	}

	c.logger.Info("websocket", "Connecting to Path of Exile WebSocket", map[string]interface{}{
		"url":       wsURL.String(),
		"search_id": searchID,
		"league":    league,
	})

	// Configurar headers exactamente como el servicio legacy
	header := http.Header{}
	if c.poeSessID != "" {
		header.Set("Cookie", "POESESSID="+c.poeSessID)
		c.logger.Info("websocket", "Using POESESSID for authentication", map[string]interface{}{
			"poesessid_length": len(c.poeSessID),
		})
	} else {
		c.logger.Warning("websocket", "No POESESSID available for authentication", nil)
	}
	header.Set("Origin", "https://www.pathofexile.com")
	header.Set("User-Agent", uarand.GetRandom())
	header.Set("Content-Type", "application/json")

	// Conectar al WebSocket exactamente como el legacy service
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL.String(), header)
	if err != nil {
		c.logger.Error("websocket", "Failed to connect to WebSocket", map[string]interface{}{
			"url":       wsURL.String(),
			"search_id": searchID,
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to connect to websocket: %w", err)
	}

	// Configurar timeouts
	conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))

	// CRÍTICO: Manejar autenticación inicial como en el servicio legacy
	var authResp struct {
		Auth bool `json:"auth"`
	}
	if err := conn.ReadJSON(&authResp); err != nil {
		c.logger.Error("websocket", "Failed to read auth response", map[string]interface{}{
			"url":       wsURL.String(),
			"search_id": searchID,
			"error":     err.Error(),
		})
		conn.Close()
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	if !authResp.Auth {
		c.logger.Error("websocket", "Authentication failed", map[string]interface{}{
			"url":       wsURL.String(),
			"search_id": searchID,
		})
		conn.Close()
		return fmt.Errorf("websocket authentication failed - check POESESSID")
	}

	c.logger.Info("websocket", "WebSocket authenticated successfully", map[string]interface{}{
		"url":       wsURL.String(),
		"search_id": searchID,
	})

	// Almacenar conexión usando la URL como string
	wsURLStr := wsURL.String()
	c.connMu.Lock()
	c.connections[wsURLStr] = conn
	c.connMu.Unlock()

	c.connState.Lock()
	c.connected = true
	c.connState.Unlock()

	c.logger.Info("websocket", "WebSocket connected successfully", map[string]interface{}{
		"url":       wsURLStr,
		"search_id": searchID,
	})

	// Iniciar goroutines para manejo de mensajes
	go c.handleMessages(ctx, conn, wsURLStr)
	go c.pingHandler(ctx, conn, wsURLStr)

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
	c.logger.Info("websocket", "Subscribing to search updates", map[string]interface{}{
		"search_id": searchID,
	})

	// Usar la liga exacta como en el servicio legacy
	league := "Rise of the Abyssal" // Sin URL encoding

	// Conectar específicamente para este search ID
	if err := c.connectToSearchID(ctx, searchID, league); err != nil {
		return fmt.Errorf("failed to connect for search ID %s: %w", searchID, err)
	}

	c.logger.Info("websocket", "Successfully subscribed to search", map[string]interface{}{
		"search_id": searchID,
		"league":    league,
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

			// Leer mensaje crudo como en el servicio legacy
			_, messageBytes, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Error("websocket", "Unexpected close error", map[string]interface{}{
						"url":   wsURL,
						"error": err.Error(),
					})
				}
				return
			}

			// Deserializar el mensaje
			var message map[string]interface{}
			if err := json.Unmarshal(messageBytes, &message); err != nil {
				c.logger.Error("websocket", "Failed to unmarshal message", map[string]interface{}{
					"url":   wsURL,
					"error": err.Error(),
				})
				continue
			}

			c.processMessage(message)
		}
	}
}

// processMessage procesa un mensaje del WebSocket según el formato de la API de PoE
func (c *DomainWebSocketClient) processMessage(message map[string]interface{}) {
	// La API de PoE envía mensajes con un array "new" que contiene IDs de nuevos items
	if newItems, ok := message["new"].([]interface{}); ok && len(newItems) > 0 {
		c.logger.Info("websocket", "Received new items", map[string]interface{}{
			"count": len(newItems),
		})

		for _, itemInterface := range newItems {
			if itemID, ok := itemInterface.(string); ok {
				// Crear ItemResult básico para cada nuevo item
				itemResult := &domain.ItemResult{
					ID:       itemID,
					SearchID: "", // Se puede extraer del contexto si es necesario
					Item:     nil,
					Listing:  nil,
				}

				// Enviar al canal si está abierto
				c.channelMu.RLock()
				if !c.channelClosed {
					select {
					case c.messageChannel <- *itemResult:
						c.logger.Info("websocket", "New item sent to channel", map[string]interface{}{
							"item_id": itemID,
						})
					default:
						// Canal lleno, loggear warning
						c.logger.Warning("websocket", "Message channel full, dropping item", map[string]interface{}{
							"item_id": itemID,
						})
					}
				}
				c.channelMu.RUnlock()
			}
		}
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
