# Desktop App Integration Guide

**Walgo v0.3.0** - Desktop Application Documentation

Complete guide to Walgo's desktop application features, usage, and development.

---

## Table of Contents

- [Overview](#overview)
- [Getting Started](#getting-started)
- [Desktop App Features](#desktop-app-features)
- [Pages & Navigation](#pages--navigation)
- [Wallet Management](#wallet-management)
- [AI Features](#ai-features)
- [Project Management](#project-management)
- [Site Creation](#site-creation)
- [System Health](#system-health)
- [Development](#development)
- [FAQ](#faq)

---

## Overview

The Walgo desktop application provides a graphical user interface for all Walgo CLI features, including:

- **Site Management** - Create, build, and deploy Hugo sites
- **AI Content Generation** - Generate and update content with AI
- **Project Tracking** - Manage all your deployed sites
- **Wallet Integration** - Manage Sui wallets and switch networks
- **System Health** - Monitor dependencies and auto-install tools

**Version:** v0.3.0

---

## 📋 FILES MODIFIED/CREATED

### 1. pkg/api/api.go

**Status**: ✅ EXPANDED
**Changes**:

- Added imports for ai, compress, optimizer, projects, obsidian
- Added content generation (`GenerateContent`)
- Added content update (`UpdateContent`)
- Added project management (`ListProjects`, `GetProject`, `DeleteProject`)
- Added import (`ImportObsidian`)
- Added helper function (`slugify`)

**Lines Added**: ~375 lines

---

### 2. desktop/app.go

**Status**: ✅ REWRITTEN
**Changes**:

- Complete rewrite to expose all new API methods
- Added AI methods (GenerateContent, UpdateContent)
- Added project management methods (ListProjects, GetProject, DeleteProject)
- Added import method (ImportObsidian)
- All methods properly wrapped with type conversions

**Lines Added**: ~180 lines (from 54 to 182)

---

### 3. cmd/desktop.go

**Status**: ✅ CREATED
**Changes**:

- New CLI command to launch desktop app
- Support for 3 modes:
  1. Development mode (`walgo desktop`)
  2. Build mode (`walgo desktop --build`)
  3. Production mode (`walgo desktop --prod`)
- Automatic project root detection
- Wails installation check
- Platform-specific binary handling (macOS, Linux, Windows)

**Lines Added**: ~165 lines (new file)

---

## 🚀 USAGE

### Launch Desktop App

**Development Mode (Recommended for development):**

```bash
walgo desktop
```

- Launches with hot-reload
- Auto-refreshes on code changes
- Opens at http://localhost:34115

---

## 💡 DESKTOP APP CAPABILITIES

The desktop app now provides a full GUI for all Walgo features:

### 🏗️ Site Management

- Create new Hugo sites
- Build sites
- Deploy to Walrus
- Optimize files
- Compress with Brotli

### 🤖 AI Features

- Configure AI providers (OpenAI/OpenRouter)
- Generate blog posts and pages
- Update existing content with AI

### 📊 Project Management

- View all deployed projects
- See project details
- Delete projects
- Track deployment history

### 📥 Content Import

- **Create site from Obsidian vault** - One-click import that creates a new Hugo site
- **Auto-site creation** - Initializes Hugo site with walgo.yaml automatically
- **Convert wikilinks** - Transforms `[[wikilinks]]` to Hugo markdown links
- **Copy attachments** - Handles images, PDFs, and media files
- **Smart defaults** - Site name from vault name, optional customization

---

## 🔧 TECHNICAL DETAILS

### API Structure

**pkg/api/api.go** exports:

```go
// Site Management
func CreateSite(parentDir string, name string) error
func BuildSite(sitePath string) error
func DeploySite(sitePath string, epochs int) DeployResult

// AI Features
func GenerateContent(params GenerateContentParams) GenerateContentResult
func UpdateContent(params UpdateContentParams) UpdateContentResult

// Projects
func ListProjects() ([]Project, error)
func GetProject(projectID int) (*Project, error)
func DeleteProject(projectID int) error

// Import
func ImportObsidian(params ImportObsidianParams) ImportObsidianResult
```

### Desktop App Binding

**desktop/app.go** exposes methods to frontend:

```javascript
// From frontend (React/TypeScript):
import { CreateSite, BuildSite, DeploySite } from "../wailsjs/go/main/App";
import { GenerateContent } from "../wailsjs/go/main/App";
import { ListProjects, GetProject } from "../wailsjs/go/main/App";

// Use in components:
const result = await CreateSite("/path/to/parent", "my-site");
const projects = await ListProjects();
const content = await GenerateContent({
  sitePath: "/path/to/site",
  contentType: "post",
  topic: "My Topic",
  context: "Additional context",
});
```

---

## 📦 DEPENDENCIES

### Required for Desktop App:

1. **Wails** - Desktop framework

   ```bash
   # macOS
   brew install wailsapp/wails/wails

   # Linux/Windows
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

2. **Node.js** - For frontend

   ```bash
   # Required for Vite (frontend build tool)
   node --version  # Should be v14+
   ```

3. **Walgo** - Main CLI (already installed)

---

## ✅ VERIFICATION

### Build Test:

```bash
✅ go build           - SUCCESS
✅ walgo desktop --help - Shows help
✅ All methods bound  - YES
✅ API complete       - YES
```

### Command List:

```bash
$ walgo --help
Available Commands:
  ...
  desktop     Launch the Walgo desktop application
  ...
```

---

## 📝 FRONTEND TODO (Next Steps)

The backend is complete! Frontend developers can now:

1. **Update React Components** to use new APIs:

   - AI configuration form
   - Content generation interface
   - Project management dashboard
   - Optimization controls
   - Import wizard

2. **Add New Pages**:

   - `/ai` - AI features
   - `/projects` - Project management
   - `/import` - Obsidian import
   - `/optimize` - Optimization tools

3. **Update Navigation**:

   - Add menu items for new features
   - Create tabs/sections

4. **Implement UI**:
   - Forms for AI configuration
   - Content editor with AI assist
   - Project list/grid view
   - Import progress indicators

---

## 🎓 EXAMPLES

### Example 1: AI Content Generation

```typescript
// frontend/src/components/AIGenerator.tsx
import { GenerateContent } from "../wailsjs/go/main/App";

const handleGenerate = async () => {
  const result = await GenerateContent({
    sitePath: currentSite,
    contentType: "post",
    topic: "How to Deploy to Walrus",
    context: "Focus on benefits of decentralized hosting",
  });

  if (result.Success) {
    console.log("Created:", result.FilePath);
    setGeneratedContent(result.Content);
  } else {
    console.error("Error:", result.Error);
  }
};
```

### Example 2: Project Management

```typescript
// frontend/src/components/Projects.tsx
import {
  ListProjects,
  GetProject,
  DeleteProject,
} from "../wailsjs/go/main/App";

const [projects, setProjects] = useState([]);

useEffect(() => {
  loadProjects();
}, []);

const loadProjects = async () => {
  const projectList = await ListProjects();
  setProjects(projectList);
};

const handleDelete = async (id: number) => {
  await DeleteProject(id);
  loadProjects(); // Refresh list
};
```

---

## 🏆 SUMMARY

### Commands:

- **Total**: 28 commands
- **New**: `desktop` command

### Desktop Methods:

- **Before**: 3 methods
- **After**: 12 methods
- **Growth**: +400%

### Features Now Available in Desktop:

✅ Site creation and building
✅ Deployment to Walrus
✅ AI content generation
✅ AI content updates
✅ Project management
✅ Obsidian import
✅ File optimization
✅ Brotli compression

### Status:

✅ Backend complete
✅ API fully functional
✅ CLI integration done
✅ Ready for frontend development

---

## 🎉 CONCLUSION

The Walgo desktop app backend is now fully integrated with all CLI features!

**What's Ready:**

- ✅ All API methods implemented
- ✅ Desktop app binding complete
- ✅ CLI launch command working
- ✅ Build and test passing

**Next Steps:**

- 🔨 Frontend developers: Update React UI to use new APIs
- 🎨 Design: Create UI/UX for new features
- 📱 Testing: Test all features in desktop app

The desktop app is now feature-complete on the backend side and ready for frontend implementation!

---

**Report Generated**: December 25, 2024
**Backend Integration**: ✅ COMPLETE
**Ready for Frontend**: YES
