# Walgo Desktop

## About

Walgo Desktop is a GUI application for managing Hugo sites and deploying them to Walrus decentralized storage. It provides a modern, user-friendly interface for all Walgo CLI functionality.

## Features

Walgo Desktop supports all Walgo CLI commands:

### Site Management
- **Quickstart**: Create, build, and optionally deploy a new Hugo site in one flow
- **Create Site**: Initialize a new Hugo site with default configuration
- **New Content**: Create new blog posts or pages with automatic content type detection
- **Build**: Compile Hugo site with optimization

### Deployment
- **Launch Wizard**: Interactive deployment wizard (recommended) - guides through network selection, wallet setup, project naming, and deployment
- **Deploy**: Direct deployment to Walrus with epoch configuration
- **Update**: Update an existing deployed site with new content
- **Serve**: Start local Hugo development server with live reload

### Project Management
- **Projects List**: View all deployed projects with status and deployment history
- **Project Details**: Show comprehensive project information and statistics
- **Edit Metadata**: Update project name, description, category, image URL, and SuiNS domain

### AI Features
- **AI Configuration**: Configure OpenAI or OpenRouter API credentials
- **AI Generate**: Create new content using AI-powered generation
- **AI Update**: Modify existing content with AI
- **AI Create**: Generate complete Hugo sites with AI

### Import
- **Obsidian Import**: Import markdown files from Obsidian vaults to Hugo content

### Diagnostics
- **Doctor**: Run environment diagnostics to check for:
  - Hugo installation (required)
  - site-builder availability (for on-chain deployment)
  - walrus CLI presence (optional)
  - Sui client configuration
  - Wallet balance
  - Configuration files

### Optimization
- **Optimize**: Minify HTML, CSS, and JavaScript files
- **Compress**: Create Brotli compressed versions of files

## Development

### Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes.

### Building

To build a redistributable, production mode package, use `wails build`.

## Architecture

### Backend (Go)
- **Location**: `app.go`
- **API Layer**: `pkg/api/api.go` - Contains all business logic and CLI command wrappers
- **Clean Architecture**: Separation of domain, application, and infrastructure layers

### Frontend (React + TypeScript)
- **Framework**: React with TypeScript
- **Styling**: Tailwind CSS with custom tech theme
- **Animations**: Framer Motion
- **Icons**: Lucide React
- **Build Tool**: Vite

### Key Files

- `app.go` - Main application struct with exposed methods
- `pkg/api/api.go` - API layer with all command implementations
- `frontend/src/App.tsx` - Main React component

## UI Structure

The desktop app is organized into tabs:

1. **Dashboard**: Overview and quick actions
2. **Create**: Site creation options (Quickstart, Create Site, New Content, AI Generate, Import)
3. **Manage**: Build, Deploy, Update, Serve, and Launch operations
4. **Projects**: List and manage deployed projects
5. **AI**: AI configuration and capabilities
6. **Doctor**: Environment diagnostics

## Tech Stack

- **Backend**: Go 1.22+ with Wails v2
- **Frontend**: React 18+ with TypeScript
- **UI**: Tailwind CSS + Framer Motion
- **Build**: Vite
- **Package**: Wails (Desktop app framework)

## Requirements

- Go 1.22 or later
- Node.js 18+ (for development)
- Hugo (for site operations)
- site-builder (for on-chain deployment)
- Sui client (for on-chain deployment)

## Documentation

For more information about Walgo:
- CLI Commands: [COMMANDS.md](../docs/COMMANDS.md)
- Quick Start: [QUICK_START.md](../QUICK_START.md)
- Build: [BUILD.md](../BUILD.md)
- Main README: [README.md](../README.md)

## License

See project root for license information.
