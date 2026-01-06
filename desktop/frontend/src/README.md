# Walgo Frontend - Refactored Structure

## Architecture Overview

The Walgo frontend is now organized using a professional, modular architecture following React best practices and Clean Architecture principles.

## Project Structure

```
src/
├── components/          # React components
│   ├── ui/             # Reusable UI primitives
│   ├── layout/         # Layout-specific components
│   ├── file-tree/      # File tree components
│   └── modals/         # Modal dialogs
├── pages/              # Route/page-level components
├── hooks/              # Custom React hooks
├── types/              # TypeScript definitions
├── utils/              # Utilities and constants
├── services/           # API service layer (future)
├── assets/             # Static assets
└── App.tsx            # Main application component
```

## Component Organization

### UI Components (`components/ui/`)

Reusable, presentational components that can be used throughout the application.

**Examples:**

- `Card`: Content container with hover effects and borders
- `WindowControls`: Window management buttons (minimize, maximize, close)
- `StatusBanner`: Status notifications (success, error, info)

### Layout Components (`components/layout/`)

Components that define the application layout structure.

**Examples:**

- `NavItem`: Navigation sidebar items with active states

### File Tree Components (`components/file-tree/`)

Components specific to file management functionality.

**Examples:**

- `TreeNode`: Recursive tree node for displaying file structures

### Modal Components (`components/modals/`)

Modal dialogs for user interactions.

**Examples:**

- `WalletModal`: Wallet configuration modal
- `CreateAddressModal`: New address creation modal
- `DeleteConfirmModal`: Confirmation dialog

## Page Components (`pages/`)

Each page is a self-contained component that manages its own state and renders using shared components and hooks.

**Existing Pages:**

- `Dashboard`: System health overview and quick stats
- `AIConfig`: AI provider configuration
- `Projects`: Project listing and management

**Pages to Extract:**

- `Create`: Site creation workflow
- `QuickStart`: Quick start wizard
- `Edit`: Project editor with file tree
- `AIGenerate`: AI content generation
- `NewContent`: New content creation
- `Import`: Obsidian import
- `AICreateSite`: AI-powered site creation

## Custom Hooks (`hooks/`)

Custom hooks encapsulate stateful logic and side effects, making it reusable across components.

**Available Hooks:**

### `useProjects()`

Manages project state and operations.

```typescript
const { projects, loading, error, reloadProjects, setProjects } = useProjects();
```

### `useAIConfig()`

Manages AI configuration state.

```typescript
const {
  configured,
  config,
  provider,
  apiKey,
  baseUrl,
  model,
  setProvider,
  setApiKey,
  setBaseUrl,
  setModel,
  loading,
  updateConfig,
  cleanConfig,
  cleanProvider,
} = useAIConfig();
```

### `useWallet()`

Manages wallet operations.

```typescript
const {
  walletInfo,
  addressList,
  loading,
  reload,
  switchAddress,
  createAddress,
  importAddress,
  switchNetwork,
} = useWallet();
```

### `useSystemHealth()`

Monitors system health and dependencies.

```typescript
const { health, version, loading, checkDeps, reload } = useSystemHealth();
```

### `useEditProject()`

Manages project editing and file operations.

```typescript
const {
  project,
  files,
  currentPath,
  selectedFile,
  fileContent,
  loading,
  saving,
  saveFile,
  createFile,
  createDirectory,
  deleteFile,
  toggleFolder,
  reset,
} = useEditProject();
```

## Type Definitions (`types/`)

Centralized TypeScript interfaces ensure type safety across the application.

**Key Types:**

- `Status`: Status notification type
- `Project`: Project data structure
- `AIConfig`: AI configuration
- `WalletInfo`: Wallet information
- `SystemHealth`: System health status
- `FileTreeNode`: File tree node structure
- `LaunchConfig`: Launch configuration
- And more...

## Utilities (`utils/`)

### `helpers.ts`

Common utility functions:

