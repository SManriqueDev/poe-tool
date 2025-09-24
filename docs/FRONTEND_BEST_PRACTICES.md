# Mejores Prácticas para Frontend React/TypeScript

## 📁 Estructura de Carpetas Recomendada

```
frontend/src/
├── components/          # Componentes reutilizables
│   ├── ui/             # Shadcn UI components
│   ├── forms/          # Componentes de formularios
│   ├── tables/         # Componentes de tablas
│   └── common/         # Componentes comunes
├── pages/              # Páginas principales
│   ├── LiveSearch/     # Página de búsqueda en vivo
│   ├── Settings/       # Página de configuración
│   └── Logs/          # Página de logs
├── services/           # Servicios para comunicación con backend
│   ├── api/           # Wrappers para bindings de Wails
│   ├── types/         # Tipos TypeScript compartidos
│   └── utils/         # Utilidades para servicios
├── hooks/              # Custom React hooks
├── contexts/           # React contexts
├── lib/               # Librerías y configuraciones
└── types/             # Tipos TypeScript globales
```

## 🔧 Patrón de Servicios Mejorado

### **Estructura Actual vs Mejorada**

#### **Antes (Actual)**
```typescript
// frontend/src/services/liveSearchService.ts
export async function addTradeLink(url: string, description: string): Promise<void> {
    return Handler.AddTradeLink(url, description);
}
```

#### **Después (Mejorado)**
```typescript
// frontend/src/services/api/livesearch.api.ts
import { Handler } from "../../bindings/...";
import type { TradeLink, AddTradeLinkRequest, ApiResponse } from "../types";

class LiveSearchApi {
    async addTradeLink(request: AddTradeLinkRequest): Promise<ApiResponse<void>> {
        try {
            await Handler.AddTradeLink(request.url, request.description);
            return { success: true, data: undefined };
        } catch (error) {
            return {
                success: false,
                error: error.message || "Failed to add trade link"
            };
        }
    }

    async listTradeLinks(): Promise<ApiResponse<TradeLink[]>> {
        try {
            const links = await Handler.ListTradeLinks();
            return { success: true, data: links };
        } catch (error) {
            return {
                success: false,
                error: error.message || "Failed to list trade links"
            };
        }
    }
}

export const liveSearchApi = new LiveSearchApi();

// frontend/src/services/livesearch.service.ts
import { liveSearchApi } from "./api/livesearch.api";
import type { TradeLink, AddTradeLinkRequest } from "./types";

export class LiveSearchService {
    async addTradeLink(url: string, description: string): Promise<void> {
        const result = await liveSearchApi.addTradeLink({ url, description });

        if (!result.success) {
            throw new Error(result.error);
        }
    }

    async getTradeLinks(): Promise<TradeLink[]> {
        const result = await liveSearchApi.listTradeLinks();

        if (!result.success) {
            throw new Error(result.error);
        }

        return result.data || [];
    }
}

export const liveSearchService = new LiveSearchService();
```

## 📊 Tipos TypeScript Organizados

```typescript
// frontend/src/services/types/common.ts
export interface ApiResponse<T> {
    success: boolean;
    data?: T;
    error?: string;
}

export interface PaginatedRequest {
    page: number;
    limit: number;
}

export interface PaginatedResponse<T> {
    items: T[];
    total: number;
    page: number;
    totalPages: number;
}

// frontend/src/services/types/livesearch.ts
export interface TradeLink {
    id: number;
    url: string;
    description: string;
    selected: boolean;
    createdAt: string;
}

export interface AddTradeLinkRequest {
    url: string;
    description: string;
}

export interface UpdateTradeLinkRequest {
    id: number;
    url: string;
    description: string;
    selected: boolean;
}

export interface LiveSearchSettings {
    goToHideoutEnabled: boolean;
    autoStartEnabled: boolean;
}

// frontend/src/services/types/index.ts
export * from './common';
export * from './livesearch';
export * from './logging';
export * from './settings';
```

## 🎣 Custom Hooks para Lógica Compleja

