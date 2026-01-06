# Walgo Frontend Architecture

## High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         App.tsx                               │
│  (Main Application Orchestrator - Routing & Global State)      │
└─────────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Pages/     │     │  Components/ │     │   Hooks/    │
│              │     │              │     │              │
│  Dashboard   │────▶│  UI/         │     │ useProjects  │
│  AIConfig    │     │  - Card      │◀────│ useAIConfig  │
│  Projects    │     │  - WindowCtrl│     │ useWallet    │
│  Create      │     │  - Status    │     │ useSysHealth │
│  Edit        │     │  Layout/     │     │ useEditProj  │
│  QuickStart  │     │  - NavItem   │     │              │
│  AIGenerate  │     │  FileTree/   │     │              │
│  NewContent  │     │  - TreeNode  │     │              │
│  Import      │     │  Modals/     │     │              │
│  AICreateSite│     │              │     │              │
└──────────────┘     └──────────────┘     └──────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │     Types/       │
                    │  (Interfaces &   │
                    │   Type Safety)   │
                    └──────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │     Utils/       │
                    │  - helpers.ts    │
                    │  - constants.ts  │
                    └──────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │   Services/     │
                    │  (Future: API)  │
                    └──────────────────┘
```

## Data Flow Diagram

```
┌─────────────┐
│   User      │
│  Action     │
└──────┬──────┘
       │
       ▼
┌──────────────────────────────────────────────────────┐
│                Page Component                     │
│  (Dashboard, AIConfig, Projects, etc.)            │
└──────┬───────────────────────────────────────┬───┘
       │                                       │
       │ User Interaction                       │ State
       │                                       ▼
       │                              ┌──────────────┐
       │                              │  Custom Hook │
       │                              │             │
       │                              │ useProjects │
       ▼                              │ useAIConfig │
┌──────────────────┐                   │ useWallet   │
│  UI Component   │                   │             │
│  (Card, NavItem│                   └──────┬───────┘
│   TreeNode)     │                          │
└──────┬──────────┘                          │
       │                                     │
       │ Render                               │
       │                                     │
       ▼                                     │
┌───────────────────────────────────────────────┴──────┐
│                    Wails API                     │
│           (Go Backend Communication)                 │
│  - CreateSite, BuildSite, etc.        │
└──────────────────────────────────────────────────────┘
```

## Component Hierarchy

```
App.tsx
│
├─ WindowControls (UI)
│
├─ NavigationRail (Layout)
│  └─ NavItem (Layout)
│
├─ Routes (Pages)
│  ├─ Dashboard
│  │  ├─ Card (UI)
│  │  └─ StatusBanner (UI)
│  │
│  ├─ AIConfig
│  │  ├─ Card (UI)
│  │  ├─ StatusBanner (UI)
│  │  └─ useAIConfig hook
│  │
│  ├─ Projects
│  │  ├─ Card (UI) x N
│  │  └─ useProjects hook
│  │
│  ├─ Create
│  │  ├─ Card (UI)
│  │  └─ CreateSiteModal (Modals)
│  │
│  ├─ Edit
│  │  ├─ Card (UI)
│  │  ├─ TreeNode (FileTree) x N
│  │  └─ useEditProject hook
│  │
│  ├─ AIGenerate
│  │  ├─ Card (UI)
│  │  ├─ AIModal (Modals)
│  │  └─ useAIConfig hook
│  │
│  ├─ QuickStart
│  │  ├─ Card (UI)
│  │  └─ QuickStartModal (Modals)
│  │
│  ├─ Import
│  │  ├─ Card (UI)
│  │  └─ ImportModal (Modals)
│  │
│  └─ AICreateSite
│     ├─ Card (UI)
│     └─ AICreateSiteModal (Modals)
│
└─ Modals (Global)
   ├─ WalletModal
   ├─ AddressModal
   ├─ CreateAddressModal
   ├─ ImportAddressModal
   ├─ DeleteConfirmModal
   └─ LaunchModal
```

## State Management Strategy

### Local Component State

Used for:

- Modal open/close states
- Form input values
- Temporary UI state

```typescript
const [showModal, setShowModal] = useState(false);
const [formData, setFormData] = useState({ name: "" });
```

### Custom Hook State

Used for:

- Business logic
- API communication
- Shared data across components

```typescript
const { projects, loading, reloadProjects } = useProjects();
const { walletInfo, switchAddress } = useWallet();
```

### Global State (Future)

For:

- User authentication
- Theme preferences
- Application settings

```typescript
// Could use Context API or Zustand
const { user, login, logout } = useAuth();
```

## Type Safety Layers

```
Application Layer
       │
       │ Uses
       ▼
┌──────────────┐
│   Types/     │
│              │
│  - Project  │
│  - AIConfig │
│  - Wallet   │
│  - Health   │
│  - etc.     │
└──────────────┘
       │
       │ Provides type definitions
       ▼
