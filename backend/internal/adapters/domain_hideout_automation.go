package adapters

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
)

// DomainHideoutAutomation implementa domain.HideoutAutomation de forma pura
type DomainHideoutAutomation struct {
	logger domain.Logger

	// Queue management
	queue       []domain.HideoutQueueItem
	queueMu     sync.RWMutex
	processing  bool
	currentItem *domain.HideoutQueueItem
	stateMu     sync.RWMutex

	// Configuration
	maxRetries   int
	retryDelay   time.Duration
	processDelay time.Duration
	maxQueueSize int

	// Processing control
	stopChan  chan struct{}
	doneChan  chan struct{}
	controlMu sync.Mutex
	running   bool

	// Dependencies
	systemAPIClient domain.SystemAPIClient
	settingsRepo    domain.LiveSearchRepository
}

// NewDomainHideoutAutomation crea una nueva instancia de automatización de hideout
func NewDomainHideoutAutomation(
	logger domain.Logger,
	systemAPIClient domain.SystemAPIClient,
	settingsRepo domain.LiveSearchRepository,
) *DomainHideoutAutomation {
	return &DomainHideoutAutomation{
		logger:          logger,
		systemAPIClient: systemAPIClient,
		settingsRepo:    settingsRepo,
		queue:           make([]domain.HideoutQueueItem, 0),
		maxRetries:      3,
		retryDelay:      2 * time.Second,
		processDelay:    8 * time.Second, // Realistic trading time between hideout visits
		maxQueueSize:    50,              // Prevent memory issues with large queues
		stopChan:        make(chan struct{}),
		doneChan:        make(chan struct{}),
	}
}

