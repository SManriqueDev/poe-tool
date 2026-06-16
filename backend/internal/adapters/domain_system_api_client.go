package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
	"github.com/corpix/uarand"
)

// DomainSystemAPIClient implementa domain.SystemAPIClient de forma pura
type DomainSystemAPIClient struct {
	logger domain.Logger

	// HTTP client configuration
	httpClient *http.Client
	timeout    time.Duration
	userAgent  string

	// Rate limiting
	rateLimitDelay  time.Duration
	lastRequestTime time.Time

	// Configuration
	maxRetries int
	retryDelay time.Duration
}

// NewDomainSystemAPIClient crea una nueva instancia del cliente de APIs del sistema
func NewDomainSystemAPIClient(logger domain.Logger, config SystemAPIClientConfig) *DomainSystemAPIClient {
	return &DomainSystemAPIClient{
		logger:         logger,
		timeout:        config.Timeout,
		userAgent:      config.UserAgent,
		rateLimitDelay: config.RateLimitDelay,
		maxRetries:     config.MaxRetries,
		retryDelay:     config.RetryDelay,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// SendHideoutRequest envía una request al API de Path of Exile para teleportar al hideout
func (s *DomainSystemAPIClient) SendHideoutRequest(ctx context.Context, hideoutToken, poeSessid string) error {
	if hideoutToken == "" {
		return fmt.Errorf("hideout token is required")
	}
	if poeSessid == "" {
		return fmt.Errorf("POESESSID is required")
	}

	// Rate limiting
	if err := s.applyRateLimit(); err != nil {
		return err
	}

	// Prepare request body
	requestBody := map[string]interface{}{
		"token":    hideoutToken,
		"continue": true,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal hideout request body: %w", err)
	}

	s.logger.Debug("system_api", "Sending hideout request", map[string]interface{}{
		"token_prefix": hideoutToken[:min(20, len(hideoutToken))] + "...",
		"url":          "https://www.pathofexile.com/api/trade2/whisper",
	})

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://www.pathofexile.com/api/trade2/whisper", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create hideout request: %w", err)
	}

	// Set headers to mimic browser request
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", uarand.GetRandom()) // Use random user agent for better success rate
	req.Header.Set("Cookie", "POESESSID="+poeSessid)
	req.Header.Set("Referer", "https://www.pathofexile.com/trade2")
	req.Header.Set("Origin", "https://www.pathofexile.com")

	// Execute request with retry logic
	return s.executeRequestWithRetry(ctx, req, hideoutToken)
}

// IsConnected verifica si el cliente puede conectarse a los APIs del sistema
func (s *DomainSystemAPIClient) IsConnected(ctx context.Context) (bool, error) {
	// Simple connectivity check to Path of Exile website
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.pathofexile.com/api", nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("User-Agent", s.userAgent)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Debug("system_api", "Connectivity check failed", map[string]interface{}{
			"error": err.Error(),
		})
		return false, nil
	}
	defer resp.Body.Close()

	connected := resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound // 404 is also OK for API endpoint

	s.logger.Debug("system_api", "Connectivity check completed", map[string]interface{}{
		"connected":   connected,
		"status_code": resp.StatusCode,
	})

	return connected, nil
}

// GetSystemInfo retorna información del sistema
func (s *DomainSystemAPIClient) GetSystemInfo(ctx context.Context) (*domain.SystemInfo, error) {
	connected, _ := s.IsConnected(ctx)

	systemInfo := &domain.SystemInfo{
		OS:           runtime.GOOS,
		Version:      runtime.Version(),
		Architecture: runtime.GOARCH,
		Environment: map[string]string{
			"GOMAXPROCS": fmt.Sprintf("%d", runtime.GOMAXPROCS(0)),
			"NumCPU":     fmt.Sprintf("%d", runtime.NumCPU()),
		},
		Connected: connected,
	}

	s.logger.Debug("system_api", "System info retrieved", map[string]interface{}{
		"os":           systemInfo.OS,
		"architecture": systemInfo.Architecture,
		"connected":    systemInfo.Connected,
	})

	return systemInfo, nil
}

// ValidatePoeSessid verifica si el POESESSID es válido
func (s *DomainSystemAPIClient) ValidatePoeSessid(ctx context.Context, poeSessid string) error {
	if poeSessid == "" {
		return fmt.Errorf("POESESSID is empty")
	}

	// Simple validation by trying to access a known Path of Exile API endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.pathofexile.com/api/profile", nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Cookie", "POESESSID="+poeSessid)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("system_api", "POESESSID validation request failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("validation request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		s.logger.Warning("system_api", "POESESSID appears to be invalid", map[string]interface{}{
			"status_code": resp.StatusCode,
		})
		return fmt.Errorf("POESESSID appears to be invalid (status: %d)", resp.StatusCode)
	}

	s.logger.Info("system_api", "POESESSID validation successful", map[string]interface{}{
		"status_code": resp.StatusCode,
	})

	return nil
}

// executeRequestWithRetry ejecuta una request HTTP con lógica de retry
func (s *DomainSystemAPIClient) executeRequestWithRetry(ctx context.Context, req *http.Request, hideoutToken string) error {
	var lastErr error

	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		if attempt > 0 {
			s.logger.Warning("system_api", "Retrying hideout request", map[string]interface{}{
				"attempt":      attempt,
				"token_prefix": hideoutToken[:min(20, len(hideoutToken))] + "...",
			})

			// Wait before retry
			select {
			case <-time.After(s.retryDelay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Execute request
		resp, err := s.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("hideout request failed: %w", err)
			continue
		}

		// Read response
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("failed to read hideout response: %w", err)
			continue
		}

		// Check response status
		if resp.StatusCode == http.StatusOK {
			s.logger.Info("system_api", "Hideout request successful", map[string]interface{}{
				"status_code":  resp.StatusCode,
				"token_prefix": hideoutToken[:min(20, len(hideoutToken))] + "...",
				"attempt":      attempt + 1,
			})
			return nil
		}

		// Handle error response
		errorMessage := fmt.Sprintf("HTTP %d", resp.StatusCode)

		// Try to parse error response for more details
		var errorResp struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}

		if json.Unmarshal(respBody, &errorResp) == nil && errorResp.Error.Message != "" {
			errorMessage = errorResp.Error.Message
		}

		lastErr = fmt.Errorf("hideout request failed (status %d): %s", resp.StatusCode, errorMessage)

		// Don't retry on certain status codes
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			s.logger.Error("system_api", "Hideout request failed with auth error, not retrying", map[string]interface{}{
				"status_code": resp.StatusCode,
				"error":       errorMessage,
			})
			break
		}

		s.logger.Warning("system_api", "Hideout request failed, will retry", map[string]interface{}{
			"status_code": resp.StatusCode,
			"error":       errorMessage,
			"attempt":     attempt + 1,
		})
	}

	return fmt.Errorf("hideout request failed after %d attempts: %w", s.maxRetries+1, lastErr)
}

// applyRateLimit aplica rate limiting entre requests
func (s *DomainSystemAPIClient) applyRateLimit() error {
	now := time.Now()
	timeSinceLastRequest := now.Sub(s.lastRequestTime)

	if timeSinceLastRequest < s.rateLimitDelay {
		waitTime := s.rateLimitDelay - timeSinceLastRequest
		s.logger.Debug("system_api", "Applying rate limit", map[string]interface{}{
			"wait_time_ms": waitTime.Milliseconds(),
		})
		time.Sleep(waitTime)
	}

	s.lastRequestTime = time.Now()
	return nil
}


