package adapters

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
	"github.com/corpix/uarand"
	"github.com/gorilla/websocket"
)

// DomainWebSocketClient implementa domain.WebSocketClient de forma pura
type DomainWebSocketClient struct {
	logger domain.Logger

	// HTTP client for fetching item details
	httpClient *http.Client

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
func NewDomainWebSocketClient(logger domain.Logger, config WebSocketConfig) *DomainWebSocketClient {
	return &DomainWebSocketClient{
		logger:         logger,
		httpClient: &http.Client{
			Timeout: config.ReadTimeout,
			Transport: &http.Transport{
				TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
			},
		},
		connections:    make(map[string]*websocket.Conn),
		messageChannel: make(chan domain.ItemResult, 100), // Buffer de 100 mensajes
		maxRetries:     config.MaxRetries,
		retryDelay:     config.RetryDelay,
		pingInterval:   config.PingInterval,
		readTimeout:    config.ReadTimeout,
		writeTimeout:   config.WriteTimeout,
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

	// Iniciar goroutine para manejo de mensajes (sin ping - PoE rechaza control frames)
	go c.handleMessages(ctx, conn, wsURLStr)

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
func (c *DomainWebSocketClient) Subscribe(ctx context.Context, searchID, league string) error {
	c.logger.Info("websocket", "Subscribing to search updates", map[string]interface{}{
		"search_id": searchID,
		"league":    league,
	})

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

			log.Printf("[WS] Raw message received from %s: %s", wsURL, string(messageBytes))

			c.processMessage(ctx, message)
		}
	}
}

// processMessage procesa un mensaje del WebSocket según el formato de la API de PoE 2
func (c *DomainWebSocketClient) processMessage(ctx context.Context, message map[string]interface{}) {
	// PoE 2 format: {"result": "<JWT>", "count": N}
	resultJWT, ok := message["result"].(string)
	if !ok || resultJWT == "" {
		log.Printf("[WS] Unexpected message format (no 'result' field): %v", message)
		return
	}

	log.Printf("[WS] PoE2 live result received, count=%v, jwt_length=%d", message["count"], len(resultJWT))

	c.logger.Info("websocket", "Received PoE 2 live search result", map[string]interface{}{
		"result_length": len(resultJWT),
	})

	// Extract iss from JWT payload (search ID)
	parts := strings.Split(resultJWT, ".")
	if len(parts) < 2 {
		c.logger.Error("websocket", "Invalid JWT format", map[string]interface{}{
			"parts_count": len(parts),
		})
		return
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		c.logger.Error("websocket", "Failed to decode JWT payload", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	var jwtPayload struct {
		Iss string `json:"iss"`
		Exp int64  `json:"exp"`
	}
	if err := json.Unmarshal(payloadBytes, &jwtPayload); err != nil {
		c.logger.Error("websocket", "Failed to parse JWT payload", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.logger.Info("websocket", "Extracted search ID from JWT", map[string]interface{}{
		"search_id": jwtPayload.Iss,
	})

	// Fetch item details from PoE API
	items, err := c.fetchTradeItems(ctx, resultJWT, jwtPayload.Iss)
	if err != nil {
		c.logger.Error("websocket", "Failed to fetch trade items", map[string]interface{}{
			"search_id": jwtPayload.Iss,
			"error":     err.Error(),
		})
		return
	}

	c.logger.Info("websocket", "Fetched trade items", map[string]interface{}{
		"count":       len(items),
		"search_id":   jwtPayload.Iss,
	})

	// Send items to channel
	c.channelMu.RLock()
	defer c.channelMu.RUnlock()

	if c.channelClosed {
		return
	}

	for _, item := range items {
		select {
		case c.messageChannel <- item:
			c.logger.Info("websocket", "Item sent to channel", map[string]interface{}{
				"item_id":   item.ID,
				"search_id": item.SearchID,
			})
		default:
			c.logger.Warning("websocket", "Message channel full, dropping item", map[string]interface{}{
				"item_id": item.ID,
			})
		}
	}
}

// fetchTradeItems fetches item details from PoE 2 trade API
func (c *DomainWebSocketClient) fetchTradeItems(ctx context.Context, resultJWT, searchID string) ([]domain.ItemResult, error) {
	fetchURL := fmt.Sprintf("https://www.pathofexile.com/api/trade2/fetch/%s?query=%s&realm=poe2",
		resultJWT,
		url.QueryEscape(searchID),
	)

	log.Printf("[WS] fetchTradeItems: GET %s", fetchURL)
	c.logger.Info("websocket", "Fetching trade items", map[string]interface{}{
		"search_id": searchID,
	})

	req, err := http.NewRequestWithContext(ctx, "GET", fetchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.poeSessID != "" {
		req.Header.Set("Cookie", "POESESSID="+c.poeSessID)
	}
	req.Header.Set("User-Agent", uarand.GetRandom())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://www.pathofexile.com")

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		log.Printf("[WS] fetchTradeItems: failed after %v: %v", elapsed, err)
		return nil, fmt.Errorf("failed to fetch items: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[WS] fetchTradeItems: HTTP %d after %v for search_id=%s", resp.StatusCode, elapsed, searchID)

	if resp.StatusCode == http.StatusUnauthorized {
		log.Printf("[WS] fetchTradeItems: authentication failed (401) for search_id=%s", searchID)
		return nil, fmt.Errorf("authentication failed - check POESESSID")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[WS] fetchTradeItems: HTTP %d for search_id=%s, body=%s", resp.StatusCode, searchID, string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var fetchResp struct {
		Result []domain.ItemResult `json:"result"`
	}
	if err := json.Unmarshal(body, &fetchResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Attach search ID to each item
	for i := range fetchResp.Result {
		fetchResp.Result[i].SearchID = searchID
	}

	return fetchResp.Result, nil
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
