package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// DomainAPIClient implementa un cliente HTTP puro para APIs externas (PoE API)
type DomainAPIClient struct {
	logger     domain.Logger
	httpClient *http.Client

	// Configuration
	baseURL    string
	userAgent  string
	timeout    time.Duration
	maxRetries int
	retryDelay time.Duration

	// Rate limiting
	rateLimitDelay  time.Duration
	lastRequestTime time.Time
}

// APIResponse representa una respuesta genérica de la API
type APIResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string][]string    `json:"headers"`
	Body       map[string]interface{} `json:"body"`
	RawBody    []byte                 `json:"raw_body,omitempty"`
}

// NewDomainAPIClient crea una nueva instancia del cliente API
func NewDomainAPIClient(logger domain.Logger) *DomainAPIClient {
	return &DomainAPIClient{
		logger:         logger,
		baseURL:        "https://www.pathofexile.com/api",
		userAgent:      "PoeTool/1.0",
		timeout:        30 * time.Second,
		maxRetries:     3,
		retryDelay:     2 * time.Second,
		rateLimitDelay: 100 * time.Millisecond, // Respetar rate limits de PoE API
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchTradeResults obtiene resultados de trade de la API de PoE
func (c *DomainAPIClient) FetchTradeResults(ctx context.Context, searchID string, itemIDs []string) (*APIResponse, error) {
	if len(itemIDs) == 0 {
		return nil, fmt.Errorf("no item IDs provided")
	}

	// Construir URL de fetch
	fetchURL := fmt.Sprintf("%s/trade2/fetch/%s", c.baseURL, strings.Join(itemIDs, ","))

	// Agregar query parameters
	params := url.Values{}
	params.Add("query", searchID)
	params.Add("realm", "poe2") // Por defecto PoE2

	finalURL := fmt.Sprintf("%s?%s", fetchURL, params.Encode())

	c.logger.Info("api", "Fetching trade results", map[string]interface{}{
		"url":      finalURL,
		"item_ids": itemIDs,
	})

	startTime := time.Now()
	response, err := c.makeRequest(ctx, "GET", finalURL, nil)
	duration := time.Since(startTime)

	if err != nil {
		c.logger.Error("api", "Failed to fetch trade results", map[string]interface{}{
			"url":         finalURL,
			"duration_ms": duration.Milliseconds(),
			"error":       err.Error(),
		})
		return nil, err
	}

	c.logger.Info("api", fmt.Sprintf("API GET %s - %d", finalURL, response.StatusCode), map[string]interface{}{
		"url":              finalURL,
		"method":           "GET",
		"status_code":      response.StatusCode,
		"response_time_ms": duration.Milliseconds(),
	})

	return response, nil
}

// SearchTrade realiza una búsqueda de trade en la API de PoE
func (c *DomainAPIClient) SearchTrade(ctx context.Context, query map[string]interface{}) (*APIResponse, error) {
	searchURL := fmt.Sprintf("%s/trade2/search/poe2", c.baseURL)

	c.logger.Info("api", "Performing trade search", map[string]interface{}{
		"url": searchURL,
	})

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	startTime := time.Now()
	response, err := c.makeRequest(ctx, "POST", searchURL, queryBytes)
	duration := time.Since(startTime)

	if err != nil {
		c.logger.Error("api", "Failed to perform trade search", map[string]interface{}{
			"url":         searchURL,
			"duration_ms": duration.Milliseconds(),
			"error":       err.Error(),
		})
		return nil, err
	}

	c.logger.Info("api", fmt.Sprintf("API POST %s - %d", searchURL, response.StatusCode), map[string]interface{}{
		"url":              searchURL,
		"method":           "POST",
		"status_code":      response.StatusCode,
		"response_time_ms": duration.Milliseconds(),
	})

	return response, nil
}

// RequestHideoutVisit solicita visita a hideout a través de la API de PoE
func (c *DomainAPIClient) RequestHideoutVisit(ctx context.Context, itemID, token string) (*APIResponse, error) {
	hideoutURL := fmt.Sprintf("%s/hideout/visit", c.baseURL)

	requestBody := map[string]interface{}{
		"item_id": itemID,
		"token":   token,
	}

	c.logger.Info("api", "Requesting hideout visit", map[string]interface{}{
		"url":     hideoutURL,
		"item_id": itemID,
	})

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	startTime := time.Now()
	response, err := c.makeRequest(ctx, "POST", hideoutURL, bodyBytes)
	duration := time.Since(startTime)

	if err != nil {
		c.logger.Error("api", "Failed to request hideout visit", map[string]interface{}{
			"url":         hideoutURL,
			"item_id":     itemID,
			"duration_ms": duration.Milliseconds(),
			"error":       err.Error(),
		})
		return nil, err
	}

	c.logger.Info("api", fmt.Sprintf("API POST %s - %d", hideoutURL, response.StatusCode), map[string]interface{}{
		"url":              hideoutURL,
		"method":           "POST",
		"status_code":      response.StatusCode,
		"response_time_ms": duration.Milliseconds(),
		"item_id":          itemID,
	})

	return response, nil
}

// makeRequest ejecuta una petición HTTP con reintentos y rate limiting
func (c *DomainAPIClient) makeRequest(ctx context.Context, method, url string, body []byte) (*APIResponse, error) {
	var response *APIResponse
	var err error

	// Aplicar rate limiting
	c.applyRateLimit()

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Info("api", "Retrying request", map[string]interface{}{
				"attempt": attempt,
				"url":     url,
			})

			// Esperar antes del retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay):
				// Continuar con el retry
			}
		}

		response, err = c.executeRequest(ctx, method, url, body)
		if err == nil {
			// Éxito
			return response, nil
		}

		// Verificar si es un error que vale la pena reintentar
		if !c.shouldRetry(response, err) {
			break
		}
	}

	return response, err
}

