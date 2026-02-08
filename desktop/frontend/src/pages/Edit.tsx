import React, { useState, useEffect, useRef, useMemo, useReducer } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  FolderOpen,
  MoreVertical,
  Folder,
  Sparkles,
  X,
  FileText,
  Save,
  Loader2,
  Play,
  StopCircle,
  Rocket,
  RefreshCw,
  Plus,
  FilePlus,
  FolderPlus,
  ChevronLeft,
  ChevronRight,
  AlertCircle,
  Search,
  Trash2,
  Info,
  Calendar,
  HardDrive,
  Files,
  Palette,
} from "lucide-react";
import { LoadingOverlay } from "../components/ui";
import { TreeNode, ContextMenu } from "../components/file-tree";
import { DeploymentModal, LaunchModal, AIGenerateModal, AIUpdateModal, InstallInstructionsModal, ThemeModal } from "../components/modals";
import type { DeploymentParams, DeploymentResult } from "../components/modals/DeploymentModal";
import type { LaunchConfig } from "../components/modals/LaunchModal";
import { buttonVariants } from "../utils/constants";
import { cn, renderMarkdown, formatFileSize } from "../utils";
import { useEditProject, useDependencyCheck } from "../hooks";
import { ListFiles } from "../../wailsjs/go/main/App";
import { SystemHealth, FileTreeNode } from "../types";

const SAVE_STATUS_DURATION = 5000; // ms - increased for better visibility

// --- Editor UI State Reducer ---
type ModalName = 'newItem' | 'launch' | 'deployment' | 'aiGenerate' | 'aiUpdate' | 'theme' | 'autoInstall' | 'delete' | 'projectActions';

interface EditorUIState {
  modals: Record<ModalName, boolean>;
  viewMode: 'split' | 'editor' | 'preview';
  sidebarCollapsed: boolean;
  showPropertiesPanel: boolean;
  isRootDragOver: boolean;
  newItemType: 'file' | 'folder';
  newItemName: string;
  contextMenuParentNode: FileTreeNode | null;
  fileToDelete: FileTreeNode | null;
  rootContextMenu: { x: number; y: number } | null;
}

type EditorUIAction =
  | { type: 'OPEN_MODAL'; modal: ModalName }
  | { type: 'CLOSE_MODAL'; modal: ModalName }
  | { type: 'TOGGLE_MODAL'; modal: ModalName }
  | { type: 'SET_VIEW_MODE'; mode: EditorUIState['viewMode'] }
  | { type: 'TOGGLE_SIDEBAR' }
  | { type: 'SET_PROPERTIES_PANEL'; value: boolean }
  | { type: 'SET_ROOT_DRAG_OVER'; value: boolean }
  | { type: 'SET_NEW_ITEM'; itemType: 'file' | 'folder'; parentNode?: FileTreeNode | null }
  | { type: 'SET_NEW_ITEM_NAME'; name: string }
  | { type: 'CLOSE_NEW_ITEM_MODAL' }
  | { type: 'SET_FILE_TO_DELETE'; file: FileTreeNode | null }
  | { type: 'SET_ROOT_CONTEXT_MENU'; position: { x: number; y: number } | null };

const initialUIState: EditorUIState = {
  modals: {
    newItem: false,
    launch: false,
    deployment: false,
    aiGenerate: false,
    aiUpdate: false,
    theme: false,
    autoInstall: false,
    delete: false,
    projectActions: false,
  },
  viewMode: 'split',
  sidebarCollapsed: false,
  showPropertiesPanel: false,
  isRootDragOver: false,
  newItemType: 'file',
  newItemName: '',
  contextMenuParentNode: null,
  fileToDelete: null,
  rootContextMenu: null,
};

function editorUIReducer(state: EditorUIState, action: EditorUIAction): EditorUIState {
  switch (action.type) {
    case 'OPEN_MODAL':
      return { ...state, modals: { ...state.modals, [action.modal]: true } };
    case 'CLOSE_MODAL':
      return { ...state, modals: { ...state.modals, [action.modal]: false } };
    case 'TOGGLE_MODAL':
      return { ...state, modals: { ...state.modals, [action.modal]: !state.modals[action.modal] } };
    case 'SET_VIEW_MODE':
      return { ...state, viewMode: action.mode };
    case 'TOGGLE_SIDEBAR':
      return { ...state, sidebarCollapsed: !state.sidebarCollapsed };
    case 'SET_PROPERTIES_PANEL':
      return { ...state, showPropertiesPanel: action.value };
    case 'SET_ROOT_DRAG_OVER':
      return { ...state, isRootDragOver: action.value };
    case 'SET_NEW_ITEM':
      return {
        ...state,
        modals: { ...state.modals, newItem: true },
        newItemType: action.itemType,
        contextMenuParentNode: action.parentNode ?? null,
      };
    case 'SET_NEW_ITEM_NAME':
      return { ...state, newItemName: action.name };
    case 'CLOSE_NEW_ITEM_MODAL':
      return {
        ...state,
        modals: { ...state.modals, newItem: false },
        newItemName: '',
        contextMenuParentNode: null,
      };
    case 'SET_FILE_TO_DELETE':
      return {
        ...state,
        fileToDelete: action.file,
        modals: { ...state.modals, delete: action.file !== null },
      };
    case 'SET_ROOT_CONTEXT_MENU':
      return { ...state, rootContextMenu: action.position };
    default:
      return state;
  }
}

interface EditProps {
  aiConfigured?: boolean;
  systemHealth?: SystemHealth;
  hasUpdates?: boolean;
  updatingTools?: string[];
  onStatusChange?: (status: { type: 'success' | 'error' | 'info'; message: string }) => void;
  onProjectUpdate?: () => Promise<void>;
  onRefreshHealth?: () => Promise<void>;
}

