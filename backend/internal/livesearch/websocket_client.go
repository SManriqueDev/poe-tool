// backend/internal/livesearch/websocket_client.go
package livesearch

import (
	"context"
	"log"
	"net/http"
	"net/url"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
	"github.com/corpix/uarand"
	"github.com/gorilla/websocket"
)

type WebSocketClient struct{}

func NewWebSocketClient() *WebSocketClient {
	return &WebSocketClient{}
}

func (c *WebSocketClient) Connect(ctx context.Context, link domain.TradeLink, poeSess string) (*websocket.Conn, *http.Response, error) {
	wsURL := url.URL{
		Scheme: "wss",
		Host:   "www.pathofexile.com",
		Path:   "/api/trade2/live/poe2/" + link.League + "/" + link.SearchID,
	}
	log.Println("Connecting to WebSocket URL:", wsURL.String())
	header := http.Header{}
	header.Set("Cookie", "POESESSID="+poeSess)
	header.Set("Origin", "https://www.pathofexile.com")
	header.Set("User-Agent", uarand.GetRandom())
	header.Set("Content-Type", "application/json")

	return websocket.DefaultDialer.DialContext(ctx, wsURL.String(), header)
}