Component & Hook Layer
```

## File System Map

```
frontend/src/
│
├─ 📁 components/           # Presentational components
│  ├─ 📁 ui/              # Generic UI primitives
│  │  ├─ Card.tsx          # Reusable card component
│  │  ├─ WindowControls.tsx # Window management
│  │  ├─ StatusBanner.tsx   # Status notifications
│  │  └─ index.ts          # Barrel export
│  │
│  ├─ 📁 layout/          # Layout-specific
│  │  ├─ NavItem.tsx      # Navigation item
│  │  └─ index.ts          # Barrel export
│  │
│  ├─ 📁 file-tree/       # File tree components
│  │  ├─ TreeNode.tsx      # Recursive tree node
│  │  └─ index.ts          # Barrel export
│  │
│  └─ 📁 modals/          # Modal dialogs
│     └─ index.ts          # Barrel export
│
├─ 📁 pages/              # Route components
│  ├─ Dashboard.tsx       # System overview
│  ├─ AIConfig.tsx        # AI configuration
│  ├─ Projects.tsx        # Project list
│  ├─ Create.tsx          # Site creation (to extract)
│  ├─ QuickStart.tsx      # Quick start (to extract)
│  ├─ Edit.tsx            # Project editor (to extract)
│  ├─ AIGenerate.tsx      # AI generation (to extract)
│  ├─ NewContent.tsx      # New content (to extract)
│  ├─ Import.tsx          # Obsidian import (to extract)
│  └─ index.ts            # Barrel export
│
├─ 📁 hooks/              # Custom React hooks
│  ├─ useProjects.ts      # Project management
│  ├─ useAIConfig.ts      # AI configuration
│  ├─ useWallet.ts        # Wallet operations
│  ├─ useSystemHealth.ts  # System health
│  ├─ useEditProject.ts   # Project editing
│  └─ index.ts           # Barrel export
│
├─ 📁 types/              # TypeScript definitions
│  └─ index.ts            # All interfaces
│
├─ 📁 utils/              # Utility functions
│  ├─ helpers.ts          # Helper functions
│  ├─ constants.ts        # App constants
│  └─ index.ts           # Barrel export
│
├─ 📁 services/           # API services (future)
│  └─ index.ts           # API abstraction
│
├─ 📁 assets/             # Static assets
│  ├─ fonts/             # Font files
│  └─ walgo-Wlogo-no_background.svg
│
├─ App.tsx               # Main app (to refactor)
├─ main.tsx              # Entry point
├─ index.css             # Global styles
└─ vite-env.d.ts         # Vite types
```

## Component Communication

### Parent to Child

```typescript
// Parent passes data via props
<Dashboard systemHealth={health} version={version} />
```

### Child to Parent

```typescript
// Child calls callback functions
<Card onClick={() => setActiveTab("quickstart")} />
```

### Via Custom Hooks

```typescript
// Both components access same state via hook
const { projects } = useProjects(); // In Projects page
const { projects } = useProjects(); // In Dashboard
```

### Via Props (Complex)

```typescript
// Pass multiple props to components
<EditProject
  project={selectedProject}
  onSave={handleSave}
  onCancel={handleCancel}
/>
```

## Reusability Examples

### Card Component

Used in:

- Dashboard (quick stats)
- Projects (project cards)
- AIConfig (configuration cards)
- Create site cards
- And many more...

### useProjects Hook

Used in:

- Projects page
- Dashboard (project count)
- Any component needing project data

### cn Helper Function

Used in:

- All components (className merging)
- Conditional styling
- Overriding base styles

## Code Organization Principles

### 1. Single Responsibility

Each component/hook has one clear purpose

- Card: Display content with styled container
- useProjects: Manage project state
- Dashboard: Show system overview

### 2. DRY (Don't Repeat Yourself)

Reusable components eliminate duplication

- Card component used 10+ times
- Shared logic in hooks
- Common utilities in utils/

### 3. Separation of Concerns

- Components: UI and user interactions
- Hooks: State and side effects
- Utils: Pure functions
- Types: Data structures

### 4. Composition Over Inheritance

Build complex UI from simple components

```typescript
<Card>
  <DashboardHeader />
  <StatusList />
  <Actions />
</Card>
```

### 5. Explicit Dependencies

Components receive what they need via props

- Clear data flow
- Easy to understand
- Better testability

## Future Enhancements

### Potential Architecture Improvements

1. **State Management**

   ```
   Currently: Hooks + useState
   Future: Zustand or Redux Toolkit
   ```

2. **Routing**

   ```
   Currently: Conditional rendering in App.tsx
   Future: React Router v6
   ```

3. **Form Handling**

   ```
   Currently: Manual state management
   Future: React Hook Form + Zod
   ```

4. **Data Fetching**

   ```
   Currently: Direct Wails calls in hooks
   Future: React Query or SWR
   ```

5. **Component Library**

   ```
   Currently: Custom components
   Future: shadcn/ui or similar
   ```

6. **Testing**
   ```
   Currently: No tests
   Future: Jest + React Testing Library
   ```

### Scalability Considerations

The current architecture scales well because:

1. **Modular Structure**: Easy to add new pages
2. **Reusable Components**: No code duplication
3. **Type Safety**: Catches errors at compile time
4. **Clear Patterns**: Easy to follow for new developers
5. **Separation**: Can optimize parts independently

## Summary

This architecture provides:

✅ **Clear Organization**: Easy to find and modify code
✅ **Reusability**: Components and hooks used multiple times
✅ **Type Safety**: TypeScript prevents common errors
✅ **Maintainability**: Changes isolated to specific files
✅ **Testability**: Components can be tested independently
✅ **Scalability**: Easy to add new features
✅ **Best Practices**: Follows React and TypeScript conventions

The structure is ready for both continued development and future enhancements.
