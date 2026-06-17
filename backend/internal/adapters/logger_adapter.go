package adapters

import (
	"log"

	"github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain"
	"github.com/SManriqueDev/poe-tool/backend/internal/logging"
)

// LoggerAdapter adapta el servicio de logging actual al contrato del dominio
type LoggerAdapter struct {
	loggingSvc *logging.Service
}

// NewLoggerAdapter crea un nuevo adaptador para el logger
func NewLoggerAdapter(loggingSvc *logging.Service) *LoggerAdapter {
	return &LoggerAdapter{loggingSvc: loggingSvc}
}

// Info registra un mensaje informativo
func (a *LoggerAdapter) Info(module, message string, metadata map[string]interface{}) error {
	// Detect item-found events and route to LogItemFound for enriched logging
	if module == "livesearch" && metadata != nil {
		if itemID, ok := metadata["item_id"].(string); ok && itemID != "" {
			searchID, _ := metadata["search_id"].(string)
			itemName, _ := metadata["item_name"].(string)

			var price *logging.Price
			if pAmount, hasAmt := metadata["price_amount"].(float64); hasAmt {
				pCurrency, _ := metadata["price_currency"].(string)
				pType, _ := metadata["price_type"].(string)
				price = &logging.Price{
					Amount:   pAmount,
					Currency: pCurrency,
					Type:     pType,
				}
			}

			return a.loggingSvc.LogItemFound(searchID, itemID, itemName, "", "", price, "")
		}
	}

	log.Printf("[%s] INFO: %s %v", module, message, metadata)
	return a.loggingSvc.Log(logging.LogModuleLiveSearch, logging.LogLevelInfo, message, metadata)
}

// Error registra un mensaje de error
func (a *LoggerAdapter) Error(module, message string, metadata map[string]interface{}) error {
	log.Printf("[%s] ERROR: %s %v", module, message, metadata)
	return a.loggingSvc.Log(logging.LogModuleLiveSearch, logging.LogLevelError, message, metadata)
}

// Warning registra un mensaje de advertencia
func (a *LoggerAdapter) Warning(module, message string, metadata map[string]interface{}) error {
	log.Printf("[%s] WARN: %s %v", module, message, metadata)
	return a.loggingSvc.Log(logging.LogModuleLiveSearch, logging.LogLevelWarning, message, metadata)
}

// Debug registra un mensaje de debug
func (a *LoggerAdapter) Debug(module, message string, metadata map[string]interface{}) error {
	log.Printf("[%s] DEBUG: %s %v", module, message, metadata)
	return a.loggingSvc.Log(logging.LogModuleLiveSearch, logging.LogLevelDebug, message, metadata)
}

// Verificar que implementa la interfaz
var _ domain.Logger = (*LoggerAdapter)(nil)