// executeRequest ejecuta una petición HTTP individual
func (c *DomainAPIClient) executeRequest(ctx context.Context, method, url string, body []byte) (*APIResponse, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Configurar headers
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Ejecutar request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Leer body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Construir respuesta
	apiResponse := &APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		RawBody:    respBody,
	}

	// Intentar parsear JSON si el content-type es adecuado
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		var jsonBody map[string]interface{}
		if err := json.Unmarshal(respBody, &jsonBody); err == nil {
			apiResponse.Body = jsonBody
		}
	}

	// Verificar código de estado
	if resp.StatusCode >= 400 {
		return apiResponse, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	return apiResponse, nil
}

// shouldRetry determina si un error debe ser reintentado
func (c *DomainAPIClient) shouldRetry(response *APIResponse, err error) bool {
	if err == nil {
		return false
	}

	// Retry en errores de red o timeouts
	if strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "connection") {
		return true
	}

	// Retry en códigos de estado 5xx
	if response != nil && response.StatusCode >= 500 {
		return true
	}

	// Retry en rate limit (429)
	if response != nil && response.StatusCode == 429 {
		return true
	}

	return false
}

// applyRateLimit aplica delay entre requests para respetar rate limits
func (c *DomainAPIClient) applyRateLimit() {
	now := time.Now()
	timeSinceLastRequest := now.Sub(c.lastRequestTime)

	if timeSinceLastRequest < c.rateLimitDelay {
		sleepTime := c.rateLimitDelay - timeSinceLastRequest
		time.Sleep(sleepTime)
	}

	c.lastRequestTime = time.Now()
}

// SetTimeout configura el timeout para requests
func (c *DomainAPIClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.httpClient.Timeout = timeout
}

// SetRateLimit configura el delay entre requests
func (c *DomainAPIClient) SetRateLimit(delay time.Duration) {
	c.rateLimitDelay = delay
}

// SetMaxRetries configura el número máximo de reintentos
func (c *DomainAPIClient) SetMaxRetries(maxRetries int) {
	c.maxRetries = maxRetries
}
