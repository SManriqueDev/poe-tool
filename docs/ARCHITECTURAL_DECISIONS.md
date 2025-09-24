# Guía de Decisiones Arquitecturales

## 🤔 ¿Cuándo crear nuevos paquetes/módulos?

### **Backend Go**

#### **✅ CREAR nuevo módulo cuando:**

1. **Nueva funcionalidad de dominio**
   ```
   Ejemplo: Sistema de notificaciones
   → backend/internal/notifications/
   ```

2. **Funcionalidad que crece >500 LOC en un solo archivo**
   ```
   Antes: livesearch/service.go (908 lines)
   Después:
   - livesearch/application/livesearch_service.go (150 lines)
   - livesearch/application/tradelink_service.go (100 lines)
   - livesearch/application/hideout_service.go (80 lines)
   ```

3. **Responsabilidades claramente separables**
   ```
   Ejemplo: Authentication independiente de LiveSearch
   → backend/internal/auth/
   ```

4. **Dependencias externas específicas**
   ```
   Ejemplo: Cliente para API de Path of Exile
   → backend/internal/poeapi/
   ```

#### **❌ NO crear nuevo módulo cuando:**

1. **Solo 1-2 funciones relacionadas**
   ```
   ❌ backend/internal/utils/stringhelper/
   ✅ backend/internal/shared/utils.go
   ```

2. **Funcionalidad muy acoplada al módulo principal**
   ```
   ❌ backend/internal/livesearch_helpers/
   ✅ backend/internal/livesearch/helpers.go
   ```

### **Frontend React**

#### **✅ CREAR nuevo directorio/módulo cuando:**

1. **Nueva página o feature completa**
   ```
   Ejemplo: Sistema de análisis de mercado
   → frontend/src/pages/MarketAnalysis/
   → frontend/src/services/marketAnalysis.service.ts
   ```

2. **Componentes reutilizables complejos (>200 LOC)**
   ```
   Ejemplo: Editor de filtros avanzado
   → frontend/src/components/FilterEditor/
   ```

3. **Lógica de estado compleja (custom hooks >100 LOC)**
   ```
   Ejemplo: Gestión de estado de WebSocket
   → frontend/src/hooks/useWebSocket.ts
   → frontend/src/hooks/websocket/ (si muy complejo)
   ```

## 🏗️ Patrones de Naming Conventions

### **Backend Go**

#### **Estructura de módulos:**
```
backend/internal/{domain}/
├── domain/              # Si usa Clean Architecture
│   ├── models.go       # Entidades del dominio
│   ├── interfaces.go   # Contratos
│   └── errors.go       # Errores específicos del dominio
├── application/         # Si usa Clean Architecture
│   └── {usecase}_service.go
├── adapters/           # Si usa Clean Architecture
│   └── {external}_adapter.go
├── service.go          # Servicio principal (patrón simple)
├── handler.go          # Handler para Wails bindings
├── repository.go       # Acceso a datos
├── model.go           # Modelos/structs
└── {specific}.go      # Archivos específicos (websocket_client.go)
```

#### **Nombres de archivos:**
```go
// Services
user_service.go          ❌ (snake_case)
userService.go          ❌ (camelCase)
service.go              ✅ (simple)
user.service.go         ✅ (si múltiples services)

// Interfaces
interfaces.go           ✅ (agrupadas en dominio)
user_repository.go      ✅ (específica)
repository.go           ✅ (simple)

// Models
models.go              ✅ (agrupados)
user.go               ✅ (específico)
```

#### **Nombres de structs y métodos:**
```go
// Structs
type UserService struct {}        ✅
type userService struct {}        ❌
type User_Service struct {}       ❌

// Métodos públicos (exportados)
func (s *Service) GetTradeLinks() ✅
func (s *Service) getTradeLinks() ❌ (privado sin razón)

// Métodos privados
func (s *Service) validateURL()   ✅
func (s *Service) ValidateURL()   ❌ (público sin necesidad)
```

### **Frontend TypeScript**

#### **Estructura de directorios:**
```
src/{domain}/
├── components/         # Componentes específicos del dominio
├── hooks/             # Hooks específicos del dominio
├── services/          # Servicios específicos del dominio
├── types/            # Tipos específicos del dominio
└── utils/            # Utilidades específicas del dominio

src/shared/            # Compartido entre dominios
├── components/
├── hooks/
├── services/
├── types/
└── utils/
```