- `cn()`: Merge Tailwind CSS classes
- `formatFileSize()`: Format file sizes
- `truncateText()`: Truncate text with ellipsis
- `formatRelativeTime()`: Format relative timestamps
- `debounce()`: Debounce function
- `generateId()`: Generate unique IDs

### `constants.ts`

Application-wide constants:

- Animation variants
- Site types
- AI providers
- Networks
- Project categories
- Configuration URLs

## Design Patterns

### Component Composition

Components are composed from smaller, reusable pieces. For example, the `Projects` page uses:

- `Card` for project cards
- State from `useProjects` hook
- Utility functions from `utils/helpers`

### Separation of Concerns

- **Components**: Render UI, handle user interactions
- **Hooks**: Manage state and side effects
- **Utils**: Pure functions for data manipulation
- **Types**: Define data structures

### Prop Drilling vs Context

Currently using prop drilling for simplicity. For deeper component trees, consider using React Context for:

- User session
- Theme
- Global settings

### Error Handling

All hooks include error handling and return error states to components for display.

## State Management

### Local State

Component-specific state using `useState`:

```typescript
const [showModal, setShowModal] = useState(false);
```

### Custom Hooks

Shared logic across components:

```typescript
const { projects, reloadProjects } = useProjects();
```

### Future: Context API

For global state like user authentication, theme, etc.

## Styling

### Tailwind CSS

Primary styling approach using utility classes.

### Helper Function

`cn()` function merges classes and resolves conflicts:

```typescript
import { cn } from "../utils/helpers";

<div className={cn("base-class", isActive && "active-class", className)} />;
```

### Custom Classes

Application-specific classes defined in CSS:

- `glass-panel-tech`: Glass morphism effect
- `scanline`: Scanline animation
- `bg-noise`: Noise texture

## Best Practices

### 1. Component Naming

- Use PascalCase for components: `Card`, `NavItem`
- Use descriptive names: `TreeNode` not `Tree`

### 2. File Organization

- One component per file
- Export from index.ts for cleaner imports
- Group related functionality

### 3. Type Safety

- Always type props interfaces
- Use TypeScript for all components
- Avoid `any` types

### 4. Performance

- Memoize expensive computations with `useMemo`
- Use `useCallback` for callbacks passed to children
- Avoid unnecessary re-renders

### 5. Accessibility

- Use semantic HTML elements
- Add ARIA labels where needed
- Ensure keyboard navigation works

## Testing Strategy

### Unit Tests

Test components in isolation:

```typescript
describe("Card", () => {
  it("renders children", () => {
    render(<Card>Test</Card>);
    expect(screen.getByText("Test")).toBeInTheDocument();
  });
});
```

### Hook Tests

Test custom hooks:

```typescript
describe("useProjects", () => {
  it("loads projects", async () => {
    const { result } = renderHook(() => useProjects());
    await waitFor(() => expect(result.current.projects).toHaveLength(3));
  });
});
```

### Integration Tests

Test user flows across components.

## Future Improvements

1. **Routing**: Consider React Router for navigation
2. **State Management**: Redux/Zustand for complex state
3. **Forms**: React Hook Form for form validation
4. **Testing**: Jest + React Testing Library
5. **Performance**: Code splitting, lazy loading
6. **Accessibility**: Full WCAG compliance
7. **i18n**: Internationalization support
8. **Theme**: Dark/light mode toggle

## Contributing

When adding new features:

1. Create type definitions in `types/`
2. Create hooks in `hooks/` for reusable logic
3. Create components in appropriate `components/` subdirectory
4. Create page components in `pages/`
5. Use existing patterns and utilities
6. Write tests for new code
7. Update documentation

## Dependencies

Key dependencies:

- React: UI framework
- Framer Motion: Animations
- Lucide React: Icons
- Tailwind CSS: Styling
- clsx & tailwind-merge: Class name utilities
- Wailsjs: Desktop runtime

## Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Run tests
npm test
```

## License

Part of the Walgo project.
