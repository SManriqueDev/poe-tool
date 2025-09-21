# Poe Tool 🎮

**A powerful desktop application for Path of Exile trade automation** featuring real-time live search monitoring and automatic hideout teleportation for competitive item acquisition.

![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)
![Wails](https://img.shields.io/badge/Wails-v3.0.0--alpha.27-blue)
![Go](https://img.shields.io/badge/Go-1.23+-00ADD8)
![React](https://img.shields.io/badge/React-18.3.1-61DAFB)
![TypeScript](https://img.shields.io/badge/TypeScript-5.8.3-3178C6)

## ✨ Features

### 🔍 **Live Search Monitoring**
- **Real-time WebSocket connections** to Path of Exile trade API
- **Multiple concurrent searches** with individual status monitoring
- **Instant notifications** when new items matching your criteria appear
- **Comprehensive logging** of all found items with detailed information

### ⚡ **Competitive Hideout Automation**
- **Automatic teleportation** to seller hideouts when items are found
- **Ultra-fast processing** with 8-second realistic timing for trading
- **Smart duplicate prevention** to avoid redundant teleports
- **Priority-based system** ensuring you're always first to competitive items
- **Robust error handling** with temporary vs critical error classification

### 🎯 **Trade Management**
- **Multiple search support** - monitor up to 3+ searches simultaneously
- **Configurable search parameters** with easy enable/disable controls
- **Real-time status updates** for each search connection
- **Detailed activity logs** with timestamps and item details

### 🛡️ **Reliability Features**
- **Session management** with automatic PoE session validation
- **Connection recovery** for dropped WebSocket connections
- **Memory management** with automatic cleanup of processed items
- **Cross-platform compatibility** (Windows, macOS, Linux)

## 🚀 Quick Start

### Prerequisites
- **Path of Exile account** with valid session ID (POESESSID)
- **Active trade searches** on the official Path of Exile trade site
- **Windows 10+**, **macOS 10.13+**, or **Linux** (64-bit)

### Installation

1. **Download** the latest release for your platform:
   - Windows: `poe-tool.exe`
   - macOS: `poe-tool.app`
   - Linux: `poe-tool` (AppImage)

2. **Configure** your PoE session:
   - Open the app
   - Go to Settings
   - Enter your POESESSID from your browser cookies
   - Enable "Go to Hideout" for automatic teleportation

3. **Add your trade searches**:
   - Copy trade search URLs from pathofexile.com/trade2
   - Add them to the Live Search tab
   - Select which searches to monitor
   - Click "Start Live Search"

## 🎮 Usage Guide

### Setting Up Live Search

1. **Create searches** on [pathofexile.com/trade2](https://pathofexile.com/trade2)
2. **Copy the URLs** of your search results
3. **Add them to Poe Tool** with descriptive names
4. **Select which searches** you want to monitor actively
5. **Start Live Search** to begin real-time monitoring

### Hideout Automation

When enabled, the tool will:
- ✅ **Detect new items** from your monitored searches
- ✅ **Extract hideout tokens** from the trade API
- ✅ **Queue teleportation requests** with smart timing
- ✅ **Teleport you instantly** to the seller's hideout
- ✅ **Wait 8 seconds** before processing the next item (realistic trading time)

### Competitive Advantage

Perfect for high-value items where **speed matters**:
- **Mirror-tier items** - be first to contact the seller
- **League start economy** - grab underpriced items instantly
- **Bulk trading** - efficient hideout hopping for large purchases
- **Snipe expensive items** - automated speed gives you the edge

## 🔧 Development

### Tech Stack
- **Backend**: Go with Wails v3 framework
- **Frontend**: React 18 + TypeScript + Vite
- **UI**: Shadcn UI components + Tailwind CSS v4
- **Database**: SQLite for local data persistence
- **WebSockets**: Real-time communication with PoE API
- **Code Quality**: Biome for linting and formatting

### Development Setup

```bash
# Clone the repository
git clone https://github.com/SManriqueDev/poe-tool.git
cd poe-tool

# Install dependencies
make deps

# Run in development mode (hot reload)
make dev

# Development in browser (Go methods available)
# Open http://localhost:34115
```

### Building

```bash
# Build for your platform
make build-mac     # macOS universal binary
make build-windows # Windows x64
make build-linux   # Linux x64

# Cross-compilation
make build-windows-cross # Build Windows from macOS/Linux
```

### Project Structure

```
poe-tool/
├── backend/
│   ├── internal/
│   │   ├── livesearch/    # Live search + hideout automation
│   │   ├── settings/      # Configuration management
│   │   └── logging/       # Activity and error logging
│   └── db/                # SQLite database layer
├── frontend/
│   ├── src/
│   │   ├── pages/         # Main application pages
│   │   ├── components/    # Reusable UI components
│   │   ├── services/      # API communication layer
│   │   └── live-search/   # Live search specific components
│   └── wailsjs/           # Generated Go-TypeScript bindings
├── migrations/            # Database schema migrations
└── build/                 # Platform-specific build configs
```

## 🛠️ Available Commands

| Command | Description |
|---------|-------------|
| `make dev` | Run in development mode with hot reload |
| `make build-windows` | Build Windows executable |
| `make build-mac` | Build macOS universal app |
| `make build-linux` | Build Linux AppImage |
| `make deps` | Install frontend dependencies |
| `make lint` | Lint code with Biome |
| `make format` | Format code with Biome |
| `make check` | Auto-fix code issues |
| `make clean` | Clean build artifacts |

## ⚠️ Important Notes

### Security
- **Never share** your POESESSID - it provides full account access
- **Keep the app updated** for latest security patches
- **Use at your own risk** - automation tools exist in a gray area

### Performance
- **Recommended**: 3-5 concurrent searches maximum
- **Network**: Stable internet connection required for WebSockets
- **System**: Modern PC for smooth operation during intensive trading

### Legal Disclaimer
This tool interacts with the official Path of Exile trade API. Use responsibly and in accordance with Grinding Gear Games' Terms of Service.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues for bugs and feature requests.

### Development Workflow
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Grinding Gear Games** - for Path of Exile and the official trade API
- **Wails Community** - for the excellent desktop application framework
- **React & TypeScript** - for modern frontend development
- **Shadcn UI** - for beautiful, accessible components

---

**Built with ❤️ for the Path of Exile community**
