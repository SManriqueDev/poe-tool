// backend/internal/livesearch/websocket_client.go
package livesearch

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

type WebSocketClient struct{}

func NewWebSocketClient() *WebSocketClient {
	return &WebSocketClient{}
}

func (c *WebSocketClient) Connect(ctx context.Context, link TradeLink, poeSess string) (*websocket.Conn, *http.Response, error) {
	wsURL := url.URL{
		Scheme: "wss",
		Host:   "www.pathofexile.com",
		Path:   "/api/trade2/live/poe2/" + link.League + "/" + link.SearchID,
	}
	header := http.Header{}
	header.Set("Cookie", "POESESSID="+poeSess)
	header.Set("Origin", "https://www.pathofexile.com")
	header.Set("User-Agent", "Mozilla/5.0 ...")
	header.Set("Content-Type", "application/json")

	return websocket.DefaultDialer.DialContext(ctx, wsURL.String(), header)
}
