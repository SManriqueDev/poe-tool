# Ejemplo de Migración Gradual del Handler

Este archivo muestra cómo migrar gradualmente el handler actual para usar la nueva arquitectura.

## Código Ejemplo (Conceptual)

```go
package livesearch

import (
    "context"
    "log"

    "github.com/SManriqueDev/poe-tool/backend/internal/livesearch/application"
    "github.com/SManriqueDev/poe-tool/backend/internal/livesearch/adapters"
    "github.com/SManriqueDev/poe-tool/backend/internal/logging"
)

// Handler actualizado con nueva arquitectura
type Handler struct {
    // Servicios de aplicación (casos de uso)
    tradeLinkAppSvc   *application.TradeLinkApplicationService
    liveSearchAppSvc  *application.LiveSearchApplicationService

    // Servicio legacy (para compatibilidad durante migración)
    legacySvc *Service
}

func NewHandler(svc *Service, loggingSvc *logging.Service) *Handler {
    // Crear adaptadores
    repoAdapter := adapters.NewRepositoryAdapter(svc.repo)
    loggerAdapter := adapters.NewLoggerAdapter(loggingSvc)

    // Crear servicios de aplicación
    tradeLinkAppSvc := application.NewTradeLinkApplicationService(repoAdapter, loggerAdapter)

    return &Handler{
        tradeLinkAppSvc: tradeLinkAppSvc,
        legacySvc:       svc, // Para funcionalidades no migradas
    }
}

// MIGRADO: Usar servicio de aplicación
func (h *Handler) AddTradeLink(url string, description string) error {
    ctx := context.Background()
    return h.tradeLinkAppSvc.AddTradeLink(ctx, url, description)
}

// MIGRADO: Usar servicio de aplicación
func (h *Handler) ListTradeLinks() ([]TradeLink, error) {
    ctx := context.Background()
    domainTradeLinks, err := h.tradeLinkAppSvc.ListTradeLinks(ctx)
    if err != nil {
        return nil, err
    }

    // Convertir a modelo actual para mantener compatibilidad con frontend
    var tradeLinks []TradeLink
    for _, dtl := range domainTradeLinks {
        tradeLinks = append(tradeLinks, TradeLink{
            ID:          dtl.ID,
            URL:         dtl.URL,
            Description: dtl.Description,
            Selected:    dtl.Selected,
        })
    }

    return tradeLinks, nil
}

// NO MIGRADO: Usar servicio legacy temporalmente
func (h *Handler) StartLiveSearch() []TradeLink {
    return h.legacySvc.StartLiveSearch()
}

// NO MIGRADO: Usar servicio legacy temporalmente
func (h *Handler) StopLiveSearch() {
    h.legacySvc.StopLiveSearch()
}
```

## Ventajas de esta Migración Gradual

### ✅ **Compatibilidad Mantenida**
- Frontend no requiere cambios durante migración
- Funcionalidades existentes siguen funcionando
- Bindings de Wails no se ven afectados

### ✅ **Testing Mejorado**
```go
func TestAddTradeLink(t *testing.T) {
    // Mock de dependencias del dominio
    mockRepo := &MockTradeLinkRepository{}
    mockLogger := &MockLogger{}

    // Servicio testeable sin dependencias externas
    svc := application.NewTradeLinkApplicationService(mockRepo, mockLogger)

    err := svc.AddTradeLink(ctx, "url", "description")
    assert.NoError(t, err)
}
```

### ✅ **Separación de Responsabilidades**
- `Handler`: Solo coordinación y conversión de tipos
- `ApplicationService`: Lógica de casos de uso
- `Domain`: Reglas de negocio puras
- `Adapters`: Integración con servicios existentes

### ✅ **Facilidad de Mantenimiento**
- Cada servicio de aplicación es pequeño y enfocado
- Testing independiente de infraestructura
- Cambios aislados por funcionalidad

## Próximos Pasos

1. **Completar Adaptadores**:
   - WebSocketAdapter
   - EventBusAdapter
   - LiveSearchRepositoryAdapter

2. **Migrar Funcionalidades una por una**:
   - ✅ Trade Link Management
   - ⏳ Live Search Start/Stop
   - ⏳ Hideout Management
   - ⏳ WebSocket Communication

3. **Eliminar Código Legacy**:
   - Una vez todo migrado, eliminar Service monolítico
   - Mantener solo servicios de aplicación específicos