```typescript
// frontend/src/hooks/useLiveSearch.ts
import { useState, useEffect, useCallback } from "react";
import { Events } from "@wailsio/runtime";
import { liveSearchService } from "../services";
import type { TradeLink } from "../services/types";

export function useLiveSearch() {
    const [tradeLinks, setTradeLinks] = useState<TradeLink[]>([]);
    const [isRunning, setIsRunning] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // Cargar trade links
    const loadTradeLinks = useCallback(async () => {
        try {
            setLoading(true);
            setError(null);
            const links = await liveSearchService.getTradeLinks();
            setTradeLinks(links);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Unknown error");
        } finally {
            setLoading(false);
        }
    }, []);

    // Añadir trade link
    const addTradeLink = useCallback(async (url: string, description: string) => {
        try {
            await liveSearchService.addTradeLink(url, description);
            await loadTradeLinks(); // Recargar lista
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to add trade link");
        }
    }, [loadTradeLinks]);

    // Escuchar eventos en tiempo real
    useEffect(() => {
        const unsubscribe = Events.On("livesearch:statusChanged", (event) => {
            setIsRunning(event.data.isRunning);
        });

        return unsubscribe;
    }, []);

    // Cargar datos iniciales
    useEffect(() => {
        loadTradeLinks();
    }, [loadTradeLinks]);

    return {
        tradeLinks,
        isRunning,
        loading,
        error,
        addTradeLink,
        loadTradeLinks,
        clearError: () => setError(null),
    };
}

// Usage en componente
export function LiveSearchPage() {
    const {
        tradeLinks,
        isRunning,
        loading,
        error,
        addTradeLink
    } = useLiveSearch();

    if (loading) return <LoadingSpinner />;
    if (error) return <ErrorDisplay error={error} />;

    return (
        <div>
            <TradeLinksTable tradeLinks={tradeLinks} />
            <AddTradeLinkForm onAdd={addTradeLink} />
        </div>
    );
}
```

## 🧩 Componentes Reutilizables

```typescript
// frontend/src/components/forms/AddTradeLinkForm.tsx
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface AddTradeLinkFormProps {
    onAdd: (url: string, description: string) => Promise<void>;
    loading?: boolean;
}

export function AddTradeLinkForm({ onAdd, loading = false }: AddTradeLinkFormProps) {
    const [url, setUrl] = useState("");
    const [description, setDescription] = useState("");
    const [submitting, setSubmitting] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!url.trim() || !description.trim()) return;

        try {
            setSubmitting(true);
            await onAdd(url.trim(), description.trim());
            setUrl("");
            setDescription("");
        } finally {
            setSubmitting(false);
        }
    };

    const isDisabled = loading || submitting || !url.trim() || !description.trim();

    return (
        <Card>
            <CardHeader>
                <CardTitle>Add Trade Link</CardTitle>
            </CardHeader>
            <CardContent>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <Input
                        placeholder="Trade search URL"
                        value={url}
                        onChange={(e) => setUrl(e.target.value)}
                        disabled={loading || submitting}
                    />
                    <Input
                        placeholder="Description"
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        disabled={loading || submitting}
                    />
                    <Button
                        type="submit"
                        disabled={isDisabled}
                        className="w-full"
                    >
                        {submitting ? "Adding..." : "Add Trade Link"}
                    </Button>
                </form>
            </CardContent>
        </Card>
    );
}

// frontend/src/components/tables/TradeLinksTable.tsx
import { useState } from "react";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import type { TradeLink } from "@/services/types";

interface TradeLinksTableProps {
    tradeLinks: TradeLink[];
    onEdit?: (tradeLink: TradeLink) => void;
    onDelete?: (id: number) => void;
    onToggle?: (id: number, selected: boolean) => void;
}

export function TradeLinksTable({
    tradeLinks,
    onEdit,
    onDelete,
    onToggle
}: TradeLinksTableProps) {
    return (
        <Table>
            <TableHeader>
                <TableRow>
                    <TableHead>Status</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>URL</TableHead>
                    <TableHead>Actions</TableHead>
                </TableRow>
            </TableHeader>
            <TableBody>
                {tradeLinks.map((link) => (
                    <TableRow key={link.id}>
                        <TableCell>
                            <Badge variant={link.selected ? "default" : "secondary"}>
                                {link.selected ? "Active" : "Inactive"}
                            </Badge>
                        </TableCell>
                        <TableCell>{link.description}</TableCell>
                        <TableCell className="font-mono text-sm">
                            {link.url.substring(0, 50)}...
                        </TableCell>
                        <TableCell>
                            <div className="flex gap-2">
                                {onToggle && (
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={() => onToggle(link.id, !link.selected)}
                                    >
                                        {link.selected ? "Disable" : "Enable"}
                                    </Button>
                                )}
                                {onEdit && (
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={() => onEdit(link)}
                                    >
                                        Edit
                                    </Button>
                                )}
                                {onDelete && (
                                    <Button
                                        variant="destructive"
                                        size="sm"
                                        onClick={() => onDelete(link.id)}
                                    >
                                        Delete
                                    </Button>
                                )}
                            </div>
                        </TableCell>
                    </TableRow>
                ))}
            </TableBody>
        </Table>
    );
}
```

