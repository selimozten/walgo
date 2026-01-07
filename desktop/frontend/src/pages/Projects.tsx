import React, { useState, useMemo, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
    Search,
    FolderOpen,
    Trash2,
    Folder,
    AlertCircle,
    ChevronLeft,
    ChevronRight,
    Info,
    Package,
    X,
    Save,
    Copy,
    Check,
    Eye,
    RefreshCw,
    ExternalLink
} from 'lucide-react';
import { Card } from '../components/ui/Card';
import { LoadingOverlay } from '../components/ui/LoadingOverlay';
import { Project } from '../types';
import { itemVariants, buttonVariants, iconButtonVariants } from '../utils/constants';
import { cn } from '../utils/helpers';
import { GetProject } from '../../wailsjs/go/main/App';
import { api } from '../../wailsjs/go/models';

interface ProjectsProps {
    projects: Project[];
    loading: boolean;
    onStatusChange?: (status: { type: 'success' | 'error' | 'info'; message: string }) => void;
    onRefresh?: () => Promise<void>;
    onNavigateToEdit?: () => void;
}

const ITEMS_PER_PAGE = 6;
const COPY_FEEDBACK_DURATION = 5000; // ms - increased for better visibility

const STATUS_OPTIONS = [
    { value: 'all', label: 'All Status' },
    { value: 'active', label: 'Active' },
    { value: 'draft', label: 'Draft' },
    { value: 'archived', label: 'Archived' }
];