export const Edit: React.FC<EditProps> = ({
  aiConfigured = false,
  systemHealth,
  hasUpdates = false,
  updatingTools = [],
  onProjectUpdate,
  onStatusChange,
  onRefreshHealth,
}) => {
  // Internal state
  const [serverRunning, setServerRunning] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const {
    project,
    files,
    selectedFile,
    fileContent,
    loading,
    saving,
    error,
    expandedFolders,
    clipboard,
    setFileContent,
    setExpandedFolders,
    loadProject,
    selectFile,
    saveFile,
    deleteFile,
    renameFile,
    moveFile,
    copyToClipboard,
    cutToClipboard,
    pasteFromClipboard,
    toggleFolder,
    reset,
  } = useEditProject();

  // UI state managed by reducer (modals, view mode, sidebar, new item form, etc.)
  const [ui, dispatch] = useReducer(editorUIReducer, initialUIState);

  // Destructure reducer state for convenient access
  const { viewMode, sidebarCollapsed, showPropertiesPanel, isRootDragOver,
    newItemType, newItemName, contextMenuParentNode, fileToDelete, rootContextMenu } = ui;
  const showProjectActionsMenu = ui.modals.projectActions;
  const showNewItemModal = ui.modals.newItem;
  const showLaunchModal = ui.modals.launch;
  const showDeploymentModal = ui.modals.deployment;
  const showAIGenerateModal = ui.modals.aiGenerate;
  const showAIUpdateModal = ui.modals.aiUpdate;
  const showThemeModal = ui.modals.theme;
  const showAutoInstallModal = ui.modals.autoInstall;
  const showDeleteModal = ui.modals.delete;

  // Dispatch helpers matching old setter signatures
  const setShowProjectActionsMenu = (v: boolean) => dispatch(v ? { type: 'OPEN_MODAL', modal: 'projectActions' } : { type: 'CLOSE_MODAL', modal: 'projectActions' });
  const setShowLaunchModal = (v: boolean) => dispatch(v ? { type: 'OPEN_MODAL', modal: 'launch' } : { type: 'CLOSE_MODAL', modal: 'launch' });
  const setShowDeploymentModal = (v: boolean) => dispatch(v ? { type: 'OPEN_MODAL', modal: 'deployment' } : { type: 'CLOSE_MODAL', modal: 'deployment' });
  const setShowAIGenerateModal = (v: boolean) => dispatch(v ? { type: 'OPEN_MODAL', modal: 'aiGenerate' } : { type: 'CLOSE_MODAL', modal: 'aiGenerate' });
  const setShowAIUpdateModal = (v: boolean) => dispatch(v ? { type: 'OPEN_MODAL', modal: 'aiUpdate' } : { type: 'CLOSE_MODAL', modal: 'aiUpdate' });
  const setShowThemeModal = (v: boolean) => dispatch(v ? { type: 'OPEN_MODAL', modal: 'theme' } : { type: 'CLOSE_MODAL', modal: 'theme' });
  const setShowAutoInstallModal = (v: boolean) => dispatch(v ? { type: 'OPEN_MODAL', modal: 'autoInstall' } : { type: 'CLOSE_MODAL', modal: 'autoInstall' });
  const setShowDeleteModal = (v: boolean) => dispatch(v ? { type: 'OPEN_MODAL', modal: 'delete' } : { type: 'CLOSE_MODAL', modal: 'delete' });
  const setViewMode = (mode: EditorUIState['viewMode']) => dispatch({ type: 'SET_VIEW_MODE', mode });
  const setShowPropertiesPanel = (v: boolean | ((prev: boolean) => boolean)) => {
    const value = typeof v === 'function' ? v(ui.showPropertiesPanel) : v;
    dispatch({ type: 'SET_PROPERTIES_PANEL', value });
  };
  const setIsRootDragOver = (v: boolean) => dispatch({ type: 'SET_ROOT_DRAG_OVER', value: v });
  const setNewItemName = (name: string) => dispatch({ type: 'SET_NEW_ITEM_NAME', name });
  const setFileToDelete = (file: FileTreeNode | null) => dispatch({ type: 'SET_FILE_TO_DELETE', file });
  const setRootContextMenu = (pos: { x: number; y: number } | null) => dispatch({ type: 'SET_ROOT_CONTEXT_MENU', position: pos });

  // Auto-switch to editor-only mode for non-markdown files
  useEffect(() => {
    if (selectedFile && !selectedFile.isDir) {
      const isMarkdown = selectedFile.name.endsWith('.md') || selectedFile.name.endsWith('.markdown');
      if (!isMarkdown && viewMode !== 'editor') {
        setViewMode('editor');
      }
    }
  }, [selectedFile]);

  // Remaining independent states
  const [savingStatus, setSavingStatus] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const [isDeployed, setIsDeployed] = useState(false);
  const [launchConfig, setLaunchConfig] = useState<LaunchConfig | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const previousSearchQuery = useRef('');
  const [folderStats, setFolderStats] = useState<{ fileCount: number; folderCount: number; totalSize: number } | null>(null);
  const [loadingStats, setLoadingStats] = useState(false);
  const [cursorPosition, setCursorPosition] = useState({ line: 1, column: 1 });

  // Dependency check
  const depCheck = useDependencyCheck({
    systemHealth,
    hasUpdates,
    updatingTools
  });

  // Project path utility - compute once and reuse
  const projectPath = useMemo(() => project?.path || project?.sitePath, [project]);

  // Helper to close new item modal and reset state
  const closeNewItemModal = () => {
    dispatch({ type: 'CLOSE_NEW_ITEM_MODAL' });
  };

  // Load project from localStorage when component mounts or becomes visible
  useEffect(() => {
    const selectedProjectStr = localStorage.getItem('selectedProject');
    if (selectedProjectStr && !project) {
      try {
        const selectedProject = JSON.parse(selectedProjectStr);
        loadProject(selectedProject);
      } catch (err) {
        console.error('Failed to load selected project:', err);
      }
    }
  }, []);

  // Check if project is deployed when project changes
  useEffect(() => {
    if (project) {
      setIsDeployed(!!(project.objectId && project.objectId.length > 0));
    }
  }, [project]);

  const handleSave = async () => {
    setSavingStatus("Saving...");
    const result = await saveFile();
    if (result?.success) {
      setSavingStatus("Saved!");
      setTimeout(() => setSavingStatus(""), SAVE_STATUS_DURATION);
    } else {
      setSavingStatus("Failed");
      setTimeout(() => setSavingStatus(""), SAVE_STATUS_DURATION);
    }
  };

  const handleDeleteFile = async (file: FileTreeNode) => {
    setFileToDelete(file);
    setShowDeleteModal(true);
  };

  const confirmDelete = async () => {
    if (!fileToDelete) return;

    await deleteFile(fileToDelete.path);
    setShowDeleteModal(false);
    setFileToDelete(null);
  };

  const handleNewFileFromContext = (parentNode: FileTreeNode) => {
    dispatch({ type: 'SET_NEW_ITEM', itemType: 'file', parentNode });
  };

  const handleNewFolderFromContext = (parentNode: FileTreeNode) => {
    dispatch({ type: 'SET_NEW_ITEM', itemType: 'folder', parentNode });
  };

  const handleRename = async (node: FileTreeNode, newName: string) => {
    const result = await renameFile(node, newName);
    if (result.success && onStatusChange) {
      onStatusChange({
        type: 'success',
        message: `Renamed to: ${newName}`
      });
    } else if (!result.success && onStatusChange) {
      onStatusChange({
        type: 'error',
        message: result.error || 'Rename failed'
      });
    }
  };

  const handleMove = async (sourcePath: string, targetPath: string, expandTargetFolder?: string) => {
    // Check if the file being moved is currently open
    const isMovingOpenedFile = selectedFile && selectedFile.path === sourcePath;

    const result = await moveFile(sourcePath, targetPath);
    if (result.success) {
      const fileName = targetPath.substring(targetPath.lastIndexOf('/') + 1);
      const targetFolder = targetPath.substring(0, targetPath.lastIndexOf('/'));

      // Close the editor if the moved file was open
      if (isMovingOpenedFile) {
        selectFile(null as any);
        setFileContent('');
      }

      if (onStatusChange) {
        if (targetFolder === projectPath) {
          onStatusChange({
            type: 'success',
            message: `Moved "${fileName}" to project root`
          });
        } else {
          onStatusChange({
            type: 'success',
            message: `Moved successfully`
          });
        }
      }

      // If we need to expand a target folder, do it after files are reloaded
      if (expandTargetFolder) {
        setTimeout(() => {
          toggleFolder(expandTargetFolder);
        }, 500);
      }
    } else if (!result.success && onStatusChange) {
      onStatusChange({
        type: 'error',
        message: result.error || 'Move failed'
      });
    }
  };

  const handlePaste = async (targetFolder: FileTreeNode) => {
    const result = await pasteFromClipboard(targetFolder);
    if (result.success && onStatusChange) {
      const itemType = clipboard?.node.isDir ? 'Folder' : 'File';
      const operation = clipboard?.operation === 'copy' ? 'copied' : 'moved';
      onStatusChange({
        type: 'success',
        message: `${itemType} ${operation} successfully`
      });
    } else if (!result.success && onStatusChange) {
      onStatusChange({
        type: 'error',
        message: result.error || 'Paste failed'
      });
    }
  };

  const handleDuplicate = async (node: FileTreeNode) => {
    if (!node) return;

    try {
      const { CopyFile } = await import('../../wailsjs/go/main/App');

      // Get parent directory
      const parentPath = node.path.substring(0, node.path.lastIndexOf('/'));
      const targetPath = `${parentPath}/${node.name}`;

      // Backend will auto-increment the name
      const result = await CopyFile(node.path, targetPath);

      if (result.success) {
        // Reload project
        if (project) {
          await loadProject(project);
        }
        if (onStatusChange) {
          onStatusChange({
            type: 'success',
            message: `Duplicated: ${node.name}`
          });
        }
      } else {
        if (onStatusChange) {
          onStatusChange({
            type: 'error',
            message: result.error || 'Duplicate failed'
          });
        }
      }
    } catch (err) {
      if (onStatusChange) {
        onStatusChange({
          type: 'error',
          message: `Duplicate failed: ${err}`
        });
      }
    }
  };

  const checkDepth = async (node: FileTreeNode): Promise<boolean> => {
    if (!node || !node.isDir) return false;

    try {
      const { CheckDirectoryDepth } = await import('../../wailsjs/go/main/App');
      const result = await CheckDirectoryDepth(node.path);
      return result.tooDeep || false;
    } catch (err) {
      console.error('Failed to check depth:', err);
      return false;
    }
  };

  const handleRootDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsRootDragOver(true);
    e.dataTransfer.dropEffect = 'move';
  };

  const handleRootDragLeave = (e: React.DragEvent) => {
    e.stopPropagation();
    setIsRootDragOver(false);
  };

  // Recursive search through all files and folders
  const searchAllFiles = async (query: string) => {
    if (!projectPath) return;

    const lowerQuery = query.toLowerCase();
    const foldersToExpand = new Set<string>();

    // Load all children for a directory
    const loadAllChildren = async (dirPath: string): Promise<any[]> => {
      try {
        const result = await ListFiles(dirPath);
        if (!result || !result.files) return [];

        const children: FileTreeNode[] = [];
        for (const file of result.files) {
          if (file.isDir) {
            const subChildren = await loadAllChildren(file.path);
            children.push({
              ...file,
              children: subChildren
            });
          } else {
            children.push(file);
          }
        }

        // Sort: directories first, then files, both alphabetically
        return children.sort((a, b) => {
          if (a.isDir === b.isDir) {
            return a.name.localeCompare(b.name);
          }
          return a.isDir ? -1 : 1;
        });
      } catch (err) {
        console.error('Error loading children:', err);
        return [];
      }
    };

    // Recursively search through directory
    const searchDirectory = async (dirPath: string): Promise<any[]> => {
      try {
        const result = await ListFiles(dirPath);
        if (!result || !result.files) return [];

        const matches: FileTreeNode[] = [];

        for (const file of result.files) {
          const nameMatches = file.name.toLowerCase().includes(lowerQuery);

          if (file.isDir) {
            // Search inside the directory
            const childMatches = await searchDirectory(file.path);

            if (nameMatches) {
              // Folder name matches - include it with ALL its children
              foldersToExpand.add(file.path);
              const allChildren = await loadAllChildren(file.path);
              matches.push({
                ...file,
                children: allChildren
              });
            } else if (childMatches.length > 0) {
              // Folder doesn't match but has matching children - include as parent
              foldersToExpand.add(file.path);
              matches.push({
                ...file,
                children: childMatches
              });
            }
          } else {
            // It's a file
            if (nameMatches) {
              matches.push(file);
            }
          }
        }

        return matches;
      } catch (err) {
        console.error('Search error in', dirPath, err);
        return [];
      }
    };

    const searchResults = await searchDirectory(projectPath);
    return { results: searchResults, foldersToExpand };
  };

  const [searchResults, setSearchResults] = useState<FileTreeNode[]>([]);
  const [searchFolders, setSearchFolders] = useState<Set<string>>(new Set());
  const [isSearching, setIsSearching] = useState(false);

  // Debounced search when query changes
  useEffect(() => {
    if (!project) return;

    if (searchQuery.trim()) {
      setIsSearching(true);
      const timer = setTimeout(async () => {
        const result = await searchAllFiles(searchQuery);
        if (result) {
          setSearchResults(result.results);
          setSearchFolders(result.foldersToExpand);
        }
        setIsSearching(false);
      }, 300); // Debounce 300ms

      return () => clearTimeout(timer);
    } else {
      setSearchResults([]);
      setSearchFolders(new Set());
      setIsSearching(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchQuery, project?.path, project?.sitePath]);

  const filteredFiles = searchQuery.trim() ? searchResults : files;
  const foldersToExpand = searchQuery.trim() ? searchFolders : new Set<string>();

  // Auto-expand folders containing search results (only when search query actually changes)
  useEffect(() => {
    if (searchQuery !== previousSearchQuery.current) {
      previousSearchQuery.current = searchQuery;

      if (searchQuery && foldersToExpand.size > 0) {
        setExpandedFolders((prev: Set<string>) => {
          const newExpanded = new Set(prev);
          foldersToExpand.forEach(path => newExpanded.add(path));
          return newExpanded;
        });
      }
    }
  }, [searchQuery, foldersToExpand, setExpandedFolders]);

  // Load folder stats when properties panel is open and file is selected
  useEffect(() => {
    const loadFolderStats = async () => {
      if (!showPropertiesPanel || !selectedFile || !selectedFile.isDir) {
        setFolderStats(null);
        return;
      }

      setLoadingStats(true);
      try {
        const { GetFolderStats } = await import('../../wailsjs/go/main/App');
        const result = await GetFolderStats(selectedFile.path);
        if (result.success) {
          setFolderStats({
            fileCount: result.fileCount,
            folderCount: result.folderCount,
            totalSize: result.totalSize
          });
        }
      } catch (err) {
        console.error('Failed to load folder stats:', err);
      } finally {
        setLoadingStats(false);
      }
    };

    loadFolderStats();
  }, [showPropertiesPanel, selectedFile]);

  // Keyboard shortcuts handler
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Check if we're in an input or textarea
      const target = e.target as HTMLElement;
      const isInputField = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA';

      // Don't handle shortcuts when typing in input fields (except for Escape)
      if (isInputField && e.key !== 'Escape') return;

      const isMac = /mac/i.test(navigator.userAgent);
      const cmdOrCtrl = isMac ? e.metaKey : e.ctrlKey;

      // Ctrl/Cmd+C: Copy
      if (cmdOrCtrl && e.key === 'c' && selectedFile && !isInputField) {
        e.preventDefault();
        copyToClipboard(selectedFile);
        if (onStatusChange) {
          onStatusChange({
            type: 'info',
            message: `Copied: ${selectedFile.name}`
          });
        }
      }

      // Ctrl/Cmd+X: Cut
      if (cmdOrCtrl && e.key === 'x' && selectedFile && !isInputField) {
        e.preventDefault();
        cutToClipboard(selectedFile);
        if (onStatusChange) {
          onStatusChange({
            type: 'info',
            message: `Cut: ${selectedFile.name}`
          });
        }
      }

      // Ctrl/Cmd+V: Paste
      if (cmdOrCtrl && e.key === 'v' && clipboard && selectedFile?.isDir && !isInputField) {
        e.preventDefault();
        handlePaste(selectedFile);
      }

      // Ctrl/Cmd+D: Duplicate
      if (cmdOrCtrl && e.key === 'd' && selectedFile && !isInputField) {
        e.preventDefault();
        handleDuplicate(selectedFile);
      }

      // Delete key: Delete selected file
      if (e.key === 'Delete' && selectedFile && !isInputField) {
        e.preventDefault();
        setFileToDelete(selectedFile);
        setShowDeleteModal(true);
      }

      // F2: Rename selected file
      if (e.key === 'F2' && selectedFile && !isInputField) {
        e.preventDefault();
        // Trigger rename on selected file - we'll need to add a ref or state for this
        // For now, show a status message
        if (onStatusChange) {
          onStatusChange({
            type: 'info',
            message: 'Right-click the file and select Rename to rename it'
          });
        }
      }

      // Ctrl/Cmd+I: Toggle properties panel
      if (cmdOrCtrl && e.key === 'i' && !isInputField) {
        e.preventDefault();
        setShowPropertiesPanel(prev => !prev);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [selectedFile, clipboard, copyToClipboard, cutToClipboard, onStatusChange]);

  const handleRootDrop = async (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsRootDragOver(false);

    if (!projectPath) return;

    try {
      const data = JSON.parse(e.dataTransfer.getData('application/json'));
      const sourcePath = data.path;

      const targetPath = `${projectPath}/${data.name}`;

      // Don't move if already at root
      if (sourcePath === targetPath) {
        return;
      }

      // Check if source is already in project root (not nested)
      const sourceParent = sourcePath.substring(0, sourcePath.lastIndexOf('/'));
      if (sourceParent === projectPath) {
        if (onStatusChange) {
          onStatusChange({
            type: 'info',
            message: 'File is already in root folder'
          });
        }
        return;
      }

      await handleMove(sourcePath, targetPath);
    } catch (err) {
      console.error('Root drop failed:', err);
      if (onStatusChange) {
        onStatusChange({
          type: 'error',
          message: `Move failed: ${err}`
        });
      }
    }
  };


  const handleCloseProject = () => {
    reset();
    // Clear selected project from localStorage
    localStorage.removeItem('selectedProject');
  };

  const handleLaunchConfig = async (config: LaunchConfig) => {
    setLaunchConfig(config);
    setShowLaunchModal(false);
    setShowDeploymentModal(true);
  };

  // Server operations
  const handleServeToggle = async () => {
    if (!projectPath) {
      onStatusChange?.({
        type: 'error',
        message: 'No project loaded',
      });
      return;
    }

    setIsProcessing(true);

    try {
      if (serverRunning) {
        const { StopServe } = await import('../../wailsjs/go/main/App');
        await StopServe();
        setServerRunning(false);
        onStatusChange?.({ type: 'info', message: 'Server stopped' });
      } else {
        // Build first
        onStatusChange?.({ type: 'info', message: 'Building site...' });
        const { BuildSite } = await import('../../wailsjs/go/main/App');
        await BuildSite(projectPath);

        // Then serve
        onStatusChange?.({ type: 'info', message: 'Starting server...' });
        const { Serve } = await import('../../wailsjs/go/main/App');
        await Serve({
          sitePath: projectPath,
          port: 1313,
          drafts: true,
          expired: false,
          future: false,
        });
        setServerRunning(true);

        const url = 'http://localhost:1313';
        onStatusChange?.({
          type: 'success',
          message: `Server running: ${url}`,
        });

        // Open browser after Hugo has time to start listening
        const { BrowserOpenURL } = await import('../../wailsjs/runtime/runtime');
        setTimeout(() => {
          BrowserOpenURL(url);
        }, 2000);
      }
    } catch (err) {
      onStatusChange?.({
        type: 'error',
        message: `Server error: ${err?.toString() || 'Unknown error'}`,
      });
    } finally {
      setIsProcessing(false);
    }
  };

  // Open folder in finder/explorer
  const handleOpenFolder = async () => {
    const projectPath = project?.path || project?.sitePath;
    if (!projectPath) return;

    try {
      const { OpenInFinder } = await import('../../wailsjs/go/main/App');
      await OpenInFinder(projectPath);
    } catch (err) {
      onStatusChange?.({
        type: 'error',
        message: `Failed to open: ${err?.toString()}`,
      });
    }
  };

  const handleDeployment = async (params: DeploymentParams): Promise<DeploymentResult> => {
    if (!projectPath) {
      return {
        success: false,
        error: "Project path not found",
      };
    }

    try {
      // Check if this is an update (project already deployed)
      if (isDeployed && project?.id) {
        // Use UpdateSite for existing deployments
        const { UpdateSite } = await import('../../wailsjs/go/main/App');

        const result = await UpdateSite({
          projectId: project.id,
          epochs: params.epochs,
        });

        // Build logs from UpdateSite result
        const logs: string[] = result.logs || [];

        if (result.success && result.objectId) {
          // Refresh projects list
          if (onProjectUpdate) {
            await onProjectUpdate();
          }

          logs.push("");
          logs.push("═══════════════════════════════════════");
          logs.push(`📋 Site Object ID: ${result.objectId}`);
          if (result.gasFee) {
            logs.push(`💰 Gas Fee: ${result.gasFee}`);
          }
          logs.push("═══════════════════════════════════════");

          return {
            success: true,
            objectId: result.objectId,
            logs,
          };
        } else {
          logs.push("");
          logs.push("❌ Update failed");
          if (result.error) {
            logs.push(`Error: ${result.error}`);
          }

          return {
            success: false,
            error: result.error || "Update failed",
            logs,
          };
        }
      }

      // New deployment - use LaunchWizard
      const { LaunchWizard } = await import('../../wailsjs/go/main/App');

      const result = await LaunchWizard({
        sitePath: projectPath,
        network: params.network,
        projectName: launchConfig?.projectName || project?.name || "My Project",
        category: launchConfig?.category || "website",
        description: launchConfig?.description || "",
        imageUrl: launchConfig?.imageUrl || "",
        epochs: params.epochs,
        skipConfirm: true,
      });

      // Build comprehensive logs from steps
      const logs: string[] = [];

      if (result.steps && result.steps.length > 0) {
        logs.push("📊 Deployment Progress:");
        logs.push("");

        result.steps.forEach((step, idx) => {
          if (step.status === "success") {
            logs.push(`✓ [${idx + 1}/${result.steps.length}] ${step.name}`);
          } else if (step.status === "error") {
            logs.push(`✗ [${idx + 1}/${result.steps.length}] ${step.name}`);
          } else if (step.status === "running") {
            logs.push(`⏳ [${idx + 1}/${result.steps.length}] ${step.name}`);
          } else {
            logs.push(`  [${idx + 1}/${result.steps.length}] ${step.name}`);
          }

          if (step.message) {
            // Split multi-line messages
            const messages = step.message.split('\n');
            messages.forEach(msg => {
              if (msg.trim()) {
                logs.push(`    ${msg}`);
              }
            });
          }

          if (step.error) {
            logs.push(`    ⚠️  Error: ${step.error}`);
          }

          logs.push("");
        });
      }

      if (result.success && result.objectId) {
        // Update project state
        setIsDeployed(true);

        // Refresh projects list
        if (onProjectUpdate) {
          await onProjectUpdate();
        }

        logs.push("═══════════════════════════════════════");
        logs.push(`📋 Site Object ID: ${result.objectId}`);
        logs.push("═══════════════════════════════════════");
        logs.push("");
        logs.push("✅ Deployment completed successfully!");

        return {
          success: true,
          objectId: result.objectId,
          logs,
        };
      } else {
        logs.push("❌ Deployment failed");
        if (result.error) {
          logs.push(`Error: ${result.error}`);
        }

        return {
          success: false,
          error: result.error || "Deployment failed",
          logs,
        };
      }
    } catch (err) {
      return {
        success: false,
        error: err instanceof Error ? err.message : "Unknown error",
        logs: [`❌ Exception: ${err instanceof Error ? err.message : "Unknown error"}`],
      };
    }
  };

  const validateSlug = (slug: string): boolean => {
    // Allow letters, numbers, hyphens, underscores, dots, and forward slashes (for subdirectories)
    const validSlug = /^[a-zA-Z0-9_\-\.\/]+$/;
    return validSlug.test(slug) && slug.length > 0 && slug.length < 100;
  };



  const formatDate = (timestamp: number | undefined): string => {
    if (!timestamp) return 'Unknown';
    const date = new Date(timestamp * 1000);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
  };

  // Editor stats helpers
  const getLineCount = (text: string): number => {
    return text.split('\n').length;
  };

  const getWordCount = (text: string): number => {
    return text.trim().split(/\s+/).filter(word => word.length > 0).length;
  };

  const handleTextareaChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const textarea = e.target;
    const value = textarea.value;
    setFileContent(value);

    // Update cursor position
    const textBeforeCursor = value.substring(0, textarea.selectionStart);
    const lines = textBeforeCursor.split('\n');
    const lineNumber = lines.length;
    const columnNumber = lines[lines.length - 1].length + 1;

    setCursorPosition({ line: lineNumber, column: columnNumber });
  };

  const handleTextareaClick = (e: React.MouseEvent<HTMLTextAreaElement>) => {
    const textarea = e.target as HTMLTextAreaElement;
    const textBeforeCursor = textarea.value.substring(0, textarea.selectionStart);
    const lines = textBeforeCursor.split('\n');
    const lineNumber = lines.length;
    const columnNumber = lines[lines.length - 1].length + 1;

    setCursorPosition({ line: lineNumber, column: columnNumber });
  };

  const handleTextareaKeyUp = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    const textarea = e.target as HTMLTextAreaElement;
    const textBeforeCursor = textarea.value.substring(0, textarea.selectionStart);
    const lines = textBeforeCursor.split('\n');
    const lineNumber = lines.length;
    const columnNumber = lines[lines.length - 1].length + 1;

    setCursorPosition({ line: lineNumber, column: columnNumber });
  };

  const handleRootContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();

    setRootContextMenu({
      x: e.clientX,
      y: e.clientY,
    });
  };

  const handleCreateNewItem = async () => {
    if (!newItemName.trim()) {
      if (onStatusChange) {
        onStatusChange({
          type: 'error',
          message: 'Name required: Please enter a name'
        });
      }
      return;
    }

    // Validate slug
    if (!validateSlug(newItemName)) {
      if (onStatusChange) {
        onStatusChange({
          type: 'error',
          message: 'Invalid name: Use only letters, numbers, hyphens, underscores, and slashes'
        });
      }
      return;
    }

    if (!projectPath) {
      if (onStatusChange) {
        onStatusChange({
          type: 'error',
          message: 'Project path not found'
        });
      }
      return;
    }

    setIsCreating(true);
    try {
      const { CreateFile, CreateDirectory } = await import('../../wailsjs/go/main/App');

      // Determine base path (either from context menu parent or default to project root)
      let basePath = projectPath;
      if (contextMenuParentNode && contextMenuParentNode.isDir) {
        basePath = contextMenuParentNode.path;
      }

      if (newItemType === 'file') {
        // Use the filename as provided by the user
        const fileName = newItemName;
        const filePath = `${basePath}/${fileName}`;

        // Extract title from filename (remove path and extension)
        const title = fileName.split('/').pop()?.replace(/\.[^.]+$/, '') || 'New Content';

        // Check if it's a markdown file
        const isMarkdown = fileName.toLowerCase().endsWith('.md') || fileName.toLowerCase().endsWith('.markdown');

        // Create file with basic frontmatter only for markdown files
        let content: string;
        if (isMarkdown) {
          content = `---
title: "${title}"
date: ${new Date().toISOString()}
draft: false
---

# ${title}

Start writing your content here...
`;
        } else {
          // Empty content for non-markdown files
          content = '';
        }

        const result = await CreateFile(filePath, content);
        if (result.success && project) {
          await loadProject(project);
          closeNewItemModal();
          if (onStatusChange) {
            onStatusChange({
              type: 'success',
              message: `File Created: ${fileName}`
            });
          }
        } else {
          if (onStatusChange) {
            onStatusChange({
              type: 'error',
              message: `Create Failed: ${result.error}`
            });
          }
        }
      } else {
        // Create folder
        const folderPath = `${basePath}/${newItemName}`;
        const result = await CreateDirectory(folderPath);
        if (result.success && project) {
          await loadProject(project);
          closeNewItemModal();
          if (onStatusChange) {
            onStatusChange({
              type: 'success',
              message: `Folder Created: ${newItemName}`
            });
          }
        } else {
          if (onStatusChange) {
            onStatusChange({
              type: 'error',
              message: `Create Failed: ${result.error}`
            });
          }
        }
      }
    } catch (err) {
      console.error('Failed to create item:', err);
      if (onStatusChange) {
        onStatusChange({
          type: 'error',
          message: 'Create failed: Unknown error'
        });
      }
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <div className="flex gap-6 h-[calc(100vh-120px)] min-h-[600px] relative">
      {/* Left Panel - File Explorer with Smooth Collapse */}
      <div 
        className="relative flex-shrink-0 transition-all duration-300 ease-in-out group"
        style={{ 
          width: sidebarCollapsed ? '0px' : 'calc(33.333% - 12px)',
          maxWidth: sidebarCollapsed ? '0px' : '400px',
          minWidth: sidebarCollapsed ? '0px' : '280px',
          opacity: sidebarCollapsed ? 0 : 1,
          overflow: 'visible'
        }}
      >
        <div className="glass-panel-tech h-full flex flex-col relative overflow-hidden rounded-sm">
          {/* Card corner decorations */}
          <div className="scanline opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
          <div className="absolute top-0 left-0 w-2 h-2 border-t border-l border-white/20 z-20" />
          <div className="absolute top-0 right-0 w-2 h-2 border-t border-r border-white/20 z-20" />
          <div className="absolute bottom-0 left-0 w-2 h-2 border-b border-l border-white/20 z-20" />
          <div className="absolute bottom-0 right-0 w-2 h-2 border-b border-r border-white/20 z-20" />
          
          {/* Header - Fixed */}
          <div className="p-4 border-b border-white/5 bg-black/20 flex-shrink-0 relative z-10 min-h-[60px]">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <FolderOpen size={18} className="text-accent" />
                <span className="text-sm font-display text-white">
                  {project?.name || "No Project Selected"}
                </span>
              </div>
              {project && (
                <div className="relative">
                  <button
                    onClick={() =>
                      dispatch({ type: 'TOGGLE_MODAL', modal: 'projectActions' })
                    }
                    className="p-2 bg-white/5 hover:bg-white/10 text-zinc-400 hover:text-white border border-white/10 rounded-sm transition-all"
                  >
                    <MoreVertical size={16} />
                  </button>
                  {showProjectActionsMenu && (
                    <>
                      <div
                        className="fixed inset-0 z-10"
                        onClick={() => setShowProjectActionsMenu(false)}
                      />
                      <div className="absolute right-0 top-full mt-2 w-48 bg-zinc-900 border border-white/10 rounded-sm shadow-lg z-20 overflow-hidden">
                        <button
                          onClick={async () => {
                            if (projectPath) {
                              await loadProject(project);
                              if (onStatusChange) {
                                onStatusChange({
                                  type: 'success',
                                  message: 'Folder refreshed'
                                });
                              }
                            }
                            setShowProjectActionsMenu(false);
                          }}
                          className="w-full px-4 py-2.5 text-left text-xs font-mono text-zinc-300 hover:bg-white/5 hover:text-white transition-all flex items-center gap-3"
                        >
                          <RefreshCw size={14} className="text-zinc-500" />
                          Refresh Folder
                        </button>
                        <button
                          onClick={() => {
                            dispatch({ type: 'SET_NEW_ITEM', itemType: 'file' });
                            setShowProjectActionsMenu(false);
                          }}
                          className="w-full px-4 py-2.5 text-left text-xs font-mono text-zinc-300 hover:bg-white/5 hover:text-white transition-all flex items-center gap-3 border-t border-white/5"
                        >
                          <FilePlus size={14} className="text-zinc-500" />
                          New File
                        </button>
                        <button
                          onClick={() => {
                            dispatch({ type: 'SET_NEW_ITEM', itemType: 'folder' });
                            setShowProjectActionsMenu(false);
                          }}
                          className="w-full px-4 py-2.5 text-left text-xs font-mono text-zinc-300 hover:bg-white/5 hover:text-white transition-all flex items-center gap-3 border-t border-white/5"
                        >
                          <FolderPlus size={14} className="text-zinc-500" />
                          New Folder
                        </button>
                        <button
                          onClick={() => {
                            handleOpenFolder();
                            setShowProjectActionsMenu(false);
                          }}
                          className="w-full px-4 py-2.5 text-left text-xs font-mono text-zinc-300 hover:bg-white/5 hover:text-white transition-all flex items-center gap-3 border-t border-white/5"
                        >
                          <Folder size={14} className="text-zinc-500" />
                          Open Folder
                        </button>
                        <button
                          onClick={() => {
                            if (aiConfigured) {
                              setShowAIGenerateModal(true);
                            } else if (onStatusChange) {
                              onStatusChange({
                                type: 'error',
                                message: 'AI not configured: Please configure AI in Settings to use AI features.'
                              });
                            }
                            setShowProjectActionsMenu(false);
                          }}
                          className={cn(
                            "w-full px-4 py-2.5 text-left text-xs font-mono transition-all flex items-center gap-3 border-t border-white/5",
                            aiConfigured
                              ? "text-purple-400 hover:bg-purple-500/10"
                              : "text-zinc-600 hover:bg-zinc-800/50"
                          )}
                        >
                          <Sparkles size={14} />
                          AI Generate
                        </button>
                        <button
                          onClick={() => {
                            setShowThemeModal(true);
                            setShowProjectActionsMenu(false);
                          }}
                          className="w-full px-4 py-2.5 text-left text-xs font-mono text-pink-400 hover:bg-pink-500/10 transition-all flex items-center gap-3 border-t border-white/5"
                        >
                          <Palette size={14} />
                          Change Theme
                        </button>
                        <button
                          onClick={() => {
                            handleCloseProject();
                            setShowProjectActionsMenu(false);
                          }}
                          className="w-full px-4 py-2.5 text-left text-xs font-mono text-red-400 hover:bg-red-500/10 transition-all flex items-center gap-3 border-t border-white/5"
                        >
                          <X size={14} />
                          Close Project
                        </button>
                      </div>
                    </>
                  )}
                </div>
              )}
            </div>
          </div>

          {/* Search Bar */}
          {project && (
            <div className="px-3 py-2 border-b border-white/5">
              <div className="relative">
                <input
                  type="text"
                  placeholder="Search files..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  autoComplete="off"
                  autoCapitalize="off"
                  autoCorrect="off"
                  spellCheck={false}
                  className="w-full px-3 py-2 pl-9 pr-8 bg-zinc-800/50 border border-white/10 rounded text-xs font-mono text-white placeholder-zinc-500 focus:outline-none focus:border-accent/50 transition-colors"
                />
                <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
                {isSearching ? (
                  <Loader2 size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-accent animate-spin" />
                ) : searchQuery ? (
                  <button
                    onClick={() => setSearchQuery('')}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-white transition-colors"
                  >
                    <X size={14} />
                  </button>
                ) : null}
              </div>
            </div>
          )}

          {/* File List - Scrollable, takes remaining space */}
          <div
            className={cn(
              "flex-1 overflow-y-auto overflow-x-hidden p-2 min-h-0 transition-colors",
              isRootDragOver && "bg-accent/10 border-2 border-accent/50 border-dashed"
            )}
            onDragOver={handleRootDragOver}
            onDragLeave={handleRootDragLeave}
            onDrop={handleRootDrop}
            onContextMenu={handleRootContextMenu}
          >
            {project ? (
              filteredFiles.length > 0 ? (
                <>
                  {filteredFiles.map((node: FileTreeNode) => (
                    <TreeNode
                      key={node.path}
                      node={node}
                      level={0}
                      expandedFolders={expandedFolders}
                      toggleFolder={toggleFolder}
                      selectFile={selectFile}
                      deleteSelectedFile={handleDeleteFile}
                      selectedFile={selectedFile}
                      onRename={handleRename}
                      onMove={handleMove}
                      onCopy={copyToClipboard}
                      onCut={cutToClipboard}
                      onPaste={handlePaste}
                      onDuplicate={handleDuplicate}
                      onNewFile={handleNewFileFromContext}
                      onNewFolder={handleNewFolderFromContext}
                      clipboard={clipboard}
                      checkDepth={checkDepth}
                    />
                  ))}
                </>
              ) : (
                <div
                  className="text-center py-8 text-zinc-500 text-xs font-mono h-full flex flex-col items-center justify-center"
                  onContextMenu={handleRootContextMenu}
                >
                  {searchQuery ? (
                    <>
                      <Search size={32} className="mx-auto mb-2 opacity-30" />
                      <p>No files found matching "{searchQuery}"</p>
                      <button
                        onClick={() => setSearchQuery('')}
                        className="mt-3 text-accent hover:text-accent/80 transition-colors"
                      >
                        Clear search
                      </button>
                    </>
                  ) : (
                    <>
                      <FolderOpen size={32} className="mx-auto mb-2 opacity-30" />
                      <p>No files in this directory</p>
                      {projectPath && (
                        <p className="text-[10px] text-zinc-600 mt-2 px-4 break-all">
                          Path: {projectPath}
                        </p>
                      )}
                      <p className="text-[10px] text-zinc-500 mt-4">
                        Right-click to create files/folders
                      </p>
                    </>
                  )}
                </div>
              )
            ) : (
              <div className="text-center py-8 text-zinc-500 text-xs font-mono">
                <FolderOpen size={32} className="mx-auto mb-2 opacity-30" />
                <p>No project loaded</p>
              </div>
            )}
          </div>

          {/* Root Context Menu */}
          {rootContextMenu && project && (
            <ContextMenu
              x={rootContextMenu.x}
              y={rootContextMenu.y}
              node={{
                name: project.name || 'Project',
                path: projectPath || '',
                isDir: true,
              }}
              onClose={() => setRootContextMenu(null)}
              onNewFile={(node) => {
                dispatch({ type: 'SET_NEW_ITEM', itemType: 'file', parentNode: node });
                setRootContextMenu(null);
              }}
              onNewFolder={(node) => {
                dispatch({ type: 'SET_NEW_ITEM', itemType: 'folder', parentNode: node });
                setRootContextMenu(null);
              }}
              hasClipboard={!!clipboard}
              onPaste={() => {
                if (projectPath) {
                  handlePaste({
                    name: project.name || 'Project',
                    path: projectPath,
                    isDir: true,
                  });
                }
                setRootContextMenu(null);
              }}
            />
          )}

          {/* Action Buttons - Fixed at bottom */}
          {project && (
            <div className="flex-shrink-0 p-3 border-t border-white/5 bg-zinc-900/95 backdrop-blur-sm z-10 min-h-[80px]">
              <div className="grid grid-cols-2 gap-2">
                <button
                  onClick={handleServeToggle}
                  disabled={isProcessing || !depCheck.canServe}
                  title={!depCheck.canServe ? "Hugo is required to serve sites. Please install Hugo manually." : ""}
                  className="px-3 py-2 bg-green-500/10 hover:bg-green-500/20 text-green-400 border border-green-500/30 rounded-sm text-xs font-mono transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {serverRunning ? (
                    <StopCircle size={14} />
                  ) : (
                    <Play size={14} />
                  )}
                  {serverRunning ? "Stop Server" : "Serve Site"}
                </button>
                <button
                  onClick={async () => {
                    // Check dependencies first
                    if (!depCheck.canLaunch) {
                      if (depCheck.needsUpdate) {
                        onStatusChange?.({
                          type: 'info',
                          message: depCheck.updateMessage || 'Please update tools to the latest version before launching to Walrus'
                        });
                      } else {
                        const missingList = depCheck.missingDeps.join(', ');
                        onStatusChange?.({
                          type: 'info',
                          message: `Missing dependencies: ${missingList}. View installation instructions.`
                        });
                        setShowAutoInstallModal(true);
                      }
                      return;
                    }

                    // If server is running, stop it first
                    if (serverRunning) {
                      onStatusChange?.({
                        type: 'info',
                        message: 'Stopping server for launch...'
                      });

                      try {
                        const { StopServe } = await import('../../wailsjs/go/main/App');
                        await StopServe();
                        setServerRunning(false);
                        onStatusChange?.({
                          type: 'success',
                          message: 'Server stopped, proceeding to launch'
                        });
                      } catch (err) {
                        onStatusChange?.({
                          type: 'error',
                          message: `Failed to stop server: ${err}`
                        });
                        return;
                      }
                    }

                    // Proceed with launch
                    if (isDeployed) {
                      // For updates, go directly to deployment modal
                      setShowDeploymentModal(true);
                    } else {
                      // For new deployments, show launch config modal first
                      setShowLaunchModal(true);
                    }
                  }}
                  disabled={isProcessing || !depCheck.canLaunch}
                  title={!depCheck.canLaunch ? (depCheck.needsUpdate ? depCheck.updateMessage : `Missing: ${depCheck.missingDeps.join(', ')}`) : ""}
                  className={cn(
                    "px-3 py-2 rounded-sm text-xs font-mono transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed border",
                    isDeployed
                      ? "bg-orange-500/10 hover:bg-orange-500/20 text-orange-400 border-orange-500/30"
                      : "bg-accent/10 hover:bg-accent/20 text-accent border-accent/30"
                  )}
                >
                  {isDeployed ? (
                    <>
                      <RefreshCw size={14} />
                      Update on Walrus
                    </>
                  ) : (
                    <>
                      <Rocket size={14} />
                      Launch to Walrus
                    </>
                  )}
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Collapse Button - Outside card, at the edge */}
        {!sidebarCollapsed && (
          <button
            onClick={() => dispatch({ type: 'TOGGLE_SIDEBAR' })}
            className="absolute -right-3.5 top-1/2 -translate-y-1/2 z-50 w-7 h-16 bg-zinc-800/95 hover:bg-zinc-700 border border-white/10 rounded-lg flex items-center justify-center transition-all shadow-xl opacity-0 group-hover:opacity-100"
            style={{ 
              backdropFilter: 'blur(8px)',
              transition: 'opacity 0.2s ease-in-out'
            }}
            title="Hide sidebar"
          >
            <ChevronLeft size={16} className="text-zinc-400" />
          </button>
        )}
      </div>

      {/* Expand Button - Shows when sidebar is collapsed */}
      {sidebarCollapsed && (
        <button
          onClick={() => dispatch({ type: 'TOGGLE_SIDEBAR' })}
          className="absolute left-3 top-1/2 -translate-y-1/2 z-40 w-7 h-16 bg-zinc-800/95 hover:bg-zinc-700 border border-white/10 rounded-lg flex items-center justify-center transition-all shadow-xl hover:w-9"
          style={{ 
            backdropFilter: 'blur(8px)',
            transition: 'all 0.2s ease-in-out'
          }}
          title="Show sidebar"
        >
          <ChevronRight size={16} className="text-zinc-400" />
        </button>
      )}

      {/* Right Panel - Editor and Preview */}
      <div
        className="flex-1 flex flex-col relative min-h-0 bg-gradient-to-br from-zinc-900/95 via-zinc-900/90 to-black/95 border border-white/10 rounded-sm overflow-hidden transition-all duration-300 ease-in-out group shadow-2xl"
        style={{ height: "100%" }}
      >
          {/* Card corner decorations */}
          <div className="absolute top-0 left-0 w-3 h-3 border-t-2 border-l-2 border-accent/30 z-20 rounded-tl-sm" />
          <div className="absolute top-0 right-0 w-3 h-3 border-t-2 border-r-2 border-accent/30 z-20 rounded-tr-sm" />
          <div className="absolute bottom-0 left-0 w-3 h-3 border-b-2 border-l-2 border-accent/30 z-20 rounded-bl-sm" />
          <div className="absolute bottom-0 right-0 w-3 h-3 border-b-2 border-r-2 border-accent/30 z-20 rounded-br-sm" />

          {selectedFile ? (
            <>
              {/* Enhanced Toolbar */}
              <div className="px-4 py-3 border-b border-white/10 bg-gradient-to-r from-black/40 via-zinc-900/40 to-black/40 flex items-center justify-between flex-shrink-0 backdrop-blur-sm relative">
                {/* Subtle glow effect */}
                <div className="absolute inset-0 bg-gradient-to-r from-transparent via-accent/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />

                <div className="flex items-center gap-3 relative z-10">
                  <div className="p-1.5 bg-accent/10 border border-accent/20 rounded">
                    <FileText size={14} className="text-accent" />
                  </div>
                  <div className="flex flex-col">
                    <span className="text-xs font-mono text-white font-semibold">
                      {selectedFile.name}
                    </span>
                    <span className="text-[10px] font-mono text-zinc-500 truncate max-w-[300px]">
                      {selectedFile.path}
                    </span>
                  </div>
                </div>

                <div className="flex items-center gap-2 relative z-10">
                  <div className="w-px h-6 bg-gradient-to-b from-transparent via-white/20 to-transparent mx-1"></div>
                  <button
                    onClick={() => setViewMode("editor")}
                    className={cn(
                      "px-3 py-1.5 rounded text-xs font-mono transition-all relative overflow-hidden",
                      viewMode === "editor"
                        ? "bg-accent/20 text-accent border border-accent/40 shadow-lg shadow-accent/20"
                        : "bg-white/5 text-zinc-400 hover:text-white hover:bg-white/10 border border-transparent"
                    )}
                  >
                    {viewMode === "editor" && (
                      <div className="absolute inset-0 bg-gradient-to-r from-accent/10 via-accent/5 to-accent/10 animate-pulse" />
                    )}
                    <span className="relative">Editor</span>
                  </button>
                  {selectedFile?.name.endsWith(".md") && (
                    <>
                      <button
                        onClick={() => setViewMode("split")}
                        className={cn(
                          "px-3 py-1.5 rounded text-xs font-mono transition-all relative overflow-hidden",
                          viewMode === "split"
                            ? "bg-accent/20 text-accent border border-accent/40 shadow-lg shadow-accent/20"
                            : "bg-white/5 text-zinc-400 hover:text-white hover:bg-white/10 border border-transparent"
                        )}
                      >
                        {viewMode === "split" && (
                          <div className="absolute inset-0 bg-gradient-to-r from-accent/10 via-accent/5 to-accent/10 animate-pulse" />
                        )}
                        <span className="relative">Split</span>
                      </button>
                      <button
                        onClick={() => setViewMode("preview")}
                        className={cn(
                          "px-3 py-1.5 rounded text-xs font-mono transition-all relative overflow-hidden",
                          viewMode === "preview"
                            ? "bg-accent/20 text-accent border border-accent/40 shadow-lg shadow-accent/20"
                            : "bg-white/5 text-zinc-400 hover:text-white hover:bg-white/10 border border-transparent"
                        )}
                      >
                        {viewMode === "preview" && (
                          <div className="absolute inset-0 bg-gradient-to-r from-accent/10 via-accent/5 to-accent/10 animate-pulse" />
                        )}
                        <span className="relative">Preview</span>
                      </button>
                    </>
                  )}
                </div>
              </div>

              {/* Content Area */}
              <div className="flex flex-col flex-1 min-h-0 overflow-hidden">
                {/* Editor and Preview Container */}
                <div className="flex flex-1 min-h-0 overflow-hidden">
                  {/* Enhanced Editor */}
                  {(viewMode === "editor" || viewMode === "split") && (
                    <div
                      className={cn(
                        "flex flex-col relative bg-gradient-to-br from-black/60 via-zinc-900/50 to-black/60 overflow-hidden",
                        viewMode === "split"
                          ? "w-1/2 border-r border-white/10"
                          : "w-full"
                      )}
                    >
                      {loading && <LoadingOverlay message="Loading file..." />}

                      {/* Editor with enhanced styling */}
                      <div className="flex-1 relative overflow-hidden">
                        {/* Subtle background pattern */}
                        <div className="absolute inset-0 opacity-5 pointer-events-none" style={{
                          backgroundImage: 'repeating-linear-gradient(0deg, transparent, transparent 20px, rgba(255,255,255,0.03) 20px, rgba(255,255,255,0.03) 21px)'
                        }} />

                        <textarea
                          value={fileContent}
                          onChange={handleTextareaChange}
                          onClick={handleTextareaClick}
                          onKeyUp={handleTextareaKeyUp}
                          className="relative z-10 flex-1 w-full h-full p-6 bg-transparent text-sm font-mono text-zinc-200 leading-relaxed resize-none focus:outline-none focus:ring-2 focus:ring-accent/20 focus:ring-inset scrollbar-thin scrollbar-thumb-zinc-700 hover:scrollbar-thumb-zinc-600 scrollbar-track-transparent transition-all"
                          placeholder="Start writing your content..."
                          autoComplete="off"
                          autoCapitalize="off"
                          autoCorrect="off"
                          spellCheck={false}
                          disabled={loading}
                          style={{
                            minHeight: 0,
                            letterSpacing: '0.01em',
                            tabSize: 2,
                          }}
                        />
                      </div>
                    </div>
                  )}

                {/* Enhanced Preview */}
                {(viewMode === "preview" || viewMode === "split") && (
                  <div
                    className={cn(
                      "overflow-y-auto bg-gradient-to-br from-black/60 via-zinc-900/40 to-black/60 min-h-0 h-full relative scrollbar-thin scrollbar-thumb-zinc-700 hover:scrollbar-thumb-zinc-600 scrollbar-track-transparent",
                      viewMode === "split" ? "w-1/2" : "w-full"
                    )}
                  >
                    {loading && <LoadingOverlay message="Loading preview..." />}

                    {/* Preview label */}
                    <div className="sticky top-0 z-10 px-6 py-2 bg-zinc-900/80 backdrop-blur-sm border-b border-white/5">
                      <span className="text-[10px] font-mono text-zinc-500 uppercase tracking-wider">
                        Preview
                      </span>
                    </div>

                    <div className="p-8 prose prose-invert prose-sm max-w-none">
                      {selectedFile.name.endsWith(".md") ? (
                        <div
                          className="prose-headings:text-white prose-p:text-zinc-300 prose-a:text-accent prose-a:no-underline hover:prose-a:underline prose-strong:text-white prose-code:text-accent prose-code:bg-accent/10 prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded prose-pre:bg-black/60 prose-pre:border prose-pre:border-white/10"
                          dangerouslySetInnerHTML={{
                            __html: renderMarkdown(fileContent),
                          }}
                        />
                      ) : (
                        <pre className="text-sm text-zinc-300 font-mono whitespace-pre-wrap leading-relaxed bg-black/40 p-6 rounded border border-white/10">
                          {fileContent}
                        </pre>
                      )}
                    </div>
                  </div>
                )}
                </div>

                {/* Enhanced Status Bar - Outside editor/preview, always at bottom */}
                <div className="px-4 py-2.5 border-t border-white/10 bg-gradient-to-r from-black/60 via-zinc-900/50 to-black/60 backdrop-blur-sm flex items-center justify-between flex-shrink-0">
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <div className="w-1.5 h-1.5 rounded-full bg-green-400 animate-pulse shadow-lg shadow-green-400/50" />
                      <span className="text-[10px] font-mono text-zinc-500 uppercase tracking-wider">
                        Ln {cursorPosition.line}, Col {cursorPosition.column}
                      </span>
                    </div>
                    <div className="w-px h-4 bg-white/10" />
                    <span className="text-[10px] font-mono text-zinc-500">
                      {getLineCount(fileContent)} lines
                    </span>
                    <div className="w-px h-4 bg-white/10" />
                    <span className="text-[10px] font-mono text-zinc-500">
                      {getWordCount(fileContent)} words
                    </span>
                    <div className="w-px h-4 bg-white/10" />
                    <span className="text-[10px] font-mono text-zinc-500">
                      {fileContent.length} chars
                    </span>
                    <div className="w-px h-4 bg-white/10" />
                    <span className="text-[10px] font-mono text-zinc-500 uppercase tracking-wider">
                      UTF-8
                    </span>
                  </div>

                  <div className="flex items-center gap-2">
                    {selectedFile && (selectedFile.path.toLowerCase().endsWith('.md') || selectedFile.path.toLowerCase().endsWith('.markdown')) && (
                      <button
                        onClick={() => {
                          if (aiConfigured) {
                            setShowAIUpdateModal(true);
                          } else if (onStatusChange) {
                            onStatusChange({
                              type: 'error',
                              message: 'AI not configured: Please configure AI in Settings to use AI features.'
                            });
                          }
                        }}
                        disabled={!aiConfigured}
                        className={cn(
                          "px-4 py-1.5 rounded text-xs font-mono flex items-center gap-2 transition-all border relative overflow-hidden group/btn",
                          aiConfigured
                            ? "bg-purple-500/10 hover:bg-purple-500/20 text-purple-400 border-purple-500/30 hover:border-purple-400/50 hover:shadow-lg hover:shadow-purple-500/20"
                            : "bg-zinc-800 text-zinc-600 border-zinc-700 cursor-not-allowed opacity-50"
                        )}
                      >
                        {aiConfigured && (
                          <div className="absolute inset-0 bg-gradient-to-r from-purple-500/0 via-purple-500/10 to-purple-500/0 translate-x-[-100%] group-hover/btn:translate-x-[100%] transition-transform duration-1000" />
                        )}
                        <Sparkles size={14} className="relative z-10" />
                        <span className="relative z-10">AI Update</span>
                      </button>
                    )}
                    <button
                      onClick={handleSave}
                      disabled={saving}
                      className={cn(
                        "px-4 py-1.5 rounded text-xs font-mono flex items-center gap-2 transition-all border relative overflow-hidden group/btn",
                        saving
                          ? "bg-accent/20 text-accent border-accent/40 opacity-70"
                          : "bg-accent/10 hover:bg-accent/20 text-accent border-accent/30 hover:border-accent/50 hover:shadow-lg hover:shadow-accent/20"
                      )}
                    >
                      {!saving && (
                        <div className="absolute inset-0 bg-gradient-to-r from-accent/0 via-accent/10 to-accent/0 translate-x-[-100%] group-hover/btn:translate-x-[100%] transition-transform duration-1000" />
                      )}
                      {saving ? (
                        <Loader2 size={14} className="animate-spin relative z-10" />
                      ) : (
                        <Save size={14} className="relative z-10" />
                      )}
                      <span className="relative z-10">{savingStatus || "Save"}</span>
                    </button>
                  </div>
                </div>
              </div>
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center relative overflow-hidden">
              {/* Animated background gradient */}
              <div className="absolute inset-0 bg-gradient-to-br from-accent/5 via-transparent to-purple-500/5 opacity-50" />
              <div className="absolute inset-0" style={{
                backgroundImage: 'radial-gradient(circle at 50% 50%, rgba(99, 102, 241, 0.1) 0%, transparent 50%)',
                backgroundSize: '100px 100px'
              }} />

              <div className="text-center relative z-10">
                {error ? (
                  <>
                    <div className="relative inline-block mb-6">
                      <div className="absolute inset-0 bg-red-400/20 blur-2xl rounded-full" />
                      <AlertCircle size={56} className="relative text-red-400 animate-pulse" />
                    </div>
                    <p className="text-base text-red-400 font-mono mb-3 font-semibold">
                      Error loading project
                    </p>
                    <p className="text-xs text-zinc-500 font-mono max-w-md px-4">
                      {error}
                    </p>
                  </>
                ) : (
                  <>
                    <div className="relative inline-block mb-6">
                      <div className="absolute inset-0 bg-accent/20 blur-3xl rounded-full animate-pulse" />
                      <FileText size={56} className="relative text-zinc-600" />
                    </div>
                    <p className="text-base text-zinc-400 font-mono mb-2">
                      No file selected
                    </p>
                    <p className="text-xs text-zinc-600 font-mono">
                      Select a file from the sidebar to start editing
                    </p>
                  </>
                )}
              </div>
            </div>
          )}
      </div>

      {/* New File/Folder Modal */}
      {showNewItemModal && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
            className="bg-zinc-900 border border-white/10 rounded-sm p-6 max-w-md w-full"
          >
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-display text-white flex items-center gap-2">
                {newItemType === 'file' ? (
                  <>
                    <FilePlus size={20} className="text-accent" />
                    New File
                  </>
                ) : (
                  <>
                    <FolderPlus size={20} className="text-accent" />
                    New Folder
                  </>
                )}
              </h2>
              <button
                onClick={closeNewItemModal}
                className="p-1 hover:bg-white/10 rounded-sm transition-colors"
              >
                <X size={18} className="text-zinc-400" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-xs font-mono text-zinc-400 mb-2">
                  {newItemType === 'file' ? 'File Name' : 'Folder Name'}
                </label>
                <input
                  type="text"
                  value={newItemName}
                  onChange={(e) => setNewItemName(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && !isCreating) {
                      handleCreateNewItem();
                    }
                  }}
                  placeholder={newItemType === 'file' ? 'my-new-post.md' : 'my-folder'}
                  autoComplete="off"
                  autoCapitalize="off"
                  autoCorrect="off"
                  spellCheck="false"
                  autoFocus
                  className="w-full px-4 py-2 bg-zinc-900 border border-zinc-700 rounded-sm text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-accent transition-colors"
                />
                {newItemType === 'file' ? (
                  <div className="text-xs text-zinc-500 font-mono mt-2 space-y-1">
                    <p>• Include file extension: my-file.md, style.css, config.json</p>
                    <p>• Use slashes for subdirectories: posts/my-post.md</p>
                  </div>
                ) : (
                  <p className="text-xs text-zinc-500 font-mono mt-2">
                    • Use slashes for nested folders: posts/2024
                  </p>
                )}
              </div>

              <div className="flex gap-3">
                <motion.button
                  onClick={closeNewItemModal}
                  disabled={isCreating}
                  variants={buttonVariants}
                  whileHover="hover"
                  whileTap="tap"
                  className="flex-1 px-4 py-2 bg-white/5 hover:bg-white/10 text-zinc-400 hover:text-white rounded-sm text-sm font-mono transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Cancel
                </motion.button>
                <motion.button
                  onClick={handleCreateNewItem}
                  disabled={isCreating || !newItemName.trim()}
                  variants={buttonVariants}
                  whileHover="hover"
                  whileTap="tap"
                  className="flex-1 px-4 py-2 bg-accent/10 hover:bg-accent/20 text-accent border border-accent/30 rounded-sm text-sm font-mono transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                >
                  {isCreating ? (
                    <>
                      <Loader2 size={14} className="animate-spin" />
                      Creating...
                    </>
                  ) : (
                    <>
                      <Plus size={14} />
                      Create
                    </>
                  )}
                </motion.button>
              </div>
            </div>
          </motion.div>
        </div>
      )}

      {/* Launch Configuration Modal */}
      <LaunchModal
        isOpen={showLaunchModal}
        onClose={() => setShowLaunchModal(false)}
        onLaunch={handleLaunchConfig}
        project={project || undefined}
      />

      {/* Deployment Modal */}
      <DeploymentModal
        isOpen={showDeploymentModal}
        isUpdate={isDeployed}
        projectName={launchConfig?.projectName || project?.name || ""}
        sitePath={projectPath}
        network={project?.network}
        currentObjectId={project?.objectId}
        deployedWallet={project?.wallet}
        currentEpochs={project?.epochs}
        onClose={() => {
          setShowDeploymentModal(false);
          setLaunchConfig(null);
        }}
        onDeploy={handleDeployment}
      />

      {/* AI Generate Modal */}
      <AIGenerateModal
        isOpen={showAIGenerateModal}
        onClose={() => setShowAIGenerateModal(false)}
        sitePath={projectPath || ""}
        onSuccess={(filePath) => {
          // Optionally select the generated file
          if (onStatusChange) {
            onStatusChange({
              type: 'success',
              message: `Content generated: ${filePath}`
            });
          }
        }}
        onStatusChange={onStatusChange}
      />

      {/* AI Update Modal */}
      <AIUpdateModal
        isOpen={showAIUpdateModal}
        onClose={() => setShowAIUpdateModal(false)}
        sitePath={projectPath || ""}
        filePath={selectedFile?.path || ""}
        currentContent={fileContent}
        onSuccess={async () => {
          // Reload the file content after update
          if (selectedFile) {
            // Small delay to ensure file is written
            await new Promise(resolve => setTimeout(resolve, 100));
            selectFile(selectedFile);
          }
        }}
        onStatusChange={onStatusChange}
      />

      {/* Install Instructions Modal */}
      <InstallInstructionsModal
        isOpen={showAutoInstallModal}
        onClose={() => setShowAutoInstallModal(false)}
        missingDeps={depCheck.missingDeps}
        onRefreshStatus={async () => {
          if (onRefreshHealth) {
            await onRefreshHealth();
          }
        }}
      />

      {/* Theme Modal */}
      <ThemeModal
        isOpen={showThemeModal}
        onClose={() => setShowThemeModal(false)}
        sitePath={projectPath || ""}
        onSuccess={(themeName) => {
          if (onStatusChange) {
            onStatusChange({
              type: 'success',
              message: `Theme '${themeName}' installed. Run Build to apply changes.`
            });
          }
        }}
        onStatusChange={onStatusChange}
      />

      {/* Delete Confirmation Modal */}
      {showDeleteModal && fileToDelete && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
            className="bg-zinc-900 border border-red-500/30 rounded-sm p-6 max-w-md w-full"
          >
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-display text-white flex items-center gap-2">
                <AlertCircle size={20} className="text-red-400" />
                Confirm Delete
              </h2>
              <button
                onClick={() => {
                  setShowDeleteModal(false);
                  setFileToDelete(null);
                }}
                className="p-1 hover:bg-white/10 rounded-sm transition-colors"
              >
                <X size={18} className="text-zinc-400" />
              </button>
            </div>

            <div className="space-y-4">
              <p className="text-sm text-zinc-300 font-mono">
                {fileToDelete.isDir ? (
                  <>
                    Are you sure you want to delete the directory <span className="text-red-400 font-semibold">"{fileToDelete.name}"</span> and all its contents?
                  </>
                ) : (
                  <>
                    Are you sure you want to delete <span className="text-red-400 font-semibold">"{fileToDelete.name}"</span>?
                  </>
                )}
              </p>

              <p className="text-xs text-zinc-500 font-mono">
                This action cannot be undone.
              </p>

              <div className="flex gap-3 mt-6">
                <motion.button
                  onClick={() => {
                    setShowDeleteModal(false);
                    setFileToDelete(null);
                  }}
                  variants={buttonVariants}
                  whileHover="hover"
                  whileTap="tap"
                  className="flex-1 px-4 py-2 bg-white/5 hover:bg-white/10 text-zinc-400 hover:text-white rounded-sm text-sm font-mono transition-all"
                >
                  Cancel
                </motion.button>
                <motion.button
                  onClick={confirmDelete}
                  variants={buttonVariants}
                  whileHover="hover"
                  whileTap="tap"
                  className="flex-1 px-4 py-2 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/30 rounded-sm text-sm font-mono transition-all flex items-center justify-center gap-2"
                >
                  <Trash2 size={14} />
                  Delete
                </motion.button>
              </div>
            </div>
          </motion.div>
        </div>
      )}

      {/* Properties Panel */}
      <AnimatePresence>
        {showPropertiesPanel && selectedFile && (
          <motion.div
            initial={{ x: '100%' }}
            animate={{ x: 0 }}
            exit={{ x: '100%' }}
            transition={{ type: 'spring', damping: 25, stiffness: 200 }}
            className="fixed right-0 top-0 bottom-0 w-80 bg-zinc-900/95 backdrop-blur-sm border-l border-white/10 z-40 shadow-2xl"
          >
            {/* Header */}
            <div className="p-4 border-b border-white/10 bg-black/20">
              <div className="flex items-center justify-between mb-2">
                <h3 className="text-sm font-display text-white flex items-center gap-2">
                  <Info size={16} className="text-accent" />
                  Properties
                </h3>
                <button
                  onClick={() => setShowPropertiesPanel(false)}
                  className="p-1 hover:bg-white/10 rounded-sm transition-colors"
                >
                  <X size={16} className="text-zinc-400" />
                </button>
              </div>
              <p className="text-xs text-zinc-500 font-mono">Press Ctrl/⌘+I to toggle</p>
            </div>

            {/* Content */}
            <div className="p-4 space-y-6 overflow-y-auto max-h-[calc(100vh-80px)]">
              {/* File/Folder Icon and Name */}
              <div className="flex items-start gap-3">
                {selectedFile.isDir ? (
                  <Folder size={40} className="text-accent flex-shrink-0" />
                ) : (
                  <FileText size={40} className="text-blue-400 flex-shrink-0" />
                )}
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold text-white break-all">{selectedFile.name}</p>
                  <p className="text-xs text-zinc-500 font-mono mt-1">{selectedFile.isDir ? 'Folder' : 'File'}</p>
                </div>
              </div>

              {/* Divider */}
              <div className="h-px bg-white/10" />

              {/* Properties */}
              <div className="space-y-4">
                {/* Size */}
                <div className="flex items-start gap-3">
                  <HardDrive size={16} className="text-zinc-500 flex-shrink-0 mt-0.5" />
                  <div className="flex-1">
                    <p className="text-xs text-zinc-500 font-mono mb-1">Size</p>
                    <p className="text-sm text-white font-mono">
                      {selectedFile.isDir ? (
                        loadingStats ? (
                          <span className="flex items-center gap-2">
                            <Loader2 size={12} className="animate-spin" />
                            Calculating...
                          </span>
                        ) : folderStats ? (
                          formatFileSize(folderStats.totalSize)
                        ) : (
                          'Unknown'
                        )
                      ) : (
                        formatFileSize(selectedFile.size ?? 0)
                      )}
                    </p>
                  </div>
                </div>

                {/* Folder Stats */}
                {selectedFile.isDir && folderStats && (
                  <div className="flex items-start gap-3">
                    <Files size={16} className="text-zinc-500 flex-shrink-0 mt-0.5" />
                    <div className="flex-1">
                      <p className="text-xs text-zinc-500 font-mono mb-1">Contents</p>
                      <p className="text-sm text-white font-mono">
                        {folderStats.fileCount} file{folderStats.fileCount !== 1 ? 's' : ''}, {folderStats.folderCount} folder{folderStats.folderCount !== 1 ? 's' : ''}
                      </p>
                    </div>
                  </div>
                )}

                {/* Modified Date */}
                <div className="flex items-start gap-3">
                  <Calendar size={16} className="text-zinc-500 flex-shrink-0 mt-0.5" />
                  <div className="flex-1">
                    <p className="text-xs text-zinc-500 font-mono mb-1">Modified</p>
                    <p className="text-sm text-white font-mono">{formatDate(selectedFile.modified)}</p>
                  </div>
                </div>

                {/* Path */}
                <div className="flex items-start gap-3">
                  <FolderOpen size={16} className="text-zinc-500 flex-shrink-0 mt-0.5" />
                  <div className="flex-1">
                    <p className="text-xs text-zinc-500 font-mono mb-1">Location</p>
                    <p className="text-xs text-zinc-400 font-mono break-all">{selectedFile.path}</p>
                  </div>
                </div>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};
