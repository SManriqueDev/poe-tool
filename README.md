# Poe Tool 🎮

**Poe Tool** es una aplicación de escritorio para *Path of Exile* diseñada para automatizar y facilitar el monitoreo de búsquedas de comercio en tiempo real, con soporte para teletransporte automático al hideout del vendedor. Está construida con **Go + Wails v3** en el backend y **React + TypeScript** en el frontend, siguiendo una arquitectura limpia por capas.

![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)
![Wails](https://img.shields.io/badge/Wails-v3-blue)
![Go](https://img.shields.io/badge/Go-1.23+-00ADD8)
![React](https://img.shields.io/badge/React-18-61DAFB)
![TypeScript](https://img.shields.io/badge/TypeScript-5.x-3178C6)

## Descripción

Poe Tool centraliza varias funciones pensadas para jugadores que realizan comercio de forma activa en Path of Exile:

- monitoreo en tiempo real de búsquedas de comercio;
- conexión continua mediante WebSockets;
- registro de actividad y resultados encontrados;
- gestión de configuraciones locales;
- automatización del acceso al hideout para acelerar el proceso de compra.

La aplicación está orientada a ofrecer una experiencia rápida, ordenada y multiplataforma.

## Características principales

### Monitoreo de Live Search
- Seguimiento en tiempo real de búsquedas activas.
- Soporte para múltiples búsquedas simultáneas.
- Estados visibles por cada búsqueda.
- Notificaciones cuando aparecen resultados nuevos.

### Automatización de Hideout
- Teletransporte automático al hideout del vendedor.
- Prevención de duplicados para evitar acciones repetidas.
- Gestión de tiempos entre acciones para un comportamiento más estable.
- Manejo de errores temporales y críticos.

### Gestión y confiabilidad
- Validación de sesión de Path of Exile.
- Recuperación de conexiones caídas.
- Limpieza automática de elementos procesados.
- Persistencia local con SQLite.

### Multiplataforma
- Compatible con Windows.
- Compatible con macOS.
- Compatible con Linux.

## Requisitos

Antes de usar la aplicación, asegúrate de cumplir con lo siguiente:

- Cuenta de Path of Exile con sesión válida.
- Token de sesión `POESESSID`.
- Búsquedas activas creadas en la web oficial de trade.
- Sistema operativo de 64 bits:
  - Windows 10 o superior
  - macOS 10.13 o superior
  - Linux

## Instalación

### Opción 1: Descargar una versión compilada
1. Descarga la última versión disponible para tu sistema operativo.
2. Ejecuta la aplicación:
   - Windows: `poe-tool.exe`
   - macOS: `poe-tool.app`
   - Linux: binario/AppImage correspondiente
3. Abre la configuración e introduce tu `POESESSID`.
4. Activa la opción de teletransporte al hideout si deseas usar la automatización.

### Opción 2: Ejecutar desde código fuente

```bash
git clone https://github.com/SManriqueDev/poe-tool.git
cd poe-tool
make deps
make dev