## 🚀 Convenciones de Desarrollo

### **Nombres de Archivos**
- Componentes: `PascalCase.tsx` (ej: `AddTradeLinkForm.tsx`)
- Hooks: `camelCase.ts` con prefijo `use` (ej: `useLiveSearch.ts`)
- Servicios: `camelCase.service.ts` (ej: `liveSearch.service.ts`)
- APIs: `camelCase.api.ts` (ej: `liveSearch.api.ts`)
- Tipos: `camelCase.ts` (ej: `livesearch.ts`, `common.ts`)

### **Exports**
```typescript
// Named exports para componentes y hooks
export { AddTradeLinkForm } from './AddTradeLinkForm';
export { useLiveSearch } from './useLiveSearch';

// Default export para servicios principales
export default class LiveSearchService { }

// Instance export para singletons
export const liveSearchService = new LiveSearchService();
```

### **Error Handling**
```typescript
// En servicios: devolver errores estructurados
async addTradeLink(url: string, description: string): Promise<void> {
    const result = await liveSearchApi.addTradeLink({ url, description });

    if (!result.success) {
        throw new LiveSearchError(result.error, 'ADD_TRADE_LINK_FAILED');
    }
}

// En componentes: usar try/catch con error states
const [error, setError] = useState<string | null>(null);

const handleAdd = async (url: string, description: string) => {
    try {
        setError(null);
        await liveSearchService.addTradeLink(url, description);
        toast.success("Trade link added successfully");
    } catch (err) {
        const message = err instanceof Error ? err.message : "Unknown error";
        setError(message);
        toast.error(message);
    }
};
```

## 📈 Beneficios de esta Estructura

### ✅ **Separación de Responsabilidades**
- **API Layer**: Solo comunicación con backend
- **Service Layer**: Lógica de negocio del frontend
- **Components**: Solo UI y interacciones
- **Hooks**: Lógica reutilizable

### ✅ **Mejor Testing**
```typescript
// Test de servicio (sin UI)
describe('LiveSearchService', () => {
    it('should add trade link', async () => {
        const mockApi = jest.mocked(liveSearchApi);
        mockApi.addTradeLink.mockResolvedValue({ success: true });

        await liveSearchService.addTradeLink('url', 'desc');

        expect(mockApi.addTradeLink).toHaveBeenCalledWith({
            url: 'url',
            description: 'desc'
        });
    });
});

// Test de hook (lógica de estado)
describe('useLiveSearch', () => {
    it('should load trade links on mount', async () => {
        const { result } = renderHook(() => useLiveSearch());

        await waitFor(() => {
            expect(result.current.tradeLinks).toHaveLength(2);
        });
    });
});
```

### ✅ **Mantenimiento Simplificado**
- Cambios en backend solo afectan API layer
- Lógica UI separada de lógica de negocio
- Componentes pequeños y enfocados
- Tipos TypeScript claramente definidos

¿Te parece bien esta estructura? ¿Quieres que continúe con más detalles o pase a la siguiente fase?
