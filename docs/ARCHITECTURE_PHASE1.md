# Propuesta de Arquitectura Mejorada - Fase 1

## 📁 Nueva Estructura de Carpetas

```
backend/internal/livesearch/
├── domain/              # Entidades y lógica de negocio pura
│   ├── models.go       # Modelos del dominio (TradeLink, ItemResult, etc.)
│   ├── interfaces.go   # Contratos/interfaces del dominio
│   └── usecases.go     # Interfaces de casos de uso
├── application/         # Servicios de aplicación (casos de uso)
│   ├── livesearch_service.go   # Lógica de búsqueda en vivo
│   ├── tradelink_service.go    # Gestión de trade links
│   └── hideout_service.go      # Gestión de hideout
├── adapters/           # Adaptadores para integración
│   ├── repository_adapter.go   # Adapta Repository actual
│   ├── logger_adapter.go       # Adapta Logging service
│   ├── websocket_adapter.go    # Adapta WebSocket client
│   └── eventbus_adapter.go     # Adapta EventBus
├── infrastructure/     # Detalles técnicos (futuro)
│   ├── websocket/     # Implementación WebSocket
│   └── persistence/   # Repositorios concretos
├── service.go         # Service actual (se refactoriza gradualmente)
├── handler.go         # Handler actual (se mantiene)
└── repository.go      # Repository actual (se mantiene)
```

## 🎯 Beneficios de esta Estructura

### **Separación Clara de Responsabilidades**

#### **Domain (Dominio)**
- **Propósito**: Contiene la lógica de negocio pura, sin dependencias externas
- **Contenido**: Entidades, interfaces, reglas de negocio, errores del dominio
- **Ventajas**: Fácil de testear, reutilizable, independiente de frameworks

#### **Application (Aplicación)**
- **Propósito**: Orquesta los casos de uso, coordina entre dominio e infraestructura
- **Contenido**: Servicios de aplicación que implementan casos de uso específicos
- **Ventajas**: Lógica de aplicación claramente definida, fácil testing de casos de uso

#### **Adapters (Adaptadores)**
- **Propósito**: Traducen entre el dominio y los servicios externos existentes
- **Contenido**: Adaptadores que implementan las interfaces del dominio
- **Ventajas**: Permite migración gradual, mantiene compatibilidad

## 🚀 Plan de Migración Gradual

### **Paso 1: Crear Nueva Estructura (✅ COMPLETADO)**
- ✅ Definir modelos del dominio
- ✅ Crear interfaces del dominio
- ✅ Implementar servicios de aplicación
- ✅ Crear adaptadores para servicios existentes

### **Paso 2: Migrar Handler (PRÓXIMO)**
```go
// Antes (en handler.go)
func (h *Handler) AddTradeLink(url string, description string) {
    h.svc.AddTradeLink(url, description)
}

// Después (en handler.go actualizado)
func (h *Handler) AddTradeLink(url string, description string) {
    ctx := context.Background()
    err := h.tradeLinkAppSvc.AddTradeLink(ctx, url, description)
    if err != nil {
        // manejo de errores
    }
}
```

### **Paso 3: Mover Lógica Gradualmente**
- Mover lógica de `service.go` a `application/` paso a paso
- Mantener tests en cada cambio
- Actualizar handler para usar nuevos servicios

### **Paso 4: Refactor Repository (FUTURO)**
- Crear implementaciones concretas en `infrastructure/`
- Migrar desde adaptadores a implementaciones nativas

## 🔧 Cómo Usar la Nueva Arquitectura

### **Ejemplo 1: Añadir Trade Link**
```go
// En el handler actualizado
func (h *Handler) AddTradeLink(url string, description string) error {
    ctx := context.Background()
    return h.tradeLinkAppSvc.AddTradeLink(ctx, url, description)
}

// El servicio de aplicación se encarga de:
// 1. Validar datos de entrada
// 2. Crear entidad del dominio
// 3. Llamar al repositorio
// 4. Logging de la operación
```

### **Ejemplo 2: Testing Simplificado**
```go
func TestAddTradeLink(t *testing.T) {
    // Crear mocks de las interfaces del dominio
    mockRepo := &MockTradeLinkRepository{}
    mockLogger := &MockLogger{}

    // Crear servicio con dependencias mockeadas
    svc := NewTradeLinkApplicationService(mockRepo, mockLogger)

    // Test del caso de uso sin dependencias externas
    err := svc.AddTradeLink(ctx, "url", "desc")
    assert.NoError(t, err)
}
```

## 📊 Métricas de Mejora Esperadas

### **Antes**
- `service.go`: 908 líneas, múltiples responsabilidades
- Testing difícil por acoplamiento
- Cambios en una funcionalidad afectan otras

### **Después**
- Servicios de aplicación: ~100-200 líneas cada uno
- Testing independiente de infraestructura
- Cambios aislados por dominio

## 🛠️ Herramientas y Patterns Aplicados

### **Dependency Injection**
```go
// Constructor con dependencias explícitas
func NewTradeLinkApplicationService(
    repo domain.TradeLinkRepository,
    logger domain.Logger,
) *TradeLinkApplicationService
```

### **Repository Pattern**
```go
// Interface del dominio
type TradeLinkRepository interface {
    GetActiveTradeLinks(ctx context.Context) ([]TradeLink, error)
    Create(ctx context.Context, tradeLink *TradeLink) error
}
```

### **Adapter Pattern**
```go
// Adapta servicios existentes a interfaces del dominio
type LoggerAdapter struct {
    loggingSvc *logging.Service
}

func (a *LoggerAdapter) Info(module, message string, metadata map[string]interface{}) error {
    return a.loggingSvc.Log(logging.LogModuleLiveSearch, logging.LogLevelInfo, message, metadata)
}
```

## ✅ Próximos Pasos

1. **Actualizar Handler** para usar nuevos servicios de aplicación
2. **Crear tests unitarios** para servicios de aplicación
3. **Mover lógica compleja** desde `service.go` a servicios específicos
4. **Implementar adaptadores restantes** (WebSocket, EventBus)
5. **Documentar convenciones** para nuevas features

¿Te parece bien este enfoque? ¿Quieres que continúe con la implementación del Paso 2?
