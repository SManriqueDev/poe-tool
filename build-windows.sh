#!/bin/bash

# Script para compilar el ejecutable de Windows de Poe Tool
# Uso: ./build-windows.sh [tipo]
# Tipos: simple, cgo, wails

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Función para mostrar ayuda
show_help() {
    echo -e "${BLUE}Poe Tool - Script de compilación para Windows${NC}"
    echo ""
    echo "Uso: $0 [tipo]"
    echo ""
    echo "Tipos disponibles:"
    echo -e "  ${GREEN}simple${NC}  - Compilación básica sin CGO (más pequeña, ~16MB)"
    echo -e "  ${GREEN}cgo${NC}     - Compilación con CGO habilitado (características completas, ~17MB)"
    echo -e "  ${GREEN}wails${NC}   - Compilación con Wails3 (aplicación de escritorio completa, ~35MB)"
    echo ""
    echo "Si no especificas un tipo, se compilarán todas las versiones."
    echo ""
    echo "Ejemplos:"
    echo "  $0 cgo      # Solo compilar con CGO"
    echo "  $0 wails    # Solo compilar con Wails3"
    echo "  $0          # Compilar todas las versiones"
}

# Función para compilar versión simple
build_simple() {
    echo -e "${YELLOW}Compilando versión simple (sin CGO)...${NC}"
    cd frontend && npm run build && cd ..
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o bin/poe-tool-simple.exe .
    echo -e "${GREEN}✓ Versión simple creada: bin/poe-tool-simple.exe${NC}"
}

# Función para compilar versión con CGO
build_cgo() {
    echo -e "${YELLOW}Compilando versión con CGO...${NC}"
    cd frontend && npm run build && cd ..
    CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -ldflags="-w -s" -o bin/poe-tool-cgo.exe .
    echo -e "${GREEN}✓ Versión con CGO creada: bin/poe-tool-cgo.exe${NC}"
}

# Función para compilar versión Wails3
build_wails() {
    echo -e "${YELLOW}Compilando versión con Wails3...${NC}"
    CC=x86_64-w64-mingw32-gcc wails3 task windows:build
    mv bin/poe-tool.exe bin/poe-tool-wails.exe 2>/dev/null || true
    echo -e "${GREEN}✓ Versión Wails3 creada: bin/poe-tool-wails.exe${NC}"
}

# Función para mostrar estadísticas de los archivos
show_stats() {
    echo ""
    echo -e "${BLUE}Estadísticas de los ejecutables:${NC}"
    echo "----------------------------------------"

    if [ -f "bin/poe-tool-simple.exe" ]; then
        size=$(ls -lh bin/poe-tool-simple.exe | awk '{print $5}')
        echo -e "Simple (sin CGO):    ${size}"
    fi

    if [ -f "bin/poe-tool-cgo.exe" ]; then
        size=$(ls -lh bin/poe-tool-cgo.exe | awk '{print $5}')
        echo -e "Con CGO:             ${size}"
    fi

    if [ -f "bin/poe-tool-wails.exe" ]; then
        size=$(ls -lh bin/poe-tool-wails.exe | awk '{print $5}')
        echo -e "Wails3 (completo):   ${size}"
    fi

    echo "----------------------------------------"
}

# Verificar dependencias
check_dependencies() {
    echo -e "${BLUE}Verificando dependencias...${NC}"

    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go no está instalado${NC}"
        exit 1
    fi

    if ! command -v npm &> /dev/null; then
        echo -e "${RED}Error: npm no está instalado${NC}"
        exit 1
    fi

    if ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
        echo -e "${RED}Error: mingw32-gcc no está instalado${NC}"
        echo "Instálalo con: brew install mingw-w64"
        exit 1
    fi

    if ! command -v wails3 &> /dev/null; then
        echo -e "${YELLOW}Advertencia: wails3 no está disponible, se omitirá la compilación con Wails3${NC}"
    fi

    echo -e "${GREEN}✓ Todas las dependencias están disponibles${NC}"
}

# Crear directorio bin si no existe
mkdir -p bin

# Main script logic
case "$1" in
    "help"|"-h"|"--help")
        show_help
        exit 0
        ;;
    "simple")
        check_dependencies
        build_simple
        show_stats
        ;;
    "cgo")
        check_dependencies
        build_cgo
        show_stats
        ;;
    "wails")
        check_dependencies
        if command -v wails3 &> /dev/null; then
            build_wails
        else
            echo -e "${RED}Error: wails3 no está disponible${NC}"
            exit 1
        fi
        show_stats
        ;;
    "")
        check_dependencies
        echo -e "${BLUE}Compilando todas las versiones...${NC}"
        echo ""

        build_simple
        echo ""

        build_cgo
        echo ""

        if command -v wails3 &> /dev/null; then
            build_wails
        else
            echo -e "${YELLOW}Omitiendo compilación con Wails3 (no disponible)${NC}"
        fi

        show_stats
        ;;
    *)
        echo -e "${RED}Error: Tipo desconocido '$1'${NC}"
        echo ""
        show_help
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}¡Compilación completada!${NC}"