#### **Nombres de archivos:**
```typescript
// Componentes
TradeLinksTable.tsx         ✅ (PascalCase)
trade-links-table.tsx       ❌ (kebab-case)
tradeLinkTable.tsx          ❌ (camelCase)

// Hooks
useLiveSearch.ts            ✅ (camelCase + use prefix)
use-live-search.ts          ❌ (kebab-case)
UseLiveSearch.ts            ❌ (PascalCase)

// Services
liveSearch.service.ts       ✅ (camelCase + .service)
LiveSearchService.ts        ❌ (PascalCase)
livesearch-service.ts       ❌ (kebab-case)

// Types
liveSearch.types.ts         ✅ (camelCase + .types)
types.ts                    ✅ (si están agrupados)
```

## 🎯 Cuándo Aplicar Clean Architecture vs Patrón Simple

### **Usar Clean Architecture cuando:**

✅ **Funcionalidad compleja** (>5 casos de uso diferentes)
✅ **Múltiples integraciones externas** (API + WebSocket + Database + Files)
✅ **Lógica de negocio compleja** que requiere testing aislado
✅ **Equipo de >3 desarrolladores** trabajando en el mismo módulo

**Ejemplo - LiveSearch es candidato:**
- Múltiples casos de uso (start, stop, manage links, process items)
- Múltiples integraciones (WebSocket, Database, Logging, Events)
- Lógica de negocio compleja (filtrado, procesado, colas)

### **Usar Patrón Simple cuando:**

✅ **Funcionalidad simple** (1-3 operaciones CRUD)
✅ **Una sola integración externa**
✅ **Lógica directa** sin reglas de negocio complejas
✅ **Una sola persona desarrollando**

**Ejemplo - Settings es candidato:**
- Operaciones simples (get, set, validate)
- Una integración (database)
- Lógica directa (key-value storage)

## 📊 Matriz de Decisión

| Criterio | Patrón Simple | Clean Architecture |
|----------|---------------|--------------------|
| Líneas de código | < 300 | > 500 |
| Casos de uso | 1-3 | > 5 |
| Integraciones externas | 1 | > 2 |
| Complejidad de testing | Baja | Alta |
| Reglas de negocio | Simples | Complejas |
| Desarrolladores | 1 | > 2 |

## 🚀 Plan de Migración Progresiva

### **Fase 1: Refactorizar módulos grandes**
```
✅ livesearch → Clean Architecture (domain, application, adapters)
⏳ logging → Mantener patrón simple (funciona bien)
⏳ settings → Mantener patrón simple (es simple)
```

### **Fase 2: Aplicar a nuevas funcionalidades**
```
⏳ notifications → Clean Architecture (si es complejo)
⏳ market-analysis → Clean Architecture (funcionalidad compleja)
⏳ user-preferences → Patrón simple (CRUD básico)
```

### **Fase 3: Estandarizar**
```
⏳ Crear templates/scaffolding para nuevos módulos
⏳ Documentar patrones establecidos
⏳ Code review guidelines
```

## 📝 Templates para Nuevos Módulos

### **Template: Patrón Simple**
```
backend/internal/{module}/
├── service.go          # Lógica principal
├── handler.go          # Wails bindings
├── repository.go       # Acceso a datos
├── model.go           # Structs/types
└── {module}_test.go   # Tests
```

### **Template: Clean Architecture**
```
backend/internal/{module}/
├── domain/
│   ├── models.go
│   ├── interfaces.go
│   └── errors.go
├── application/
│   ├── {usecase1}_service.go
│   └── {usecase2}_service.go
├── adapters/
│   ├── repository_adapter.go
│   └── {external}_adapter.go
├── handler.go          # Wails bindings
└── service.go          # Legacy (durante migración)
```

### **Template: Frontend Feature**
```
frontend/src/{feature}/
├── components/
│   ├── {Feature}Table.tsx
│   ├── Add{Feature}Form.tsx
│   └── index.ts
├── hooks/
│   └── use{Feature}.ts
├── services/
│   ├── {feature}.service.ts
│   └── api/{feature}.api.ts
├── types/
│   └── {feature}.types.ts
└── index.ts           # Export barrel
```

## ✅ Checklist para Nuevas Funcionalidades

### **Antes de crear código:**
- [ ] ¿Es funcionalidad nueva o extensión de existente?
- [ ] ¿Qué patrón arquitectural aplica? (Simple vs Clean)
- [ ] ¿Requiere nuevos módulos backend/frontend?
- [ ] ¿Cómo se integra con funcionalidades existentes?

### **Durante desarrollo:**
- [ ] Seguir naming conventions establecidas
- [ ] Crear interfaces para dependencias externas
- [ ] Aplicar dependency injection
- [ ] Escribir tests unitarios

### **Antes de PR:**
- [ ] Documentar decisiones arquitecturales
- [ ] Verificar que sigue patrones establecidos
- [ ] Tests passing
- [ ] Error handling implementado

¿Te parece útil esta guía? ¿Quieres que elabore más sobre algún punto específico?
