import React, { useState, useEffect } from "react";
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
} from "lucide-react";
import { LoadingOverlay } from "../components/ui";
import { TreeNode } from "../components/file-tree";
import { DeploymentModal, LaunchModal, AIGenerateModal, AIUpdateModal, InstallInstructionsModal } from "../components/modals";
import type { DeploymentParams, DeploymentResult } from "../components/modals/DeploymentModal";
import type { LaunchConfig } from "../components/modals/LaunchModal";
import { itemVariants, buttonVariants, iconButtonVariants } from "../utils/constants";
import { cn, renderMarkdown } from "../utils";
import { useEditProject, useDependencyCheck } from "../hooks";
import { Project, SystemHealth } from "../types";

const SAVE_STATUS_DURATION = 2000; // ms

interface EditProps {
  aiConfigured?: boolean;
  systemHealth?: SystemHealth;
  hasUpdates?: boolean;
  updatingTools?: string[];
  onStatusChange?: (status: { type: 'success' | 'error' | 'info'; message: string }) => void;
  onProjectUpdate?: () => Promise<void>;
  onInstallDeps?: (tools: string[]) => Promise<void>;
  onRefreshHealth?: () => Promise<void>;
}

export const Edit: React.FC<EditProps> = ({
  aiConfigured = false,
  systemHealth,
  hasUpdates = false,
  updatingTools = [],
  onProjectUpdate,
  onStatusChange,
  onInstallDeps,
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
    setFileContent,
    loadProject,
    selectFile,
    saveFile,
    deleteFile,
    toggleFolder,
    reset,
  } = useEditProject();

  const [showProjectActionsMenu, setShowProjectActionsMenu] = useState(false);
  const [viewMode, setViewMode] = useState<"split" | "editor" | "preview">("split");
  const [savingStatus, setSavingStatus] = useState("");
  const [showNewItemModal, setShowNewItemModal] = useState(false);
  const [newItemType, setNewItemType] = useState<"file" | "folder">("file");
  const [newItemName, setNewItemName] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const [showLaunchModal, setShowLaunchModal] = useState(false);
  const [showDeploymentModal, setShowDeploymentModal] = useState(false);
  const [isDeployed, setIsDeployed] = useState(false);
  const [launchConfig, setLaunchConfig] = useState<LaunchConfig | null>(null);
  const [showAIGenerateModal, setShowAIGenerateModal] = useState(false);
  const [showAIUpdateModal, setShowAIUpdateModal] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [showAutoInstallModal, setShowAutoInstallModal] = useState(false);

  // Dependency check
  const depCheck = useDependencyCheck({
    systemHealth,
    hasUpdates,
    updatingTools
  });

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

  const handleDeleteFile = async (file: any) => {
    if (file.isDir) {
      const confirmed = window.confirm(
        `Are you sure you want to delete the directory "${file.name}" and all its contents?`
      );
      if (!confirmed) return;
    } else {
      const confirmed = window.confirm(
        `Are you sure you want to delete "${file.name}"?`
      );
      if (!confirmed) return;
    }

    await deleteFile(file.path);
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
    const projectPath = project?.path || project?.sitePath;
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
        const { Serve, GetServerURL, OpenInBrowser } = await import('../../wailsjs/go/main/App');
        await Serve({
          sitePath: projectPath,
          port: 1313,
          drafts: true,
          expired: false,
          future: false,
        });
        setServerRunning(true);

        // Get URL and open browser
        setTimeout(async () => {
          try {
            const url = await GetServerURL();
            if (url) {
              await OpenInBrowser(url);
              onStatusChange?.({
                type: 'success',
                message: `Server running: ${url}`,
              });
            }
          } catch (err) {
            console.error('Failed to get server URL:', err);
          }
        }, 1500);
      }
    } catch (err: any) {
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
    } catch (err: any) {
      onStatusChange?.({
        type: 'error',
        message: `Failed to open: ${err?.toString()}`,
      });
    }
  };

  const handleDeployment = async (params: DeploymentParams): Promise<DeploymentResult> => {
    try {
      const { LaunchWizard } = await import('../../wailsjs/go/main/App');
      
      const projectPath = project?.path || project?.sitePath;
      if (!projectPath) {
        return {
          success: false,
          error: "Project path not found",
        };
      }

      // Call the actual LaunchWizard API
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
    // Remove .md extension for validation
    const cleanSlug = slug.replace(/\.md$/, '');
    // Allow letters, numbers, hyphens, underscores, and forward slashes (for subdirectories)
    const validSlug = /^[a-zA-Z0-9_\-/]+$/;
    return validSlug.test(cleanSlug) && cleanSlug.length > 0 && cleanSlug.length < 100;
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

    const projectPath = project?.path || project?.sitePath;
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
      
      if (newItemType === 'file') {
        // Ensure .md extension for files
        const fileName = newItemName.endsWith('.md') ? newItemName : `${newItemName}.md`;
        const filePath = `${projectPath}/content/${fileName}`;
        
        // Extract title from filename (remove path and extension)
        const title = fileName.split('/').pop()?.replace('.md', '') || 'New Content';
        
        // Create file with basic frontmatter
        const content = `---
title: "${title}"
date: ${new Date().toISOString()}
draft: false
---

# ${title}

Start writing your content here...
`;
        
        const result = await CreateFile(filePath, content);
        if (result.success) {
          await loadProject(project);
          setShowNewItemModal(false);
          setNewItemName('');
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
        const folderPath = `${projectPath}/content/${newItemName}`;
        const result = await CreateDirectory(folderPath);
        if (result.success) {
          await loadProject(project);
          setShowNewItemModal(false);
          setNewItemName('');
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
                      setShowProjectActionsMenu(!showProjectActionsMenu)
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
                            const projectPath = project.path || project.sitePath;
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
                            setNewItemType("file");
                            setShowNewItemModal(true);
                            setShowProjectActionsMenu(false);
                          }}
                          className="w-full px-4 py-2.5 text-left text-xs font-mono text-zinc-300 hover:bg-white/5 hover:text-white transition-all flex items-center gap-3 border-t border-white/5"
                        >
                          <FilePlus size={14} className="text-zinc-500" />
                          New File
                        </button>
                        <button
                          onClick={() => {
                            setNewItemType("folder");
                            setShowNewItemModal(true);
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

          {/* File List - Scrollable, takes remaining space */}
          <div className="flex-1 overflow-y-auto overflow-x-hidden p-2 min-h-0">
            {project ? (
              files.length > 0 ? (
                files.map((node: any) => (
                  <TreeNode
                    key={node.path}
                    node={node}
                    level={0}
                    expandedFolders={expandedFolders}
                    toggleFolder={toggleFolder}
                    selectFile={selectFile}
                    deleteSelectedFile={handleDeleteFile}
                    selectedFile={selectedFile}
                  />
                ))
              ) : (
                <div className="text-center py-8 text-zinc-500 text-xs font-mono">
                  <FolderOpen size={32} className="mx-auto mb-2 opacity-30" />
                  <p>No files in this directory</p>
                  {(project.path || project.sitePath) && (
                    <p className="text-[10px] text-zinc-600 mt-2 px-4 break-all">
                      Path: {project.path || project.sitePath}
                    </p>
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
            onClick={() => setSidebarCollapsed(true)}
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
          onClick={() => setSidebarCollapsed(false)}
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
        className="flex-1 flex flex-col relative min-h-0 bg-zinc-900/20 border border-white/10 rounded-sm overflow-hidden transition-all duration-300 ease-in-out"
        style={{ height: "100%" }}
      >
          {selectedFile ? (
            <>
              {/* Toolbar */}
              <div className="p-3 border-b border-white/5 bg-black/20 flex items-center justify-between flex-shrink-0">
                <div className="flex items-center gap-2">
                  <FileText size={16} className="text-zinc-500" />
                  <span className="text-xs font-mono text-zinc-300 truncate max-w-[200px]">
                    {selectedFile.name}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-px h-6 bg-white/10 mx-1"></div>
                  <button
                    onClick={() => setViewMode("editor")}
                    className={cn(
                      "px-3 py-1.5 rounded-sm text-xs font-mono transition-all",
                      viewMode === "editor"
                        ? "bg-accent/10 text-accent border border-accent/30"
                        : "bg-white/5 text-zinc-400 hover:text-white"
                    )}
                  >
                    Editor
                  </button>
                  {selectedFile?.name.endsWith(".md") && (
                    <>
                      <button
                        onClick={() => setViewMode("split")}
                        className={cn(
                          "px-3 py-1.5 rounded-sm text-xs font-mono transition-all",
                          viewMode === "split"
                            ? "bg-accent/10 text-accent border border-accent/30"
                            : "bg-white/5 text-zinc-400 hover:text-white"
                        )}
                      >
                        Split
                      </button>
                      <button
                        onClick={() => setViewMode("preview")}
                        className={cn(
                          "px-3 py-1.5 rounded-sm text-xs font-mono transition-all",
                          viewMode === "preview"
                            ? "bg-accent/10 text-accent border border-accent/30"
                            : "bg-white/5 text-zinc-400 hover:text-white"
                        )}
                      >
                        Preview
                      </button>
                    </>
                  )}
                </div>
              </div>

              {/* Content Area */}
              <div className="flex flex-1 min-h-0 overflow-hidden">
                {/* Editor */}
                {(viewMode === "editor" || viewMode === "split") && (
                  <div
                    className={cn(
                      "flex flex-col bg-black/40 relative",
                      viewMode === "split"
                        ? "w-1/2 border-r border-white/5"
                        : "w-full"
                    )}
                  >
                    {loading && <LoadingOverlay message="Loading file..." />}
                    <textarea
                      value={fileContent}
                      onChange={(e) => setFileContent(e.target.value)}
                      className="flex-1 w-full p-4 bg-transparent text-xs font-mono text-zinc-300 resize-none focus:outline-none scrollbar-thin scrollbar-thumb-zinc-800 scrollbar-track-transparent"
                      placeholder="Start typing..."
                      autoComplete="off"
                      autoCapitalize="off"
                      autoCorrect="off"
                      spellCheck={false}
                      disabled={loading}
                      style={{ minHeight: 0 }}
                    />
                    <div className="p-2 border-t border-white/5 bg-black/30 flex items-center justify-between flex-shrink-0 min-h-[50px]">
                      <span className="text-[10px] font-mono text-zinc-600">
                        {fileContent.length} characters
                      </span>
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
                              "px-4 py-1.5 rounded-sm text-xs font-mono flex items-center gap-2 transition-all border",
                              aiConfigured
                                ? "bg-purple-500/10 hover:bg-purple-500/20 text-purple-400 border-purple-500/30"
                                : "bg-zinc-800 text-zinc-600 border-zinc-700 cursor-not-allowed opacity-50"
                            )}
                          >
                            <Sparkles size={14} />
                            AI Update
                          </button>
                        )}
                        <button
                          onClick={handleSave}
                          disabled={saving}
                          className="px-4 py-1.5 bg-accent/10 hover:bg-accent/20 text-accent border border-accent/30 rounded-sm text-xs font-mono flex items-center gap-2 transition-all disabled:opacity-50"
                        >
                          {saving ? (
                            <Loader2 size={14} className="animate-spin" />
                          ) : (
                            <Save size={14} />
                          )}
                          {savingStatus || "Save"}
                        </button>
                      </div>
                    </div>
                  </div>
                )}

                {/* Preview */}
                {(viewMode === "preview" || viewMode === "split") && (
                  <div
                    className={cn(
                      "overflow-y-auto bg-black/40 min-h-0 h-full relative",
                      viewMode === "split" ? "w-1/2" : "w-full"
                    )}
                  >
                    {loading && <LoadingOverlay message="Loading preview..." />}
                    <div className="p-6 prose prose-invert prose-sm max-w-none">
                      {selectedFile.name.endsWith(".md") ? (
                        <div
                          dangerouslySetInnerHTML={{
                            __html: renderMarkdown(fileContent),
                          }}
                        />
                      ) : (
                        <pre className="text-xs text-zinc-300 font-mono whitespace-pre-wrap">
                          {fileContent}
                        </pre>
                      )}
                    </div>
                  </div>
                )}
              </div>
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center">
              <div className="text-center">
                {error ? (
                  <>
                    <AlertCircle size={48} className="mx-auto mb-4 text-red-400" />
                    <p className="text-sm text-red-400 font-mono mb-2">
                      Error loading project
                    </p>
                    <p className="text-xs text-zinc-500 font-mono">
                      {error}
                    </p>
                  </>
                ) : (
                  <>
                    <FileText size={48} className="mx-auto mb-4 text-zinc-600" />
                    <p className="text-sm text-zinc-500 font-mono">
                      Select a file to edit
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
                onClick={() => {
                  setShowNewItemModal(false);
                  setNewItemName('');
                }}
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
                  placeholder={newItemType === 'file' ? 'my-new-post' : 'my-folder'}
                  autoComplete="off"
                  autoCapitalize="off"
                  autoCorrect="off"
                  spellCheck="false"
                  autoFocus
                  className="w-full px-4 py-2 bg-zinc-900 border border-zinc-700 rounded-sm text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-accent transition-colors"
                />
                {newItemType === 'file' ? (
                  <div className="text-xs text-zinc-500 font-mono mt-2 space-y-1">
                    <p>• .md extension will be added automatically</p>
                    <p>• Use slashes for subdirectories: posts/my-post</p>
                  </div>
                ) : (
                  <p className="text-xs text-zinc-500 font-mono mt-2">
                    • Use slashes for nested folders: posts/2024
                  </p>
                )}
              </div>

              <div className="flex gap-3">
                <motion.button
                  onClick={() => {
                    setShowNewItemModal(false);
                    setNewItemName('');
                  }}
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
        sitePath={project?.path || project?.sitePath}
        network={project?.network}
        currentObjectId={project?.objectId}
        deployedWallet={project?.wallet}
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
        sitePath={project?.path || project?.sitePath || ""}
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
        sitePath={project?.path || project?.sitePath || ""}
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
    </div>
  );
};