export const Projects: React.FC<ProjectsProps> = ({
    projects,
    loading,
    onStatusChange,
    onRefresh,
    onNavigateToEdit
}) => {
    const [search, setSearch] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [currentPage, setCurrentPage] = useState(1);
    const [isRefreshing, setIsRefreshing] = useState(false);

    // Auto-refresh on mount
    useEffect(() => {
        if (onRefresh) {
            onRefresh();
        }
    }, []);

    const handleRefresh = async () => {
        if (onRefresh && !isRefreshing) {
            setIsRefreshing(true);
            try {
                await onRefresh();
            } finally {
                setIsRefreshing(false);
            }
        }
    };
    const [deleteConfirm, setDeleteConfirm] = useState<Project | null>(null);
    const [isDeleting, setIsDeleting] = useState(false);
    const [editProject, setEditProject] = useState<Project | null>(null);
    const [editForm, setEditForm] = useState({
        name: '',
        description: '',
        category: '',
        imageUrl: '',
        status: 'draft' as 'draft' | 'active' | 'archived'
    });
    const [copiedObjectId, setCopiedObjectId] = useState<string | null>(null);
    const [showProjectDetails, setShowProjectDetails] = useState<Project | null>(null);
    const [loadingDetails, setLoadingDetails] = useState(false);

    const filteredProjects = useMemo(() => {
        return projects.filter(project => {
            const matchesSearch = project.name.toLowerCase().includes(search.toLowerCase()) ||
                                (project.category?.toLowerCase().includes(search.toLowerCase())) ||
                                (project.objectId?.toLowerCase().includes(search.toLowerCase()));
            const matchesStatus = statusFilter === 'all' || project.status === statusFilter;
            return matchesSearch && matchesStatus;
        });
    }, [projects, search, statusFilter]);

    // Pagination
    const totalPages = Math.ceil(filteredProjects.length / ITEMS_PER_PAGE);
    const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
    const endIndex = startIndex + ITEMS_PER_PAGE;
    const paginatedProjects = filteredProjects.slice(startIndex, endIndex);

    // Reset to page 1 when filters change
    React.useEffect(() => {
        setCurrentPage(1);
    }, [search, statusFilter]);

    // Copy Object ID handler
    const handleCopyObjectId = async (objectId: string) => {
        try {
            await navigator.clipboard.writeText(objectId);
            setCopiedObjectId(objectId);
            setTimeout(() => setCopiedObjectId(null), COPY_FEEDBACK_DURATION);
        } catch (err) {
            console.error('Failed to copy:', err);
        }
    };

    // Fetch full project details with deployment history
    const handleShowDetails = async (project: Project) => {
        if (!project.id) {
            setShowProjectDetails(project);
            return;
        }

        setLoadingDetails(true);
        try {
            const fullProject = await GetProject(project.id);
            if (fullProject) {
                // Convert API deployment records to our frontend format
                const deploymentHistory = fullProject.deployments?.map((d: api.DeploymentRecord) => ({
                    timestamp: d.createdAt,
                    objectId: d.objectId,
                    network: d.network,
                    size: undefined, // Not available in DeploymentRecord
                    status: d.success ? 'success' as const : 'failed' as const,
                    wallet: undefined // Not available in DeploymentRecord
                })) || [];

                setShowProjectDetails({
                    ...project,
                    deploymentHistory
                });
            }
        } catch (err) {
            console.error('Failed to load project details:', err);
            // Show project without deployment history if fetch fails
            setShowProjectDetails(project);
        } finally {
            setLoadingDetails(false);
        }
    };

    const formatDate = (dateString?: string) => {
        if (!dateString || dateString === '0001-01-01 00:00') return 'Never';
        try {
            return new Date(dateString).toLocaleString('en-US', {
                year: 'numeric',
                month: '2-digit',
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit',
                hour12: false
            });
        } catch {
            return 'Unknown';
        }
    };

    const formatRelativeTime = (dateString?: string) => {
        if (!dateString || dateString === '0001-01-01 00:00') return 'Never';
        try {
            const date = new Date(dateString);
            const now = new Date();
            const diffMs = now.getTime() - date.getTime();
            const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
            const diffDays = Math.floor(diffHours / 24);

            if (diffHours < 1) return 'Just now';
            if (diffHours < 24) return `${diffHours} hours ago`;
            if (diffDays === 1) return '1 day ago';
            return `${diffDays} days ago`;
        } catch {
            return 'Unknown';
        }
    };

    const handleDelete = (project: Project) => {
        setDeleteConfirm(project);
    };

    const confirmDelete = async () => {
        if (!deleteConfirm) return;

        setIsDeleting(true);
        try {
            const { DeleteProject } = await import('../../wailsjs/go/main/App');

            const result = await DeleteProject({
                projectId: deleteConfirm.id || 0,
            });

            if (result.success) {
                // Clear from localStorage if this was the selected project
                const selectedProjectStr = localStorage.getItem('selectedProject');
                if (selectedProjectStr) {
                    try {
                        const selectedProject = JSON.parse(selectedProjectStr);
                        if (selectedProject.id === deleteConfirm.id || selectedProject.name === deleteConfirm.name) {
                            localStorage.removeItem('selectedProject');
                        }
                    } catch (err) {
                        console.error('Failed to check selected project:', err);
                    }
                }

                onStatusChange?.({
                    type: 'success',
                    message: result.onChainDestroyed
                        ? `Deleted: ${deleteConfirm.name} (including on-chain destruction)`
                        : `Deleted: ${deleteConfirm.name}`,
                });
                await onRefresh?.();
                setDeleteConfirm(null);
            } else {
                onStatusChange?.({
                    type: 'error',
                    message: `Delete failed: ${result.error}`,
                });
            }
        } catch (err: any) {
            onStatusChange?.({
                type: 'error',
                message: `Delete failed: ${err?.toString()}`,
            });
        } finally {
            setIsDeleting(false);
        }
    };

    const handleEdit = (project: Project) => {
        setEditProject(project);
        setEditForm({
            name: project.name || '',
            description: project.description || '',
            category: project.category || '',
            imageUrl: project.imageUrl || '',
            status: (project.status as 'draft' | 'active' | 'archived') || 'draft'
        });
    };

    const handleSaveEdit = async () => {
        if (!editProject) return;

        try {
            const { EditProject, ArchiveProject } = await import('../../wailsjs/go/main/App');

            // Update project info
            const result = await EditProject({
                projectId: editProject.id || 0,
                name: editForm.name,
                description: editForm.description,
                category: editForm.category,
                imageUrl: editForm.imageUrl,
                suins: editProject.suins || '',
            });

            if (!result.success) {
                onStatusChange?.({
                    type: 'error',
                    message: `Update failed: ${result.error}`,
                });
                return;
            }

            // Handle status change
            if (editForm.status !== editProject.status) {
                if (editForm.status === 'archived') {
                    // Archive the project
                    const archiveResult = await ArchiveProject(editProject.id || 0);
                    if (!archiveResult.success) {
                        onStatusChange?.({
                            type: 'error',
                            message: `Archive failed: ${archiveResult.error}`,
                        });
                        return;
                    }
                } else {
                    // Unarchive: Determine status based on deployments
                    const newStatus = (editProject.deployments && editProject.deployments > 0) ? 'active' : 'draft';

                    try {
                        // Call SetStatus API (needs to be added to backend)
                        const { SetStatus } = await import('../../wailsjs/go/main/App');
                        const statusResult = await SetStatus({
                            projectId: editProject.id || 0,
                            status: newStatus,
                        });

                        if (!statusResult.success) {
                            onStatusChange?.({
                                type: 'error',
                                message: `Unarchive failed: ${statusResult.error}`,
                            });
                            return;
                        }
                    } catch (err: any) {
                        onStatusChange?.({
                            type: 'error',
                            message: `Unarchive failed: SetStatus API not found. Please add it to backend.`,
                        });
                        return;
                    }
                }
            }

            onStatusChange?.({
                type: 'success',
                message: `Project updated: ${editForm.name}`,
            });
            await onRefresh?.();
            setEditProject(null);
        } catch (err: any) {
            onStatusChange?.({
                type: 'error',
                message: `Update failed: ${err?.toString()}`,
            });
        }
    };

    const handleOpenFolder = async (path: string) => {
        if (!path) return;

        try {
            const { OpenInFinder } = await import('../../wailsjs/go/main/App');
            await OpenInFinder(path);
        } catch (err: any) {
            onStatusChange?.({
                type: 'error',
                message: `Failed to open: ${err?.toString()}`,
            });
        }
    };

    const handleOpenProject = (project: Project) => {
        // Save project to localStorage for Edit page to load
        localStorage.setItem('selectedProject', JSON.stringify(project));
        // Navigate to edit page
        if (onNavigateToEdit) {
            onNavigateToEdit();
        }
    };

    const handleOpenSuiscan = async (project: Project) => {
        if (!project.objectId) return;

        const network = project.network || 'testnet';
        const suiscanUrl = `https://suiscan.xyz/${network}/object/${project.objectId}`;

        try {
            const { OpenInBrowser } = await import('../../wailsjs/go/main/App');
            await OpenInBrowser(suiscanUrl);
        } catch (err: any) {
            onStatusChange?.({
                type: 'error',
                message: `Failed to open Suiscan: ${err?.toString()}`,
            });
        }
    };

    const getStatusColor = (status?: string) => {
        switch (status) {
            case 'active':
                return 'text-green-400';
            case 'draft':
                return 'text-yellow-400';
            case 'archived':
                return 'text-zinc-500';
            default:
                return 'text-zinc-400';
        }
    };

    const getNetworkColor = (network?: string) => {
        switch (network) {
            case 'mainnet':
                return 'text-blue-400';
            case 'testnet':
                return 'text-purple-400';
            default:
                return 'text-zinc-500';
        }
    };

    return (
        <motion.div 
            variants={itemVariants}
            initial="hidden"
            animate="show"
            className="space-y-6 max-w-7xl mx-auto"
        >
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-display font-bold text-white mb-2">
                        Projects
                    </h1>
                    <p className="text-zinc-500 text-sm font-mono">
                        Manage your Walgo sites
                    </p>
                </div>
                <div className="flex items-center gap-3">
                    <motion.button
                        onClick={handleRefresh}
                        disabled={isRefreshing || loading}
                        variants={buttonVariants}
                        whileHover="hover"
                        whileTap="tap"
                        className="px-4 py-2 bg-accent/10 hover:bg-accent/20 text-accent border border-accent/30 rounded-sm text-sm font-mono transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                    >
                        <RefreshCw size={14} className={cn(isRefreshing && "animate-spin")} />
                        Refresh
                    </motion.button>
                    <div className="px-4 py-2 bg-zinc-900 border border-zinc-800 rounded-sm">
                        <span className="text-xs font-mono text-zinc-500 uppercase">Total: </span>
                        <span className="text-sm font-mono text-accent font-semibold">{projects.length}</span>
                    </div>
                </div>
            </div>

            {/* Filters */}
            <Card className="border-white/10 bg-zinc-900/20">
                <div className="flex items-center gap-4">
                    {/* Search Bar */}
                    <div className="relative flex-1">
                        <Search size={18} className="absolute left-4 top-1/2 -translate-y-1/2 text-zinc-500" />
                        <input
                            type="text"
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            placeholder="Search by name, category, or object ID..."
                            autoComplete="off"
                            autoCapitalize="off"
                            autoCorrect="off"
                            spellCheck="false"
                            className="w-full pl-12 pr-4 py-3 bg-transparent border-0 text-sm text-white placeholder-zinc-600 focus:outline-none font-mono"
                        />
                    </div>

                    {/* Status Filter */}
                    <div className="relative min-w-[180px]">
                        <ChevronRight size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500 pointer-events-none rotate-90" />
                        <select
                            value={statusFilter}
                            onChange={(e) => setStatusFilter(e.target.value)}
                            className="w-full pl-10 pr-4 py-3 bg-zinc-800 border border-zinc-700 rounded-sm text-sm text-white focus:border-accent focus:outline-none transition-colors appearance-none cursor-pointer font-mono"
                        >
                            {STATUS_OPTIONS.map(option => (
                                <option key={option.value} value={option.value}>
                                    {option.label}
                                </option>
                            ))}
                        </select>
                    </div>
                </div>

                {/* Results Info */}
                {(search || statusFilter !== 'all') && (
                    <div className="mt-3 pt-3 border-t border-white/5 flex items-center justify-between">
                        <p className="text-xs text-zinc-500 font-mono">
                            Found {filteredProjects.length} project{filteredProjects.length !== 1 ? 's' : ''}
                        </p>
                        {(search || statusFilter !== 'all') && (
                            <motion.button
                                onClick={() => {
                                    setSearch('');
                                    setStatusFilter('all');
                                }}
                                className="text-xs text-accent hover:text-accent/80 font-mono transition-colors"
                                variants={buttonVariants}
                                whileHover="hover"
                                whileTap="tap"
                            >
                                Clear filters
                            </motion.button>
                        )}
                    </div>
                )}
            </Card>

            {/* Projects List */}
            {loading ? (
                <div className="flex items-center justify-center py-20">
                    <div className="text-center">
                        <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-accent mb-4"></div>
                        <div className="text-zinc-500 font-mono text-sm">Loading projects...</div>
                    </div>
                </div>
            ) : filteredProjects.length === 0 ? (
                <Card className="border-white/10 bg-zinc-900/20">
                    <div className="flex flex-col items-center justify-center py-16">
                        <FolderOpen size={64} className="text-zinc-700 mb-4" />
                        <h3 className="text-lg font-semibold text-white mb-2">
                            {search ? 'No projects found' : 'No projects yet'}
                        </h3>
                        <p className="text-sm text-zinc-500 font-mono">
                            {search ? 'Try a different search term' : 'Create your first project to get started'}
                        </p>
                    </div>
                </Card>
            ) : (
                <>
                    <Card className="border-white/10 bg-zinc-900/20 p-0 overflow-hidden">
                        <div className="divide-y divide-white/5">
                            <AnimatePresence mode="popLayout">
                                {paginatedProjects.map((project, index) => (
                                    <motion.div
                                        key={project.id || project.path}
                                        initial={{ opacity: 0, x: -20 }}
                                        animate={{ opacity: 1, x: 0 }}
                                        exit={{ opacity: 0, x: 20 }}
                                        transition={{ delay: index * 0.02 }}
                                        className="p-4 hover:bg-white/5 transition-colors group border-b border-white/5 last:border-b-0"
                                    >
                                        {/* Compact Project Row */}
                                        <div className="flex items-center justify-between gap-4">
                                            {/* Left: Name & Basic Info */}
                                            <div className="flex items-center gap-3 flex-1 min-w-0">
                                                <Package size={18} className="text-accent flex-shrink-0" />
                                                <div className="flex items-center gap-3 flex-1 min-w-0">
                                                    <h3 className="text-base font-semibold text-white font-mono truncate">
                                                        {project.name}
                                                    </h3>
                                                    {project.id !== undefined && (
                                                        <span className="text-sm text-zinc-500 font-mono">#{project.id}</span>
                                                    )}
                                                    {project.network && (
                                                        <span className={cn("text-sm font-mono", getNetworkColor(project.network))}>
                                                            {project.network}
                                                        </span>
                                                    )}
                                                    {project.status && (
                                                        <span className={cn("text-sm font-mono", getStatusColor(project.status))}>
                                                            {project.status}
                                                        </span>
                                                    )}
                                                </div>
                                            </div>

                                            {/* Right: Actions */}
                                            <div className="flex items-center gap-1 flex-shrink-0">
                                                <motion.button
                                                    onClick={() => handleShowDetails(project)}
                                                    disabled={loadingDetails}
                                                    className="p-2 hover:bg-white/10 rounded-sm transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                                    title="Show Details"
                                                    variants={iconButtonVariants}
                                                    whileHover="hover"
                                                    whileTap="tap"
                                                >
                                                    <Eye size={16} className="text-zinc-400" />
                                                </motion.button>
                                                <motion.button
                                                    onClick={() => handleEdit(project)}
                                                    className="p-2 hover:bg-white/10 rounded-sm transition-colors"
                                                    title="Edit Info"
                                                    variants={iconButtonVariants}
                                                    whileHover="hover"
                                                    whileTap="tap"
                                                >
                                                    <Info size={16} className="text-zinc-400" />
                                                </motion.button>
                                                <motion.button
                                                    onClick={() => handleOpenFolder(project.path || project.sitePath || '')}
                                                    className="p-2 hover:bg-white/10 rounded-sm transition-colors"
                                                    title="Open Folder"
                                                    variants={iconButtonVariants}
                                                    whileHover="hover"
                                                    whileTap="tap"
                                                >
                                                    <Folder size={16} className="text-zinc-400" />
                                                </motion.button>
                                                {project.objectId && (
                                                    <motion.button
                                                        onClick={() => handleOpenSuiscan(project)}
                                                        className="p-2 hover:bg-blue-500/20 rounded-sm transition-colors"
                                                        title="View on Suiscan"
                                                        variants={iconButtonVariants}
                                                        whileHover="hover"
                                                        whileTap="tap"
                                                    >
                                                        <ExternalLink size={16} className="text-blue-400" />
                                                    </motion.button>
                                                )}
                                                <motion.button
                                                    onClick={() => handleOpenProject(project)}
                                                    className="px-3 py-2 bg-accent/10 hover:bg-accent/20 text-accent rounded-sm transition-colors text-sm font-mono"
                                                    variants={buttonVariants}
                                                    whileHover="hover"
                                                    whileTap="tap"
                                                >
                                                    Open
                                                </motion.button>
                                                <motion.button
                                                    onClick={() => handleDelete(project)}
                                                    className="p-2 hover:bg-red-500/20 rounded-sm transition-colors"
                                                    title="Delete"
                                                    variants={iconButtonVariants}
                                                    whileHover="hover"
                                                    whileTap="tap"
                                                >
                                                    <Trash2 size={16} className="text-red-400" />
                                                </motion.button>
                                            </div>
                                        </div>
                                    </motion.div>
                                ))}
                            </AnimatePresence>
                        </div>
                    </Card>

                    {/* Pagination */}
                    {totalPages > 1 && (
                        <Card className="border-white/10 bg-zinc-900/20 mt-6">
                            <div className="flex items-center justify-between">
                                <div className="text-sm text-zinc-500 font-mono">
                                    Page {currentPage} of {totalPages}
                                    <span className="text-zinc-600 mx-2">•</span>
                                    Showing {startIndex + 1}-{Math.min(endIndex, filteredProjects.length)} of {filteredProjects.length}
                                </div>

                                <div className="flex items-center gap-2">
                                    <motion.button
                                        onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                                        disabled={currentPage === 1}
                                        className={cn(
                                            "p-2 rounded-sm transition-colors",
                                            currentPage === 1
                                                ? "text-zinc-700 cursor-not-allowed"
                                                : "text-zinc-400 hover:text-white hover:bg-white/5"
                                        )}
                                        variants={buttonVariants}
                                        whileHover={currentPage !== 1 ? "hover" : undefined}
                                        whileTap={currentPage !== 1 ? "tap" : undefined}
                                    >
                                        <ChevronLeft size={20} />
                                    </motion.button>

                                    <div className="flex items-center gap-1">
                                        {Array.from({ length: totalPages }, (_, i) => i + 1).map(page => {
                                            const showPage = 
                                                page === 1 || 
                                                page === totalPages || 
                                                Math.abs(page - currentPage) <= 1;
                                            
                                            const showEllipsis = 
                                                (page === 2 && currentPage > 3) ||
                                                (page === totalPages - 1 && currentPage < totalPages - 2);

                                            if (!showPage && !showEllipsis) return null;

                                            if (showEllipsis) {
                                                return (
                                                    <span key={page} className="px-2 text-zinc-600">
                                                        ...
                                                    </span>
                                                );
                                            }

                                            return (
                                                <motion.button
                                                    key={page}
                                                    onClick={() => setCurrentPage(page)}
                                                    className={cn(
                                                        "min-w-[32px] h-8 px-2 rounded-sm text-sm font-mono transition-colors",
                                                        page === currentPage
                                                            ? "bg-accent text-black font-semibold"
                                                            : "text-zinc-400 hover:text-white hover:bg-white/5"
                                                    )}
                                                    variants={buttonVariants}
                                                    whileHover="hover"
                                                    whileTap="tap"
                                                >
                                                    {page}
                                                </motion.button>
                                            );
                                        })}
                                    </div>

                                    <motion.button
                                        onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                                        disabled={currentPage === totalPages}
                                        className={cn(
                                            "p-2 rounded-sm transition-colors",
                                            currentPage === totalPages
                                                ? "text-zinc-700 cursor-not-allowed"
                                                : "text-zinc-400 hover:text-white hover:bg-white/5"
                                        )}
                                        variants={buttonVariants}
                                        whileHover={currentPage !== totalPages ? "hover" : undefined}
                                        whileTap={currentPage !== totalPages ? "tap" : undefined}
                                    >
                                        <ChevronRight size={20} />
                                    </motion.button>
                                </div>
                            </div>
                        </Card>
                    )}
                </>
            )}

            {/* Edit Project Modal */}
            <AnimatePresence>
                {editProject && (
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4"
                        onClick={() => setEditProject(null)}
                    >
                        <motion.div
                            initial={{ scale: 0.9, opacity: 0 }}
                            animate={{ scale: 1, opacity: 1 }}
                            exit={{ scale: 0.9, opacity: 0 }}
                            onClick={(e) => e.stopPropagation()}
                            className="bg-zinc-900 border border-accent/30 rounded-sm p-6 max-w-lg w-full"
                        >
                            <div className="flex items-center justify-between mb-6">
                                <div className="flex items-center gap-3">
                                    <div className="p-2 bg-accent/10 rounded-sm">
                                        <Info size={20} className="text-accent" />
                                    </div>
                                    <div>
                                        <h3 className="text-lg font-semibold text-white">
                                            Edit Project Info
                                        </h3>
                                        <p className="text-xs text-zinc-500 font-mono">
                                            {editProject.path}
                                        </p>
                                    </div>
                                </div>
                                <motion.button
                                    onClick={() => setEditProject(null)}
                                    className="p-1.5 hover:bg-white/5 rounded-sm transition-colors"
                                    variants={iconButtonVariants}
                                    whileHover="hover"
                                    whileTap="tap"
                                >
                                    <X size={20} className="text-zinc-500" />
                                </motion.button>
                            </div>

                            <div className="space-y-4">
                                <div>
                                    <label className="block text-xs font-mono text-zinc-400 uppercase mb-2">
                                        Project Name
                                    </label>
                                    <input
                                        type="text"
                                        value={editForm.name}
                                        onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
                                        className="w-full px-4 py-2 bg-zinc-800 border border-zinc-700 rounded-sm text-sm text-white focus:border-accent focus:outline-none transition-colors font-mono"
                                        autoComplete="off"
                                        autoCapitalize="off"
                                        autoCorrect="off"
                                        spellCheck="false"
                                    />
                                </div>

                                <div>
                                    <label className="block text-xs font-mono text-zinc-400 uppercase mb-2">
                                        Description
                                    </label>
                                    <textarea
                                        value={editForm.description}
                                        onChange={(e) => setEditForm({ ...editForm, description: e.target.value })}
                                        className="w-full px-4 py-2 bg-zinc-800 border border-zinc-700 rounded-sm text-sm text-white focus:border-accent focus:outline-none transition-colors resize-none font-mono"
                                        rows={3}
                                        autoComplete="off"
                                        autoCapitalize="off"
                                        autoCorrect="off"
                                        spellCheck="false"
                                    />
                                </div>

                                <div>
                                    <label className="block text-xs font-mono text-zinc-400 uppercase mb-2">
                                        Category
                                    </label>
                                    <input
                                        type="text"
                                        value={editForm.category}
                                        onChange={(e) => setEditForm({ ...editForm, category: e.target.value })}
                                        className="w-full px-4 py-2 bg-zinc-800 border border-zinc-700 rounded-sm text-sm text-white focus:border-accent focus:outline-none transition-colors font-mono"
                                        placeholder="e.g., blog, website, portfolio"
                                        autoComplete="off"
                                        autoCapitalize="off"
                                        autoCorrect="off"
                                        spellCheck="false"
                                    />
                                </div>

                                <div>
                                    <label className="block text-xs font-mono text-zinc-400 uppercase mb-2">
                                        Image URL
                                    </label>
                                    <input
                                        type="text"
                                        value={editForm.imageUrl}
                                        onChange={(e) => setEditForm({ ...editForm, imageUrl: e.target.value })}
                                        className="w-full px-4 py-2 bg-zinc-800 border border-zinc-700 rounded-sm text-sm text-white focus:border-accent focus:outline-none transition-colors font-mono"
                                        placeholder="https://example.com/image.png"
                                        autoComplete="off"
                                        autoCapitalize="off"
                                        autoCorrect="off"
                                        spellCheck="false"
                                    />
                                </div>

                                <div className="flex items-center gap-3 p-4 bg-zinc-800/50 border border-zinc-700 rounded-sm">
                                    <input
                                        type="checkbox"
                                        id="archive-checkbox"
                                        checked={editForm.status === 'archived'}
                                        onChange={(e) => setEditForm({
                                            ...editForm,
                                            status: e.target.checked ? 'archived' : (editProject?.deployments && editProject.deployments > 0 ? 'active' : 'draft')
                                        })}
                                        className="w-4 h-4 bg-zinc-900 border-zinc-600 rounded text-accent focus:ring-accent focus:ring-2"
                                    />
                                    <label htmlFor="archive-checkbox" className="text-sm font-mono text-zinc-300 cursor-pointer select-none">
                                        Archive this project
                                    </label>
                                </div>
                            </div>

                            <div className="flex gap-3 mt-6">
                                <motion.button
                                    onClick={() => setEditProject(null)}
                                    className="flex-1 px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-white rounded-sm text-sm font-semibold transition-colors"
                                    variants={buttonVariants}
                                    whileHover="hover"
                                    whileTap="tap"
                                >
                                    Cancel
                                </motion.button>
                                <motion.button
                                    onClick={handleSaveEdit}
                                    className="flex-1 px-4 py-2 bg-accent hover:bg-accent/90 text-black rounded-sm text-sm font-semibold transition-colors flex items-center justify-center gap-2"
                                    variants={buttonVariants}
                                    whileHover="hover"
                                    whileTap="tap"
                                >
                                    <Save size={16} />
                                    Save Changes
                                </motion.button>
                            </div>
                        </motion.div>
                    </motion.div>
                )}
            </AnimatePresence>

            {/* Delete Confirmation Modal */}
            <AnimatePresence>
                {deleteConfirm && (
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4"
                        onClick={() => !isDeleting && setDeleteConfirm(null)}
                    >
                        <motion.div
                            initial={{ scale: 0.9, opacity: 0 }}
                            animate={{ scale: 1, opacity: 1 }}
                            exit={{ scale: 0.9, opacity: 0 }}
                            onClick={(e) => e.stopPropagation()}
                            className="bg-zinc-900 border border-red-500/30 rounded-sm p-6 max-w-md w-full relative"
                        >
                            {isDeleting && (
                                <LoadingOverlay message="Deleting project..." />
                            )}
                            <div className="flex items-start gap-4 mb-6">
                                <div className="p-2 bg-red-500/10 rounded-sm">
                                    <AlertCircle size={24} className="text-red-400" />
                                </div>
                                <div className="flex-1">
                                    <h3 className="text-lg font-semibold text-white mb-2">
                                        Delete Project?
                                    </h3>
                                    <p className="text-sm text-zinc-400 mb-2">
                                        Are you sure you want to delete <span className="font-mono text-white">{deleteConfirm.name}</span>?
                                    </p>
                                    <p className="text-xs text-red-400 font-mono">
                                        This action cannot be undone.
                                    </p>
                                </div>
                            </div>

                            <div className="flex gap-3">
                                <motion.button
                                    onClick={() => setDeleteConfirm(null)}
                                    disabled={isDeleting}
                                    className="flex-1 px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-white rounded-sm text-sm font-semibold transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                    variants={buttonVariants}
                                    whileHover="hover"
                                    whileTap="tap"
                                >
                                    Cancel
                                </motion.button>
                                <motion.button
                                    onClick={confirmDelete}
                                    disabled={isDeleting}
                                    className="flex-1 px-4 py-2 bg-red-500 hover:bg-red-600 text-white rounded-sm text-sm font-semibold transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                    variants={buttonVariants}
                                    whileHover="hover"
                                    whileTap="tap"
                                >
                                    Delete
                                </motion.button>
                            </div>
                        </motion.div>
                    </motion.div>
                )}
            </AnimatePresence>

            {/* Project Details Modal (walgo show) */}
            <AnimatePresence>
                {showProjectDetails && (
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4"
                        onClick={() => setShowProjectDetails(null)}
                    >
                        <motion.div
                            initial={{ scale: 0.9, opacity: 0 }}
                            animate={{ scale: 1, opacity: 1 }}
                            exit={{ scale: 0.9, opacity: 0 }}
                            onClick={(e) => e.stopPropagation()}
                            className="bg-zinc-900 border border-white/10 rounded-sm p-6 max-w-3xl w-full max-h-[80vh] overflow-y-auto"
                        >
                            {/* Header */}
                            <div className="flex items-start justify-between mb-6">
                                <div>
                                    <h2 className="text-2xl font-display text-white mb-2">
                                        {showProjectDetails.name}
                                    </h2>
                                    <p className="text-sm text-zinc-500 font-mono">
                                        Project details
                                        {loadingDetails && <span className="ml-2 text-accent">Loading...</span>}
                                    </p>
                                </div>
                                <motion.button
                                    onClick={() => setShowProjectDetails(null)}
                                    className="p-2 hover:bg-white/10 rounded-sm transition-colors"
                                    variants={iconButtonVariants}
                                    whileHover="hover"
                                    whileTap="tap"
                                >
                                    <X size={20} className="text-zinc-400" />
                                </motion.button>
                            </div>

                            {/* Project Information Grid */}
                            <div className="space-y-4">
                                {/* Basic Info */}
                                <div className="bg-zinc-800/50 rounded-sm p-4 border border-white/5">
                                    <h3 className="text-sm font-mono text-accent mb-3 uppercase">Basic Information</h3>
                                    <div className="grid grid-cols-2 gap-3 text-sm font-mono">
                                        {showProjectDetails.id !== undefined && (
                                            <div className="flex">
                                                <span className="text-zinc-500 w-32">ID:</span>
                                                <span className="text-white">{showProjectDetails.id}</span>
                                            </div>
                                        )}
                                        {showProjectDetails.name && (
                                            <div className="flex">
                                                <span className="text-zinc-500 w-32">Name:</span>
                                                <span className="text-white">{showProjectDetails.name}</span>
                                            </div>
                                        )}
                                        {showProjectDetails.category && (
                                            <div className="flex">
                                                <span className="text-zinc-500 w-32">Category:</span>
                                                <span className="text-white">{showProjectDetails.category}</span>
                                            </div>
                                        )}
                                        {showProjectDetails.status && (
                                            <div className="flex">
                                                <span className="text-zinc-500 w-32">Status:</span>
                                                <span className={getStatusColor(showProjectDetails.status)}>{showProjectDetails.status}</span>
                                            </div>
                                        )}
                                    </div>
                                </div>

                                {/* Network & Deployment */}
                                <div className="bg-zinc-800/50 rounded-sm p-4 border border-white/5">
                                    <h3 className="text-sm font-mono text-accent mb-3 uppercase">Network & Deployment</h3>
                                    <div className="grid grid-cols-2 gap-3 text-sm font-mono">
                                        {showProjectDetails.network && (
                                            <div className="flex">
                                                <span className="text-zinc-500 w-32">Network:</span>
                                                <span className={getNetworkColor(showProjectDetails.network)}>{showProjectDetails.network}</span>
                                            </div>
                                        )}
                                        {showProjectDetails.deployments !== undefined && (
                                            <div className="flex">
                                                <span className="text-zinc-500 w-32">Deployments:</span>
                                                <span className="text-white">{showProjectDetails.deployments}</span>
                                            </div>
                                        )}
                                        {showProjectDetails.lastDeploy && (
                                            <div className="flex col-span-2">
                                                <span className="text-zinc-500 w-32">Last Deploy:</span>
                                                <span className="text-zinc-400">{formatDate(showProjectDetails.lastDeploy)}</span>
                                            </div>
                                        )}
                                        {showProjectDetails.deployedAt && (
                                            <div className="flex col-span-2">
                                                <span className="text-zinc-500 w-32">Deployed At:</span>
                                                <span className="text-zinc-400">{formatDate(showProjectDetails.deployedAt)}</span>
                                            </div>
                                        )}
                                    </div>
                                </div>

                                {/* Object ID */}
                                {showProjectDetails.objectId && (
                                    <div className="bg-zinc-800/50 rounded-sm p-4 border border-white/5">
                                        <h3 className="text-sm font-mono text-accent mb-3 uppercase">Walrus Object</h3>
                                        <div className="flex items-center gap-2">
                                            <span className="text-zinc-500 text-sm font-mono w-32">Object ID:</span>
                                            <code className="flex-1 text-xs text-zinc-400 bg-black/30 px-3 py-2 rounded-sm border border-white/5 break-all">
                                                {showProjectDetails.objectId}
                                            </code>
                                            <motion.button
                                                onClick={() => handleCopyObjectId(showProjectDetails.objectId!)}
                                                className="p-2 rounded-sm text-zinc-400 hover:text-accent hover:bg-white/5"
                                                variants={iconButtonVariants}
                                                whileHover="hover"
                                                whileTap="tap"
                                                title="Copy Object ID"
                                            >
                                                {copiedObjectId === showProjectDetails.objectId ? (
                                                    <Check size={16} className="text-green-400" />
                                                ) : (
                                                    <Copy size={16} />
                                                )}
                                            </motion.button>
                                        </div>
                                    </div>
                                )}

                                {/* URLs */}
                                {(showProjectDetails.url || showProjectDetails.suins) && (
                                    <div className="bg-zinc-800/50 rounded-sm p-4 border border-white/5">
                                        <h3 className="text-sm font-mono text-accent mb-3 uppercase">URLs</h3>
                                        <div className="space-y-2 text-sm font-mono">
                                            {showProjectDetails.url && (
                                                <div className="flex">
                                                    <span className="text-zinc-500 w-32">Live URL:</span>
                                                    <a href={showProjectDetails.url} target="_blank" rel="noopener noreferrer" className="text-accent hover:underline break-all">
                                                        {showProjectDetails.url}
                                                    </a>
                                                </div>
                                            )}
                                            {showProjectDetails.suins && (
                                                <div className="flex">
                                                    <span className="text-zinc-500 w-32">SuiNS:</span>
                                                    <span className="text-white">{showProjectDetails.suins}</span>
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                )}

                                {/* Deployment History */}
                                {showProjectDetails.deploymentHistory && showProjectDetails.deploymentHistory.length > 0 && (
                                    <div className="bg-zinc-800/50 rounded-sm p-4 border border-white/5">
                                        <h3 className="text-sm font-mono text-accent mb-3 uppercase">Deployment History</h3>
                                        <div className="space-y-3">
                                            {showProjectDetails.deploymentHistory.map((deployment, idx) => (
                                                <div key={idx} className="bg-black/30 rounded-sm p-3 border border-white/5">
                                                    <div className="flex items-start justify-between mb-2">
                                                        <div className="flex items-center gap-2">
                                                            <span className={cn(
                                                                "text-xs font-mono px-2 py-0.5 rounded-sm",
                                                                deployment.status === 'success' 
                                                                    ? "bg-green-500/20 text-green-400" 
                                                                    : "bg-red-500/20 text-red-400"
                                                            )}>
                                                                {deployment.status}
                                                            </span>
                                                            <span className={cn("text-xs font-mono", getNetworkColor(deployment.network))}>
                                                                {deployment.network}
                                                            </span>
                                                        </div>
                                                        <span className="text-xs text-zinc-500 font-mono">
                                                            {formatDate(deployment.timestamp)}
                                                        </span>
                                                    </div>
                                                    <div className="space-y-1 text-xs font-mono">
                                                        <div className="flex items-center gap-2">
                                                            <span className="text-zinc-500">Object ID:</span>
                                                            <code className="text-zinc-400 bg-black/30 px-2 py-0.5 rounded-sm flex-1 truncate" title={deployment.objectId}>
                                                                {deployment.objectId}
                                                            </code>
                                                            <motion.button
                                                                onClick={() => handleCopyObjectId(deployment.objectId)}
                                                                className="p-1 rounded-sm text-zinc-400 hover:text-accent hover:bg-white/5"
                                                                variants={iconButtonVariants}
                                                                whileHover="hover"
                                                                whileTap="tap"
                                                                title="Copy Object ID"
                                                            >
                                                                {copiedObjectId === deployment.objectId ? (
                                                                    <Check size={12} className="text-green-400" />
                                                                ) : (
                                                                    <Copy size={12} />
                                                                )}
                                                            </motion.button>
                                                        </div>
                                                        {deployment.size !== undefined && (
                                                            <div className="flex">
                                                                <span className="text-zinc-500">Size:</span>
                                                                <span className="text-zinc-400 ml-2">{(deployment.size / 1024).toFixed(2)} KB</span>
                                                            </div>
                                                        )}
                                                        {deployment.wallet && (
                                                            <div className="flex">
                                                                <span className="text-zinc-500">Wallet:</span>
                                                                <span className="text-zinc-400 ml-2 truncate" title={deployment.wallet}>
                                                                    {deployment.wallet.slice(0, 8)}...{deployment.wallet.slice(-6)}
                                                                </span>
                                                            </div>
                                                        )}
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                )}

                                {/* File Info */}
                                {(showProjectDetails.size !== undefined || showProjectDetails.fileCount !== undefined) && (
                                    <div className="bg-zinc-800/50 rounded-sm p-4 border border-white/5">
                                        <h3 className="text-sm font-mono text-accent mb-3 uppercase">File Information</h3>
                                        <div className="grid grid-cols-2 gap-3 text-sm font-mono">
                                            {showProjectDetails.size !== undefined && (
                                                <div className="flex">
                                                    <span className="text-zinc-500 w-32">Size:</span>
                                                    <span className="text-white">{(showProjectDetails.size / 1024).toFixed(2)} KB</span>
                                                </div>
                                            )}
                                            {showProjectDetails.fileCount !== undefined && (
                                                <div className="flex">
                                                    <span className="text-zinc-500 w-32">Files:</span>
                                                    <span className="text-white">{showProjectDetails.fileCount}</span>
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                )}

                                {/* Timestamps */}
                                <div className="bg-zinc-800/50 rounded-sm p-4 border border-white/5">
                                    <h3 className="text-sm font-mono text-accent mb-3 uppercase">Timestamps</h3>
                                    <div className="grid grid-cols-2 gap-3 text-sm font-mono">
                                        {showProjectDetails.createdAt && (
                                            <div className="flex">
                                                <span className="text-zinc-500 w-32">Created:</span>
                                                <span className="text-zinc-400">{formatDate(showProjectDetails.createdAt)}</span>
                                            </div>
                                        )}
                                        {showProjectDetails.updatedAt && (
                                            <div className="flex">
                                                <span className="text-zinc-500 w-32">Updated:</span>
                                                <span className="text-zinc-400">{formatRelativeTime(showProjectDetails.updatedAt)}</span>
                                            </div>
                                        )}
                                    </div>
                                </div>

                                {/* Description */}
                                {showProjectDetails.description && (
                                    <div className="bg-zinc-800/50 rounded-sm p-4 border border-white/5">
                                        <h3 className="text-sm font-mono text-accent mb-3 uppercase">Description</h3>
                                        <p className="text-sm text-zinc-400">{showProjectDetails.description}</p>
                                    </div>
                                )}

                                {/* Path */}
                                {showProjectDetails.path && (
                                    <div className="bg-zinc-800/50 rounded-sm p-4 border border-white/5">
                                        <h3 className="text-sm font-mono text-accent mb-3 uppercase">Path</h3>
                                        <code className="text-xs text-zinc-400 bg-black/30 px-3 py-2 rounded-sm border border-white/5 block break-all">
                                            {showProjectDetails.path}
                                        </code>
                                    </div>
                                )}
                            </div>

                            {/* Footer Actions */}
                            <div className="flex gap-3 mt-6 pt-6 border-t border-white/10">
                                {(showProjectDetails.path || showProjectDetails.sitePath) && (
                                    <motion.button
                                        onClick={() => {
                                            handleOpenFolder(showProjectDetails.path || showProjectDetails.sitePath || '');
                                            setShowProjectDetails(null);
                                        }}
                                        className="flex-1 px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-white rounded-sm text-sm font-semibold transition-colors"
                                        variants={buttonVariants}
                                        whileHover="hover"
                                        whileTap="tap"
                                    >
                                        Open Folder
                                    </motion.button>
                                )}
                                {showProjectDetails.objectId && (
                                    <motion.button
                                        onClick={() => {
                                            handleOpenSuiscan(showProjectDetails);
                                            setShowProjectDetails(null);
                                        }}
                                        className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-sm text-sm font-semibold transition-colors flex items-center justify-center gap-2"
                                        variants={buttonVariants}
                                        whileHover="hover"
                                        whileTap="tap"
                                    >
                                        <ExternalLink size={16} />
                                        View on Suiscan
                                    </motion.button>
                                )}
                            </div>
                        </motion.div>
                    </motion.div>
                )}
            </AnimatePresence>
        </motion.div>
    );
};