// ProcessHideoutVisit procesa inmediatamente una visita a hideout (sin cola)
func (h *DomainHideoutAutomation) ProcessHideoutVisit(ctx context.Context, hideoutToken, itemID string) error {
	if hideoutToken == "" {
		return fmt.Errorf("hideout token is required")
	}

	// Check if hideout automation is enabled
	settings, err := h.settingsRepo.GetHideoutSettings(ctx)
	if err != nil {
		h.logger.Error("hideout_automation", "Failed to get hideout settings", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	if !settings.Enabled {
		h.logger.Debug("hideout_automation", "Hideout automation is disabled, skipping visit", map[string]interface{}{
			"item_id": itemID,
		})
		return nil
	}

	// Get POESESSID from settings
	poeSessid, err := h.settingsRepo.GetSetting(ctx, "poesessid")
	if err != nil {
		return fmt.Errorf("failed to get POESESSID: %w", err)
	}

	poeSessidStr, ok := poeSessid.(string)
	if !ok || poeSessidStr == "" {
		return fmt.Errorf("POESESSID not configured or invalid")
	}

	h.logger.Info("hideout_automation", "Processing immediate hideout visit", map[string]interface{}{
		"item_id":      itemID,
		"token_prefix": hideoutToken[:min(20, len(hideoutToken))] + "...",
	})

	// Execute hideout request
	return h.executeHideoutRequest(ctx, hideoutToken, poeSessidStr, itemID)
}

// QueueHideoutVisit añade una visita a la cola de hideout
func (h *DomainHideoutAutomation) QueueHideoutVisit(ctx context.Context, hideoutToken, itemID string) error {
	if hideoutToken == "" {
		return fmt.Errorf("hideout token is required")
	}

	// Check if hideout automation is enabled
	settings, err := h.settingsRepo.GetHideoutSettings(ctx)
	if err != nil {
		return err
	}

	if !settings.Enabled {
		h.logger.Debug("hideout_automation", "Hideout automation is disabled, skipping queue", map[string]interface{}{
			"item_id": itemID,
		})
		return nil
	}

	h.queueMu.Lock()
	defer h.queueMu.Unlock()

	// Check queue size
	if len(h.queue) >= h.maxQueueSize {
		h.logger.Warning("hideout_automation", "Queue is full, removing oldest item", map[string]interface{}{
			"queue_size":   len(h.queue),
			"max_size":     h.maxQueueSize,
			"dropped_item": itemID,
		})
		// Remove oldest item
		h.queue = h.queue[1:]
	}

	// Create queue item
	queueItem := domain.HideoutQueueItem{
		Token:     hideoutToken,
		ItemID:    itemID,
		Timestamp: time.Now(),
		Priority:  false,
		Retries:   0,
	}

	// Add to queue
	h.queue = append(h.queue, queueItem)

	h.logger.Info("hideout_automation", "Item added to hideout queue", map[string]interface{}{
		"item_id":      itemID,
		"queue_size":   len(h.queue),
		"token_prefix": hideoutToken[:min(20, len(hideoutToken))] + "...",
	})

	return nil
}

// GetQueueSize retorna el tamaño actual de la cola
func (h *DomainHideoutAutomation) GetQueueSize(ctx context.Context) (int, error) {
	h.queueMu.RLock()
	defer h.queueMu.RUnlock()
	return len(h.queue), nil
}

// IsProcessing verifica si actualmente se está procesando un item
func (h *DomainHideoutAutomation) IsProcessing(ctx context.Context) (bool, error) {
	h.stateMu.RLock()
	defer h.stateMu.RUnlock()
	return h.processing, nil
}

// StartProcessingQueue inicia el procesamiento de la cola
func (h *DomainHideoutAutomation) StartProcessingQueue(ctx context.Context) error {
	h.controlMu.Lock()
	defer h.controlMu.Unlock()

	if h.running {
		return fmt.Errorf("queue processing is already running")
	}

	h.running = true
	h.stopChan = make(chan struct{})
	h.doneChan = make(chan struct{})

	h.logger.Info("hideout_automation", "Starting hideout queue processing", nil)

	go h.processQueueWorker(ctx)

	return nil
}

// StopProcessingQueue detiene el procesamiento de la cola
func (h *DomainHideoutAutomation) StopProcessingQueue(ctx context.Context) error {
	h.controlMu.Lock()
	defer h.controlMu.Unlock()

	if !h.running {
		return nil
	}

	close(h.stopChan)
	<-h.doneChan
	h.running = false

	h.logger.Info("hideout_automation", "Hideout queue processing stopped", nil)
	return nil
}

// ClearQueue limpia completamente la cola
func (h *DomainHideoutAutomation) ClearQueue(ctx context.Context) error {
	h.queueMu.Lock()
	defer h.queueMu.Unlock()

	queueSize := len(h.queue)
	h.queue = h.queue[:0] // Clear the slice

	h.logger.Info("hideout_automation", "Hideout queue cleared", map[string]interface{}{
		"cleared_items": queueSize,
	})

	return nil
}

// processQueueWorker es el worker que procesa la cola continuamente
func (h *DomainHideoutAutomation) processQueueWorker(ctx context.Context) {
	defer close(h.doneChan)

	ticker := time.NewTicker(1 * time.Second) // Check queue every second
	defer ticker.Stop()

	for {
		select {
		case <-h.stopChan:
			h.logger.Debug("hideout_automation", "Queue processing stopped by signal", nil)
			return
		case <-ctx.Done():
			h.logger.Debug("hideout_automation", "Queue processing stopped by context cancellation", nil)
			return
		case <-ticker.C:
			h.processNextItemIfAvailable(ctx)
		}
	}
}

// processNextItemIfAvailable procesa el siguiente item de la cola si está disponible
func (h *DomainHideoutAutomation) processNextItemIfAvailable(ctx context.Context) {
	// Check if already processing
	h.stateMu.RLock()
	isProcessing := h.processing
	h.stateMu.RUnlock()

	if isProcessing {
		return
	}

	// Get next item from queue
	h.queueMu.Lock()
	if len(h.queue) == 0 {
		h.queueMu.Unlock()
		return
	}

	item := h.queue[0]
	h.queue = h.queue[1:]
	h.queueMu.Unlock()

	// Set processing state
	h.stateMu.Lock()
	h.processing = true
	h.currentItem = &item
	h.stateMu.Unlock()

	// Process item
	go func() {
		defer func() {
			h.stateMu.Lock()
			h.processing = false
			h.currentItem = nil
			h.stateMu.Unlock()
		}()

		h.processQueueItem(ctx, item)

		// Wait before processing next item
		select {
		case <-time.After(h.processDelay):
		case <-h.stopChan:
		case <-ctx.Done():
		}
	}()
}

// processQueueItem procesa un item específico de la cola
func (h *DomainHideoutAutomation) processQueueItem(ctx context.Context, item domain.HideoutQueueItem) {
	// Get POESESSID from settings
	poeSessid, err := h.settingsRepo.GetSetting(ctx, "poesessid")
	if err != nil {
		h.logger.Error("hideout_automation", "Failed to get POESESSID for queue item", map[string]interface{}{
			"item_id": item.ItemID,
			"error":   err.Error(),
		})
		return
	}

	poeSessidStr, ok := poeSessid.(string)
	if !ok || poeSessidStr == "" {
		h.logger.Error("hideout_automation", "POESESSID not configured for queue item", map[string]interface{}{
			"item_id": item.ItemID,
		})
		return
	}

	h.logger.Info("hideout_automation", "Processing queue item", map[string]interface{}{
		"item_id":      item.ItemID,
		"retries":      item.Retries,
		"token_prefix": item.Token[:min(20, len(item.Token))] + "...",
	})

	// Execute hideout request with retry logic
	err = h.executeHideoutRequestWithRetry(ctx, item.Token, poeSessidStr, item.ItemID, item.Retries)
	if err != nil {
		h.logger.Error("hideout_automation", "Failed to process queue item after all retries", map[string]interface{}{
			"item_id": item.ItemID,
			"retries": item.Retries,
			"error":   err.Error(),
		})
	}
}

// executeHideoutRequest ejecuta una request de hideout
func (h *DomainHideoutAutomation) executeHideoutRequest(ctx context.Context, hideoutToken, poeSessid, itemID string) error {
	err := h.systemAPIClient.SendHideoutRequest(ctx, hideoutToken, poeSessid)
	if err != nil {
		h.logger.Error("hideout_automation", "Hideout request failed", map[string]interface{}{
			"item_id": itemID,
			"error":   err.Error(),
		})
		return err
	}

	h.logger.Info("hideout_automation", "Hideout visit completed successfully", map[string]interface{}{
		"item_id": itemID,
	})

	return nil
}

// executeHideoutRequestWithRetry ejecuta una request de hideout con retry logic
func (h *DomainHideoutAutomation) executeHideoutRequestWithRetry(ctx context.Context, hideoutToken, poeSessid, itemID string, currentRetries int) error {
	err := h.executeHideoutRequest(ctx, hideoutToken, poeSessid, itemID)
	if err == nil {
		return nil
	}

	// Check if we should retry
	if currentRetries >= h.maxRetries {
		return fmt.Errorf("max retries exceeded: %w", err)
	}

	// Wait before retry
	select {
	case <-time.After(h.retryDelay):
	case <-ctx.Done():
		return ctx.Err()
	}

	// Retry
	h.logger.Warning("hideout_automation", "Retrying hideout request", map[string]interface{}{
		"item_id": itemID,
		"retry":   currentRetries + 1,
		"error":   err.Error(),
	})

	return h.executeHideoutRequestWithRetry(ctx, hideoutToken, poeSessid, itemID, currentRetries+1)
}

// GetCurrentItem retorna el item actualmente siendo procesado
func (h *DomainHideoutAutomation) GetCurrentItem() *domain.HideoutQueueItem {
	h.stateMu.RLock()
	defer h.stateMu.RUnlock()
	return h.currentItem
}

// GetQueueSnapshot retorna una copia de la cola actual
func (h *DomainHideoutAutomation) GetQueueSnapshot() []domain.HideoutQueueItem {
	h.queueMu.RLock()
	defer h.queueMu.RUnlock()

	snapshot := make([]domain.HideoutQueueItem, len(h.queue))
	copy(snapshot, h.queue)
	return snapshot
}

// SetConfiguration actualiza la configuración del sistema
func (h *DomainHideoutAutomation) SetConfiguration(maxRetries int, retryDelay, processDelay time.Duration, maxQueueSize int) {
	h.maxRetries = maxRetries
	h.retryDelay = retryDelay
	h.processDelay = processDelay
	h.maxQueueSize = maxQueueSize

	h.logger.Info("hideout_automation", "Configuration updated", map[string]interface{}{
		"max_retries":    maxRetries,
		"retry_delay":    retryDelay,
		"process_delay":  processDelay,
		"max_queue_size": maxQueueSize,
	})
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
