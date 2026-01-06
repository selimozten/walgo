import { useState, useEffect } from 'react';
import {
    CreateSite, BuildSite, DeploySite, ConfigureAI, GenerateContent, UpdateContent,
    ListProjects, ImportObsidian, OptimizeSite, CompressSite,
    QuickStart, Serve, NewContent, UpdateDeployment, Doctor, Status,
    EditProject, ArchiveProject, LaunchWizard, DeleteProject, GetProject,
    GetAIConfig, UpdateAIConfig as UpdateAIConfigAPI, CleanAIConfig,
    SelectDirectory, AICreateSite, GetVersion, CheckSetupDeps,
    OpenInBrowser, OpenInFinder, StopServe, IsServing, GetServerURL
} from '../wailsjs/go/main/App';
import { WindowMinimise, WindowToggleMaximise, Quit } from '../wailsjs/runtime/runtime';
import {
    LayoutGrid,
    Plus,
    Settings,
    Globe,
    Terminal,
    Check,
    AlertCircle,
    Loader2,
    Box,
    ArrowUpRight,
    Folder,
    ChevronRight,
    Command,
    Cpu,
    HardDrive,
    Activity,
    Zap,
    Hexagon,
    Radio,
    Minus,
    Square,
    X,
    Play,
    RefreshCw,
    Sparkles,
    Stethoscope,
    FileText,
    Archive,
    Rocket,
    FileCode,
    Database,
    Search,
    Save,
    Trash2,
    Key,
    Sliders,
    Wand2,
    Users,
    Briefcase,
    BookOpen,
    Palette,
    ExternalLink,
    Edit3,
    Eye,
    StopCircle,
    FolderOpen
} from 'lucide-react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { motion, AnimatePresence } from 'framer-motion';

function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

function App() {
    const [activeTab, setActiveTab] = useState('dashboard');
    const [siteName, setSiteName] = useState('');
    const [sitePath, setSitePath] = useState('');
    const [status, setStatus] = useState<{ type: 'success' | 'error' | 'info', message: string } | null>(null);
    const [isProcessing, setIsProcessing] = useState(false);
    const [logs, setLogs] = useState<string[]>([]);
    const [projects, setProjects] = useState<any[]>([]);
    const [aiConfigured, setAiConfigured] = useState(false);
    const [aiConfig, setAiConfig] = useState<any>(null);
    const [aiProvider, setAiProvider] = useState('');
    const [aiApiKey, setAiApiKey] = useState('');
    const [aiBaseUrl, setAiBaseUrl] = useState('');
    const [aiModel, setAiModel] = useState('');
    const [parentDir, setParentDir] = useState('');
    const [vaultPath, setVaultPath] = useState('');
    const [includeDrafts, setIncludeDrafts] = useState(false);
    const [contentSlug, setContentSlug] = useState('');
    const [aiTopic, setAiTopic] = useState('');
    const [aiContext, setAiContext] = useState('');
    const [aiContentType, setAiContentType] = useState('post');
    const [serverRunning, setServerRunning] = useState(false);
    const [serverUrl, setServerUrl] = useState('');
    const [version, setVersion] = useState('');
    // AI Create Site wizard
    const [aiSiteType, setAiSiteType] = useState('blog');
    const [aiSiteDescription, setAiSiteDescription] = useState('');
    const [aiSiteAudience, setAiSiteAudience] = useState('');
    const [aiSiteFeatures, setAiSiteFeatures] = useState('');
    const [aiSiteTheme, setAiSiteTheme] = useState('ananke');
    // Project management
    const [selectedProject, setSelectedProject] = useState<any>(null);
    const [editProjectName, setEditProjectName] = useState('');
    const [editProjectCategory, setEditProjectCategory] = useState('');
    const [editProjectDescription, setEditProjectDescription] = useState('');
    const [editProjectSuins, setEditProjectSuins] = useState('');

    useEffect(() => {
        setStatus(null);
        loadProjects();
        loadAIConfig();
        loadVersion();
        checkServerStatus();
    }, [activeTab]);

    const loadVersion = async () => {
        try {
            const result = await GetVersion();
            if (result.version) {
                setVersion(result.version);
            }
        } catch (err) {
            console.error('Failed to load version:', err);
        }
    };

    const checkServerStatus = async () => {
        try {
            const running = await IsServing();
            setServerRunning(running);
            if (running) {
                const url = await GetServerURL();
                setServerUrl(url);
            }
        } catch (err) {
            console.error('Failed to check server status:', err);
        }
    };

    const loadProjects = async () => {
        try {
            const result = await ListProjects();
            if (result) {
                setProjects(result);
            }
        } catch (err) {
            console.error('Failed to load projects:', err);
        }
    };

    const loadAIConfig = async () => {
        try {
            const result = await GetAIConfig();
            if (result.success) {
                setAiConfigured(result.enabled);
                setAiConfig(result);
                if (result.currentProvider) {
                    setAiProvider(result.currentProvider);
                    setAiModel(result.currentModel || '');
                }
            }
        } catch (err) {
            console.error('Failed to load AI config:', err);
        }
    };

    const handleGetAIConfig = async () => {
        loadAIConfig();
    };

    const handleUpdateAIConfig = async () => {
        if (!aiProvider || !aiApiKey) {
            setStatus({ type: 'error', message: 'PROVIDER_AND_API_KEY_REQUIRED' });
            addLog('Error: Provider and API key are required');
            return;
        }
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'UPDATING_AI_CONFIG...' });
        addLog('Updating AI configuration...');
        try {
            await UpdateAIConfigAPI({
                provider: aiProvider,
                apiKey: aiApiKey,
                baseURL: aiBaseUrl,
                model: aiModel
            });
            setStatus({ type: 'success', message: 'AI_CONFIG_UPDATED' });
            addLog('AI configuration updated successfully');
            loadAIConfig();
        } catch (err: any) {
            setStatus({ type: 'error', message: `AI_CONFIG_FAILED: ${err}` });
            addLog(`Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleCleanAIConfig = async () => {
        if (!confirm('Are you sure you want to clear all AI credentials?')) {
            return;
        }
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'CLEANING_AI_CONFIG...' });
        addLog('Cleaning AI credentials...');
        try {
            await CleanAIConfig();
            setStatus({ type: 'success', message: 'AI_CONFIG_CLEARED' });
            addLog('AI credentials cleared');
            setAiProvider('');
            setAiApiKey('');
            setAiBaseUrl('');
            setAiModel('');
            setAiConfigured(false);
            setAiConfig(null);
        } catch (err: any) {
            setStatus({ type: 'error', message: `CLEAN_FAILED: ${err}` });
            addLog(`Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const addLog = (msg: string) => {
        setLogs(prev => [...prev, `[${new Date().toLocaleTimeString()}] ${msg}`]);
    };

    const handleSelectParentDir = async () => {
        try {
            const dir = await SelectDirectory('Select Parent Directory');
            if (dir) {
                setParentDir(dir);
                addLog(`Selected directory: ${dir}`);
            }
        } catch (err: any) {
            addLog(`Error selecting directory: ${err}`);
        }
    };

    const handleSelectVaultPath = async () => {
        try {
            const dir = await SelectDirectory('Select Obsidian Vault');
            if (dir) {
                setVaultPath(dir);
                addLog(`Selected vault: ${dir}`);
            }
        } catch (err: any) {
            addLog(`Error selecting vault: ${err}`);
        }
    };

    const handleSelectSitePath = async () => {
        try {
            const dir = await SelectDirectory('Select Hugo Site');
            if (dir) {
                setSitePath(dir);
                addLog(`Selected site: ${dir}`);
            }
        } catch (err: any) {
            addLog(`Error selecting site: ${err}`);
        }
    };

    const handleQuickStart = async () => {
        if (!siteName || !parentDir) {
            setStatus({ type: 'error', message: 'PARENT_DIR_AND_NAME_REQUIRED' });
            addLog('Error: Please select a parent directory and enter a site name');
            return;
        }
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'INITIALIZING_QUICKSTART...' });
        addLog(`Starting quickstart: ${siteName} in ${parentDir}`);
        try {
            const result = await QuickStart({
                parentDir: parentDir,
                name: siteName,
                skipBuild: false
            });
            if (result.success) {
                setStatus({ type: 'success', message: `QUICKSTART_COMPLETE: ${siteName}` });
                addLog(`Success: Site created at ${result.sitePath}`);
                setSitePath(result.sitePath);
                setTimeout(() => setActiveTab('manage'), 1000);
            } else {
                setStatus({ type: 'error', message: `QUICKSTART_FAILED: ${result.error}` });
                addLog(`Error: ${result.error}`);
            }
        } catch (err: any) {
            setStatus({ type: 'error', message: `SYSTEM_ERROR: ${err}` });
            addLog(`System Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleCreate = async () => {
        if (!siteName || !parentDir) {
            setStatus({ type: 'error', message: 'PARENT_DIR_AND_NAME_REQUIRED' });
            addLog('Error: Please select a parent directory and enter a site name');
            return;
        }
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'INITIALIZING_SEQUENCE...' });
        addLog(`Initializing project: ${siteName} in ${parentDir}`);
        try {
            await CreateSite(parentDir, siteName);
            setStatus({ type: 'success', message: `PROJECT_CREATED: ${siteName}` });
            addLog(`Success: Project created at ${parentDir}/${siteName}`);
            setSitePath(`${parentDir}/${siteName}`);
            setTimeout(() => setActiveTab('manage'), 1000);
        } catch (err: any) {
            setStatus({ type: 'error', message: `SEQUENCE_FAILED: ${err}` });
            addLog(`Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleBuild = async () => {
        if (!sitePath) return;
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'COMPILING_ASSETS...' });
        addLog(`Starting build for ${sitePath}`);
        try {
            await BuildSite(sitePath);
            setStatus({ type: 'success', message: 'BUILD_COMPLETE' });
            addLog('Build completed successfully');
        } catch (err: any) {
            setStatus({ type: 'error', message: `BUILD_FAILED: ${err}` });
            addLog(`Build Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleDeploy = async () => {
        if (!sitePath) return;
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'UPLOADING_TO_WALRUS...' });
        addLog('Initiating deployment to Walrus...');
        try {
            const result = await DeploySite(sitePath, 1);
            if (result.success) {
                setStatus({ type: 'success', message: `DEPLOYED: ${result.objectId}` });
                addLog(`Deployment Success: ${result.objectId}`);
                loadProjects();
            } else {
                setStatus({ type: 'error', message: `DEPLOY_FAILED: ${result.error}` });
                addLog(`Deployment Failed: ${result.error}`);
            }
        } catch (err: any) {
            setStatus({ type: 'error', message: `SYSTEM_ERROR: ${err}` });
            addLog(`System Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleServe = async () => {
        if (!sitePath) return;
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'STARTING_DEV_SERVER...' });
        addLog('Starting Hugo development server...');
        try {
            const result = await Serve({
                sitePath,
                port: 1313,
                drafts: true,
                expired: false,
                future: false
            });
            if (result.success) {
                setStatus({ type: 'success', message: `SERVER_RUNNING: ${result.url}` });
                addLog(`Dev server running at: ${result.url}`);
                setServerRunning(true);
                setServerUrl(result.url);
            } else {
                setStatus({ type: 'error', message: `SERVER_FAILED: ${result.error}` });
                addLog(`Server Error: ${result.error}`);
            }
        } catch (err: any) {
            setStatus({ type: 'error', message: `SYSTEM_ERROR: ${err}` });
            addLog(`System Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleUpdate = async () => {
        if (!sitePath) return;
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'UPDATING_SITE...' });
        addLog('Updating site on Walrus...');
        try {
            const result = await UpdateDeployment({
                sitePath,
                objectId: '',
                epochs: 1
            });
            if (result.success) {
                setStatus({ type: 'success', message: `UPDATED: ${result.objectId}` });
                addLog(`Site updated: ${result.objectId}`);
                loadProjects();
            } else {
                setStatus({ type: 'error', message: `UPDATE_FAILED: ${result.error}` });
                addLog(`Update Error: ${result.error}`);
            }
        } catch (err: any) {
            setStatus({ type: 'error', message: `SYSTEM_ERROR: ${err}` });
            addLog(`System Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleDoctor = async () => {
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'RUNNING_DIAGNOSTICS...' });
        addLog('Running environment diagnostics...');
        try {
            const result = await Doctor();
            if (result.success) {
                setStatus({ type: 'success', message: 'DIAGNOSTICS_COMPLETE' });
                addLog(`Diagnostics complete: ${result.summary.issues} issues, ${result.summary.warnings} warnings`);
                result.checks.forEach((check: any) => {
                    addLog(`  ${check.name}: ${check.status} - ${check.message}`);
                });
            } else {
                setStatus({ type: 'error', message: `DIAGNOSTICS_FAILED: ${result.error}` });
                addLog(`Diagnostics Error: ${result.error}`);
            }
        } catch (err: any) {
            setStatus({ type: 'error', message: `SYSTEM_ERROR: ${err}` });
            addLog(`System Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleLaunch = async () => {
        if (!sitePath) return;
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'LAUNCHING_WIZARD...' });
        addLog('Starting launch wizard...');
        try {
            const result = await LaunchWizard({
                sitePath,
                network: 'testnet',
                projectName: siteName || 'My Site',
                category: 'website',
                description: 'Site created with Walgo Desktop',
                epochs: 1
            });
            if (result.success) {
                setStatus({ type: 'success', message: `DEPLOYED: ${result.objectId}` });
                addLog(`Launch complete: ${result.objectId}`);
                result.steps.forEach((step: any) => {
                    addLog(`  ${step.name}: ${step.status}`);
                });
                loadProjects();
            } else {
                setStatus({ type: 'error', message: `LAUNCH_FAILED: ${result.error}` });
                addLog(`Launch Error: ${result.error}`);
            }
        } catch (err: any) {
            setStatus({ type: 'error', message: `SYSTEM_ERROR: ${err}` });
            addLog(`System Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleImportObsidian = async () => {
        if (!sitePath || !vaultPath) {
            setStatus({ type: 'error', message: 'SITE_AND_VAULT_REQUIRED' });
            addLog('Error: Please select both a Hugo site and an Obsidian vault');
            return;
        }
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'IMPORTING_VAULT...' });
        addLog(`Importing from: ${vaultPath}`);
        try {
            const result = await ImportObsidian({
                sitePath,
                vaultPath,
                includeDrafts,
                attachmentDir: 'attachments'
            });
            if (result.success) {
                setStatus({ type: 'success', message: `IMPORTED: ${result.filesImported} files` });
                addLog(`Import complete: ${result.filesImported} files imported`);
            } else {
                setStatus({ type: 'error', message: `IMPORT_FAILED: ${result.error}` });
                addLog(`Import Error: ${result.error}`);
            }
        } catch (err: any) {
            setStatus({ type: 'error', message: `SYSTEM_ERROR: ${err}` });
            addLog(`System Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleAIGenerate = async () => {
        if (!sitePath || !aiTopic) {
            setStatus({ type: 'error', message: 'SITE_AND_TOPIC_REQUIRED' });
            addLog('Error: Please select a site and enter a topic');
            return;
        }
        if (!aiConfigured) {
            setStatus({ type: 'error', message: 'AI_NOT_CONFIGURED' });
            addLog('Error: Please configure AI credentials first');
            return;
        }
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'GENERATING_CONTENT...' });
        addLog(`Generating ${aiContentType}: ${aiTopic}`);
        try {
            const result = await GenerateContent({
                sitePath,
                contentType: aiContentType,
                topic: aiTopic,
                context: aiContext
            });
            if (result.success) {
                setStatus({ type: 'success', message: `GENERATED: ${result.filePath}` });
                addLog(`Content generated: ${result.filePath}`);
            } else {
                setStatus({ type: 'error', message: `GENERATE_FAILED: ${result.error}` });
                addLog(`Generate Error: ${result.error}`);
            }
        } catch (err: any) {
            setStatus({ type: 'error', message: `SYSTEM_ERROR: ${err}` });
            addLog(`System Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleCreateNewContent = async () => {
        if (!sitePath || !contentSlug) {
            setStatus({ type: 'error', message: 'SITE_AND_SLUG_REQUIRED' });
            addLog('Error: Please select a site and enter a content slug');
            return;
        }
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'CREATING_CONTENT...' });
        addLog(`Creating new content: ${contentSlug}`);
        try {
            const result = await NewContent({
                sitePath,
                slug: contentSlug,
                contentType: ''
            });
            if (result.success) {
                setStatus({ type: 'success', message: `CONTENT_CREATED: ${result.filePath}` });
                addLog(`Content created: ${result.filePath}`);
                setContentSlug('');
            } else {
                setStatus({ type: 'error', message: `CONTENT_FAILED: ${result.error}` });
                addLog(`Content Error: ${result.error}`);
            }
        } catch (err: any) {
            setStatus({ type: 'error', message: `SYSTEM_ERROR: ${err}` });
            addLog(`System Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleStopServe = async () => {
        try {
            const stopped = await StopServe();
            if (stopped) {
                setServerRunning(false);
                setServerUrl('');
                setStatus({ type: 'success', message: 'SERVER_STOPPED' });
                addLog('Development server stopped');
            }
        } catch (err: any) {
            addLog(`Error stopping server: ${err}`);
        }
    };

    const handleOpenInBrowser = async (url: string) => {
        try {
            await OpenInBrowser(url);
        } catch (err: any) {
            addLog(`Error opening browser: ${err}`);
        }
    };

    const handleOpenInFinder = async (path: string) => {
        try {
            await OpenInFinder(path);
        } catch (err: any) {
            addLog(`Error opening finder: ${err}`);
        }
    };

    // Theme options for AI Create Site
    const themeOptions = [
        { name: 'ananke', url: 'https://github.com/theNewDynamic/gohugo-theme-ananke.git', label: 'Ananke (Recommended)' },
        { name: 'papermod', url: 'https://github.com/adityatelange/hugo-PaperMod.git', label: 'PaperMod' },
        { name: 'stack', url: 'https://github.com/CaiJimmy/hugo-theme-stack.git', label: 'Stack' },
        { name: 'blowfish', url: 'https://github.com/nunocoracao/blowfish.git', label: 'Blowfish' },
        { name: 'docsy', url: 'https://github.com/google/docsy.git', label: 'Docsy (Docs)' },
    ];

    const handleAICreateSite = async () => {
        if (!parentDir || !siteName) {
            setStatus({ type: 'error', message: 'PARENT_DIR_AND_NAME_REQUIRED' });
            addLog('Error: Please select a parent directory and enter a site name');
            return;
        }
        if (!aiConfigured) {
            setStatus({ type: 'error', message: 'AI_NOT_CONFIGURED' });
            addLog('Error: Please configure AI credentials first');
            return;
        }
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'CREATING_AI_SITE...' });
        addLog(`Creating AI-powered ${aiSiteType} site: ${siteName}`);

        const theme = themeOptions.find(t => t.name === aiSiteTheme);

        try {
            const result = await AICreateSite({
                parentDir,
                siteName,
                siteType: aiSiteType,
                description: aiSiteDescription,
                audience: aiSiteAudience,
                features: aiSiteFeatures,
                themeName: theme?.name || 'ananke',
                themeUrl: theme?.url || ''
            });
            if (result.success) {
                setStatus({ type: 'success', message: `AI_SITE_CREATED: ${result.sitePath}` });
                addLog(`Success: Site created at ${result.sitePath}`);
                addLog(`Files created: ${result.filesCreated}`);
                result.steps?.forEach((step: string) => addLog(`  ${step}`));
                setSitePath(result.sitePath);
                loadProjects();
                setTimeout(() => setActiveTab('manage'), 1500);
            } else {
                setStatus({ type: 'error', message: `AI_CREATE_FAILED: ${result.error}` });
                addLog(`Error: ${result.error}`);
            }
        } catch (err: any) {
            setStatus({ type: 'error', message: `SYSTEM_ERROR: ${err}` });
            addLog(`System Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleDeleteProject = async (projectId: number) => {
        if (!confirm('Are you sure you want to delete this project? This cannot be undone.')) {
            return;
        }
        setIsProcessing(true);
        addLog(`Deleting project ${projectId}...`);
        try {
            await DeleteProject(projectId);
            setStatus({ type: 'success', message: 'PROJECT_DELETED' });
            addLog('Project deleted successfully');
            loadProjects();
            setSelectedProject(null);
        } catch (err: any) {
            setStatus({ type: 'error', message: `DELETE_FAILED: ${err}` });
            addLog(`Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleArchiveProject = async (projectId: number) => {
        setIsProcessing(true);
        addLog(`Archiving project ${projectId}...`);
        try {
            const result = await ArchiveProject(projectId);
            if (result.success) {
                setStatus({ type: 'success', message: 'PROJECT_ARCHIVED' });
                addLog('Project archived successfully');
                loadProjects();
            } else {
                setStatus({ type: 'error', message: `ARCHIVE_FAILED: ${result.error}` });
                addLog(`Error: ${result.error}`);
            }
        } catch (err: any) {
            setStatus({ type: 'error', message: `SYSTEM_ERROR: ${err}` });
            addLog(`System Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleEditProject = async () => {
        if (!selectedProject) return;
        setIsProcessing(true);
        addLog(`Updating project ${selectedProject.id}...`);
        try {
            const result = await EditProject({
                projectId: selectedProject.id,
                name: editProjectName,
                category: editProjectCategory,
                description: editProjectDescription,
                imageUrl: '',
                suins: editProjectSuins
            });
            if (result.success) {
                setStatus({ type: 'success', message: 'PROJECT_UPDATED' });
                addLog('Project updated successfully');
                loadProjects();
                setSelectedProject(null);
            } else {
                setStatus({ type: 'error', message: `UPDATE_FAILED: ${result.error}` });
                addLog(`Error: ${result.error}`);
            }
        } catch (err: any) {
            setStatus({ type: 'error', message: `SYSTEM_ERROR: ${err}` });
            addLog(`System Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const selectProjectForEdit = (project: any) => {
        setSelectedProject(project);
        setEditProjectName(project.name || '');
        setEditProjectCategory(project.category || '');
        setEditProjectDescription(project.description || '');
        setEditProjectSuins(project.suinsDomain || '');
    };

    const NavItem = ({ id, icon: Icon, label }: { id: string, icon: any, label: string }) => (
        <button
            onClick={() => setActiveTab(id)}
            className={cn(
                "group relative flex items-center justify-center w-12 h-12 mb-4 transition-all duration-300",
                activeTab === id ? "text-accent" : "text-zinc-600 hover:text-zinc-400"
            )}
        >
            <div className={cn(
                "absolute inset-0 bg-accent/5 rounded-md scale-0 transition-transform duration-300",
                activeTab === id && "scale-100"
            )} />
            <div className={cn(
                "absolute left-0 w-1 h-8 bg-accent rounded-r-full transition-all duration-300",
                activeTab === id ? "opacity-100 translate-x-0" : "opacity-0 -translate-x-2"
            )} />
            <Icon size={20} strokeWidth={1.5} className="relative z-10" />
        </button>
    );

    const Card = ({ children, className, onClick }: { children: React.ReactNode, className?: string, onClick?: () => void }) => (
        <motion.div
            whileHover={onClick ? { scale: 1.01 } : {}}
            onClick={onClick}
            className={cn(
                "glass-panel-tech rounded-sm p-6 relative overflow-hidden group",
                onClick && "cursor-pointer",
                className
            )}
        >
            <div className="scanline opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
            <div className="absolute top-0 left-0 w-2 h-2 border-t border-l border-white/20" />
            <div className="absolute top-0 right-0 w-2 h-2 border-t border-r border-white/20" />
            <div className="absolute bottom-0 left-0 w-2 h-2 border-b border-l border-white/20" />
            <div className="absolute bottom-0 right-0 w-2 h-2 border-b border-r border-white/20" />
            <div className="relative z-10">{children}</div>
        </motion.div>
    );

    const WindowControls = () => (
        <div className="fixed top-0 right-0 z-[60] flex items-center h-8 px-2">
            <button
                onClick={WindowMinimise}
                className="w-8 h-8 flex items-center justify-center text-zinc-500 hover:text-white hover:bg-white/5 transition-colors"
            >
                <Minus size={14} />
            </button>
            <button
                onClick={WindowToggleMaximise}
                className="w-8 h-8 flex items-center justify-center text-zinc-500 hover:text-white hover:bg-white/5 transition-colors"
            >
                <Square size={12} />
            </button>
            <button
                onClick={Quit}
                className="w-8 h-8 flex items-center justify-center text-zinc-500 hover:text-white hover:bg-red-500/20 hover:text-red-400 transition-colors"
            >
                <X size={14} />
            </button>
        </div>
    );

    const containerVariants = {
        hidden: { opacity: 0 },
        show: {
            opacity: 1,
            transition: {
                staggerChildren: 0.1
            }
        }
    };

    const itemVariants = {
        hidden: { opacity: 0, y: 20 },
        show: { opacity: 1, y: 0 }
    };

    return (
        <div className="min-h-screen flex font-sans overflow-hidden relative">
            <div className="bg-noise" />

            {/* Window Drag Region */}
            <div className="fixed top-0 left-0 right-0 h-8 z-50 wails-drag" />

            {/* Window Controls */}
            <WindowControls />

            {/* Command Rail */}
            <div className="w-20 border-r border-white/5 flex flex-col items-center py-8 z-20 bg-[#0a0f14]/80 backdrop-blur-xl relative">
                <div className="mb-12">
                    <div className="w-10 h-10 bg-accent/10 rounded-sm flex items-center justify-center border border-accent/20">
                        <Hexagon className="text-accent" size={20} strokeWidth={1.5} />
                    </div>
                </div>

                <nav className="flex-1 w-full flex flex-col items-center">
                    <NavItem id="dashboard" icon={LayoutGrid} label="Dashboard" />
                    <NavItem id="create" icon={Plus} label="Create" />
                    <NavItem id="manage" icon={Settings} label="Manage" />
                    <NavItem id="projects" icon={Database} label="Projects" />
                    <NavItem id="ai" icon={Sparkles} label="AI" />
                    <NavItem id="diagnostics" icon={Stethoscope} label="Doctor" />
                </nav>

                <div className="mt-auto">
                    <div className="px-2 py-1 rounded-sm bg-zinc-900 border border-white/10 flex items-center justify-center text-[10px] font-mono text-zinc-500">
                        {version || 'v?'}
                    </div>
                </div>
            </div>

            {/* Main Content */}
            <div className="flex-1 flex flex-col h-screen overflow-hidden relative z-10">
                <header className="px-10 py-8 flex justify-between items-end z-10 border-b border-white/5 bg-[#0a0f14]/50 backdrop-blur-sm">
                    <div>
                        <h1 className="text-4xl font-display font-light tracking-wide text-white mb-1 uppercase">
                            {activeTab}
                        </h1>
                        <div className="flex items-center gap-2 text-zinc-500 text-xs font-mono tracking-widest">
                            <span className="text-accent">WALGO</span>
                            <ChevronRight size={10} />
                            <span>DESKTOP</span>
                            <ChevronRight size={10} />
                            <span className="uppercase text-zinc-300">{activeTab}</span>
                        </div>
                    </div>

                    <div className="flex items-center gap-6">
                        <div className="flex items-center gap-2 text-xs font-mono text-zinc-500">
                            <Radio size={14} className="text-accent animate-pulse" />
                            <span>NET_ONLINE</span>
                        </div>
                        <div className="px-3 py-1 rounded-sm bg-accent/5 border border-accent/20 text-xs font-mono text-accent">
                            READY
                        </div>
                    </div>
                </header>

                <main className="flex-1 overflow-y-auto px-10 py-10">
                    <AnimatePresence mode="wait">
                        <motion.div
                            key={activeTab}
                            variants={containerVariants}
                            initial="hidden"
                            animate="show"
                            exit={{ opacity: 0, scale: 0.98 }}
                            className="h-full"
                        >
                            {activeTab === 'dashboard' && (
                                <div className="grid grid-cols-3 gap-6 h-full max-h-[600px]">
                                    <motion.div variants={itemVariants} className="col-span-2 row-span-2 h-full">
                                        <Card onClick={() => setActiveTab('create')} className="h-full flex flex-col justify-between bg-zinc-900/20 hover:bg-zinc-900/30 transition-colors relative overflow-hidden group">
                                            <div className="absolute -right-10 -bottom-10 text-accent/5 group-hover:text-accent/10 transition-colors duration-500">
                                                <Plus size={200} strokeWidth={0.5} />
                                            </div>

                                            <div className="flex justify-between items-start relative z-10">
                                                <div className="p-2 bg-accent/10 rounded-sm border border-accent/20">
                                                    <Rocket size={20} className="text-accent" />
                                                </div>
                                                <ArrowUpRight className="text-zinc-600 group-hover:text-accent transition-colors" />
                                            </div>

                                            <div className="relative z-10">
                                                <h2 className="text-3xl font-display font-light text-white mb-2">QUICKSTART</h2>
                                                <p className="text-zinc-500 max-w-md font-light">Create, build, and deploy a new Hugo site in one command.</p>
                                            </div>
                                        </Card>
                                    </motion.div>

                                    <motion.div variants={itemVariants} className="col-span-1">
                                        <Card className="bg-zinc-900/20 h-full">
                                            <div className="flex items-center gap-3 mb-6 text-zinc-500 border-b border-white/5 pb-2">
                                                <Activity size={16} />
                                                <span className="text-xs font-mono uppercase tracking-widest">System_Status</span>
                                            </div>
                                            <div className="space-y-4 font-mono text-xs">
                                                <div className="flex justify-between items-center">
                                                    <span className="text-zinc-500">WALRUS_NODE</span>
                                                    <span className="text-accent">CONNECTED</span>
                                                </div>
                                                <div className="flex justify-between items-center">
                                                    <span className="text-zinc-500">DAEMON</span>
                                                    <span className="text-accent">ACTIVE</span>
                                                </div>
                                                <div className="h-px bg-white/5 my-4" />
                                                <div className="text-center py-2">
                                                    <div className="text-4xl font-display font-light text-white mb-1">{projects.length}</div>
                                                    <div className="text-[10px] text-zinc-600 uppercase tracking-widest">Projects</div>
                                                </div>
                                            </div>
                                        </Card>
                                    </motion.div>

                                    <motion.div variants={itemVariants} className="col-span-1">
                                        <Card onClick={() => setActiveTab('manage')} className="bg-zinc-900/20 h-full">
                                            <div className="flex justify-between items-start mb-6">
                                                <div className="p-2 bg-white/5 rounded-sm">
                                                    <Command size={20} className="text-zinc-300" />
                                                </div>
                                                <ArrowUpRight className="text-zinc-600 group-hover:text-white transition-colors" />
                                            </div>
                                            <h3 className="text-xl font-display text-white mb-1">OPERATIONS</h3>
                                            <p className="text-xs text-zinc-500 font-mono">BUILD_AND_DEPLOY</p>
                                        </Card>
                                    </motion.div>
                                </div>
                            )}

                            {activeTab === 'create' && (
                                <motion.div variants={itemVariants} className="max-w-3xl mx-auto pt-10">
                                    <div className="grid grid-cols-3 gap-6 mb-8">
                                        <Card onClick={() => setActiveTab('ai-create-site')} className="bg-gradient-to-br from-accent/20 to-purple-500/10 border-accent/30 h-full flex flex-col justify-center items-center gap-4 p-8 col-span-1">
                                            <Wand2 size={32} className="text-accent" />
                                            <div className="text-center">
                                                <h3 className="text-xl font-display text-white mb-2">AI Create</h3>
                                                <p className="text-xs text-zinc-500 font-mono">FULL_AI_WIZARD</p>
                                            </div>
                                        </Card>
                                        <Card onClick={() => setActiveTab('quickstart')} className="bg-zinc-900/20 h-full flex flex-col justify-center items-center gap-4 p-8">
                                            <Rocket size={32} className="text-accent" />
                                            <div className="text-center">
                                                <h3 className="text-xl font-display text-white mb-2">Quickstart</h3>
                                                <p className="text-xs text-zinc-500 font-mono">CREATE_BUILD</p>
                                            </div>
                                        </Card>
                                        <Card onClick={() => setActiveTab('create-site')} className="bg-zinc-900/20 h-full flex flex-col justify-center items-center gap-4 p-8">
                                            <Plus size={32} className="text-zinc-400" />
                                            <div className="text-center">
                                                <h3 className="text-xl font-display text-white mb-2">Basic</h3>
                                                <p className="text-xs text-zinc-500 font-mono">INITIALIZE_HUGO</p>
                                            </div>
                                        </Card>
                                    </div>

                                    <div className="grid grid-cols-4 gap-4">
                                        <Card onClick={() => setActiveTab('new-content')} className="bg-zinc-900/20 h-full flex flex-col justify-center items-center gap-3 p-5">
                                            <FileText size={22} className="text-zinc-400" />
                                            <div className="text-center">
                                                <h3 className="text-sm font-display text-white mb-1">New Content</h3>
                                                <p className="text-[9px] text-zinc-600 font-mono">CREATE_POST</p>
                                            </div>
                                        </Card>
                                        <Card onClick={() => setActiveTab('ai-generate')} className="bg-zinc-900/20 h-full flex flex-col justify-center items-center gap-3 p-5">
                                            <Sparkles size={22} className="text-zinc-400" />
                                            <div className="text-center">
                                                <h3 className="text-sm font-display text-white mb-1">AI Generate</h3>
                                                <p className="text-[9px] text-zinc-600 font-mono">AI_CONTENT</p>
                                            </div>
                                        </Card>
                                        <Card onClick={() => setActiveTab('import')} className="bg-zinc-900/20 h-full flex flex-col justify-center items-center gap-3 p-5">
                                            <Folder size={22} className="text-zinc-400" />
                                            <div className="text-center">
                                                <h3 className="text-sm font-display text-white mb-1">Import</h3>
                                                <p className="text-[9px] text-zinc-600 font-mono">OBSIDIAN</p>
                                            </div>
                                        </Card>
                                        <Card onClick={() => setActiveTab('manage')} className="bg-zinc-900/20 h-full flex flex-col justify-center items-center gap-3 p-5">
                                            <Settings size={22} className="text-zinc-400" />
                                            <div className="text-center">
                                                <h3 className="text-sm font-display text-white mb-1">Manage</h3>
                                                <p className="text-[9px] text-zinc-600 font-mono">BUILD_DEPLOY</p>
                                            </div>
                                        </Card>
                                    </div>
                                </motion.div>
                            )}

                            {activeTab === 'quickstart' && (
                                <motion.div variants={itemVariants} className="max-w-2xl mx-auto pt-10">
                                    <Card className="border-white/10 bg-zinc-900/20">
                                        <div className="flex items-center gap-4 mb-8 border-b border-white/5 pb-6">
                                            <div className="w-10 h-10 bg-accent text-black rounded-sm flex items-center justify-center shadow-[0_0_15px_rgba(77,162,255,0.3)]">
                                                <Rocket size={20} />
                                            </div>
                                            <div>
                                                <h2 className="text-xl font-display font-medium text-white">QUICKSTART</h2>
                                                <p className="text-zinc-500 text-xs font-mono">CREATE_BUILD_DEPLOY</p>
                                            </div>
                                        </div>

                                        <div className="space-y-6">
                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Parent_Directory</label>
                                                <div className="flex gap-2">
                                                    <input
                                                        type="text"
                                                        value={parentDir}
                                                        onChange={(e) => setParentDir(e.target.value)}
                                                        className="flex-1 bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                        placeholder="/path/to/parent"
                                                    />
                                                    <button
                                                        onClick={handleSelectParentDir}
                                                        className="px-4 py-3 bg-white/5 border border-white/10 rounded-sm hover:bg-white/10 transition-all"
                                                    >
                                                        <Folder size={18} className="text-zinc-400" />
                                                    </button>
                                                </div>
                                            </div>

                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Project_Name</label>
                                                <input
                                                    type="text"
                                                    value={siteName}
                                                    onChange={(e) => setSiteName(e.target.value)}
                                                    className="w-full bg-black/40 border border-white/10 rounded-sm px-6 py-4 text-lg text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-display tracking-wide"
                                                    placeholder="my-blog"
                                                    autoFocus
                                                />
                                            </div>

                                            <button
                                                onClick={handleQuickStart}
                                                disabled={isProcessing || !siteName || !parentDir}
                                                className="w-full bg-white text-black hover:bg-accent hover:text-black py-4 rounded-sm font-medium transition-all flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed uppercase tracking-wider text-sm"
                                            >
                                                {isProcessing ? <Loader2 className="animate-spin" /> : <Zap size={18} />}
                                                <span>Execute_Quickstart</span>
                                            </button>
                                        </div>
                                    </Card>
                                </motion.div>
                            )}

                            {activeTab === 'create-site' && (
                                <motion.div variants={itemVariants} className="max-w-2xl mx-auto pt-10">
                                    <Card className="border-white/10 bg-zinc-900/20">
                                        <div className="flex items-center gap-4 mb-8 border-b border-white/5 pb-6">
                                            <div className="w-10 h-10 bg-white/10 rounded-sm flex items-center justify-center">
                                                <Plus size={20} className="text-zinc-300" />
                                            </div>
                                            <div>
                                                <h2 className="text-xl font-display font-medium text-white">CREATE_SITE</h2>
                                                <p className="text-zinc-500 text-xs font-mono">INITIALIZE_HUGO</p>
                                            </div>
                                        </div>

                                        <div className="space-y-6">
                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Parent_Directory</label>
                                                <div className="flex gap-2">
                                                    <input
                                                        type="text"
                                                        value={parentDir}
                                                        onChange={(e) => setParentDir(e.target.value)}
                                                        className="flex-1 bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                        placeholder="/path/to/parent"
                                                    />
                                                    <button
                                                        onClick={handleSelectParentDir}
                                                        className="px-4 py-3 bg-white/5 border border-white/10 rounded-sm hover:bg-white/10 transition-all"
                                                    >
                                                        <Folder size={18} className="text-zinc-400" />
                                                    </button>
                                                </div>
                                            </div>

                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Project_Name</label>
                                                <input
                                                    type="text"
                                                    value={siteName}
                                                    onChange={(e) => setSiteName(e.target.value)}
                                                    className="w-full bg-black/40 border border-white/10 rounded-sm px-6 py-4 text-lg text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-display tracking-wide"
                                                    placeholder="my-blog"
                                                    autoFocus
                                                />
                                            </div>

                                            <button
                                                onClick={handleCreate}
                                                disabled={isProcessing || !siteName || !parentDir}
                                                className="w-full bg-white text-black hover:bg-accent hover:text-black py-4 rounded-sm font-medium transition-all flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed uppercase tracking-wider text-sm"
                                            >
                                                {isProcessing ? <Loader2 className="animate-spin" /> : <Plus size={18} />}
                                                <span>Create_Site</span>
                                            </button>
                                        </div>
                                    </Card>
                                </motion.div>
                            )}

                            {activeTab === 'new-content' && (
                                <motion.div variants={itemVariants} className="max-w-2xl mx-auto pt-10">
                                    <Card className="border-white/10 bg-zinc-900/20">
                                        <div className="flex items-center gap-4 mb-8 border-b border-white/5 pb-6">
                                            <div className="w-10 h-10 bg-white/10 rounded-sm flex items-center justify-center">
                                                <FileText size={20} className="text-zinc-300" />
                                            </div>
                                            <div>
                                                <h2 className="text-xl font-display font-medium text-white">NEW_CONTENT</h2>
                                                <p className="text-zinc-500 text-xs font-mono">CREATE_PAGE_OR_POST</p>
                                            </div>
                                        </div>

                                        <div className="space-y-6">
                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Hugo_Site</label>
                                                <div className="flex gap-2">
                                                    <input
                                                        type="text"
                                                        value={sitePath}
                                                        onChange={(e) => setSitePath(e.target.value)}
                                                        className="flex-1 bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                        placeholder="/path/to/hugo-site"
                                                    />
                                                    <button
                                                        onClick={handleSelectSitePath}
                                                        className="px-4 py-3 bg-white/5 border border-white/10 rounded-sm hover:bg-white/10 transition-all"
                                                    >
                                                        <Folder size={18} className="text-zinc-400" />
                                                    </button>
                                                </div>
                                            </div>

                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Content_Slug</label>
                                                <input
                                                    type="text"
                                                    value={contentSlug}
                                                    onChange={(e) => setContentSlug(e.target.value)}
                                                    className="w-full bg-black/40 border border-white/10 rounded-sm px-6 py-4 text-lg text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-display tracking-wide"
                                                    placeholder="posts/my-first-post.md"
                                                    autoFocus
                                                />
                                                <p className="text-[10px] font-mono text-zinc-600 mt-2 ml-1">
                                                    Examples: posts/hello.md, pages/about.md, blog/intro.md
                                                </p>
                                            </div>

                                            <button
                                                onClick={handleCreateNewContent}
                                                disabled={isProcessing || !sitePath || !contentSlug}
                                                className="w-full bg-white text-black hover:bg-accent hover:text-black py-4 rounded-sm font-medium transition-all flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed uppercase tracking-wider text-sm"
                                            >
                                                {isProcessing ? <Loader2 className="animate-spin" /> : <FileText size={18} />}
                                                <span>Create_Content</span>
                                            </button>
                                        </div>
                                    </Card>
                                </motion.div>
                            )}

                            {activeTab === 'ai-generate' && (
                                <motion.div variants={itemVariants} className="max-w-2xl mx-auto pt-10">
                                    <Card className="border-white/10 bg-zinc-900/20">
                                        <div className="flex items-center gap-4 mb-8 border-b border-white/5 pb-6">
                                            <div className="w-10 h-10 bg-accent/20 rounded-sm flex items-center justify-center">
                                                <Sparkles size={20} className="text-accent" />
                                            </div>
                                            <div>
                                                <h2 className="text-xl font-display font-medium text-white">AI_GENERATE</h2>
                                                <p className="text-zinc-500 text-xs font-mono">AI_POWERED_CONTENT</p>
                                            </div>
                                            {!aiConfigured && (
                                                <div className="ml-auto px-3 py-1 bg-yellow-500/10 border border-yellow-500/30 rounded-sm text-xs text-yellow-500 font-mono">
                                                    AI_NOT_CONFIGURED
                                                </div>
                                            )}
                                        </div>

                                        <div className="space-y-6">
                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Hugo_Site</label>
                                                <div className="flex gap-2">
                                                    <input
                                                        type="text"
                                                        value={sitePath}
                                                        onChange={(e) => setSitePath(e.target.value)}
                                                        className="flex-1 bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                        placeholder="/path/to/hugo-site"
                                                    />
                                                    <button
                                                        onClick={handleSelectSitePath}
                                                        className="px-4 py-3 bg-white/5 border border-white/10 rounded-sm hover:bg-white/10 transition-all"
                                                    >
                                                        <Folder size={18} className="text-zinc-400" />
                                                    </button>
                                                </div>
                                            </div>

                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Content_Type</label>
                                                <select
                                                    value={aiContentType}
                                                    onChange={(e) => setAiContentType(e.target.value)}
                                                    className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                >
                                                    <option value="post">Blog Post</option>
                                                    <option value="page">Page</option>
                                                </select>
                                            </div>

                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Topic</label>
                                                <input
                                                    type="text"
                                                    value={aiTopic}
                                                    onChange={(e) => setAiTopic(e.target.value)}
                                                    className="w-full bg-black/40 border border-white/10 rounded-sm px-6 py-4 text-lg text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-display tracking-wide"
                                                    placeholder="Getting started with blockchain"
                                                    autoFocus
                                                />
                                            </div>

                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Additional_Context (Optional)</label>
                                                <textarea
                                                    value={aiContext}
                                                    onChange={(e) => setAiContext(e.target.value)}
                                                    className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono h-24 resize-none"
                                                    placeholder="Target audience, tone, specific points to cover..."
                                                />
                                            </div>

                                            <button
                                                onClick={handleAIGenerate}
                                                disabled={isProcessing || !sitePath || !aiTopic || !aiConfigured}
                                                className="w-full bg-accent text-black hover:bg-accent/80 py-4 rounded-sm font-medium transition-all flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed uppercase tracking-wider text-sm"
                                            >
                                                {isProcessing ? <Loader2 className="animate-spin" /> : <Sparkles size={18} />}
                                                <span>Generate_Content</span>
                                            </button>

                                            {!aiConfigured && (
                                                <p className="text-center text-xs text-zinc-500 font-mono">
                                                    Configure AI credentials in the <button onClick={() => setActiveTab('ai')} className="text-accent hover:underline">AI tab</button> first
                                                </p>
                                            )}
                                        </div>
                                    </Card>
                                </motion.div>
                            )}

                            {activeTab === 'manage' && (
                                <div className="grid grid-cols-3 gap-6 h-full">
                                    <div className="col-span-2 space-y-6">
                                        <motion.div variants={itemVariants}>
                                            <Card className="border-white/10 bg-zinc-900/20">
                                                <div className="flex items-center gap-3 mb-4">
                                                    <Folder size={16} className="text-zinc-500" />
                                                    <span className="text-xs font-mono text-zinc-500 uppercase tracking-widest">Target_Directory</span>
                                                    {sitePath && (
                                                        <button
                                                            onClick={() => handleOpenInFinder(sitePath)}
                                                            className="ml-auto text-xs text-accent hover:underline font-mono"
                                                        >
                                                            Open in Finder
                                                        </button>
                                                    )}
                                                </div>
                                                <div className="flex gap-2">
                                                    <input
                                                        type="text"
                                                        value={sitePath}
                                                        onChange={(e) => setSitePath(e.target.value)}
                                                        className="flex-1 bg-black/40 border border-white/10 rounded-sm px-5 py-3 text-sm font-mono text-zinc-300 focus:outline-none focus:border-accent/50 transition-all"
                                                        placeholder="/PATH/TO/SITE"
                                                    />
                                                    <button
                                                        onClick={handleSelectSitePath}
                                                        className="px-4 py-3 bg-white/5 border border-white/10 rounded-sm hover:bg-white/10 transition-all"
                                                    >
                                                        <Folder size={18} className="text-zinc-400" />
                                                    </button>
                                                </div>
                                                {serverRunning && serverUrl && (
                                                    <div className="mt-3 flex items-center gap-2">
                                                        <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                                                        <span className="text-xs font-mono text-green-400">Server running at {serverUrl}</span>
                                                        <button
                                                            onClick={() => handleOpenInBrowser(serverUrl)}
                                                            className="ml-2 text-xs text-accent hover:underline font-mono"
                                                        >
                                                            Open
                                                        </button>
                                                        <button
                                                            onClick={handleStopServe}
                                                            className="ml-2 text-xs text-red-400 hover:underline font-mono"
                                                        >
                                                            Stop
                                                        </button>
                                                    </div>
                                                )}
                                            </Card>
                                        </motion.div>

                                        <div className="grid grid-cols-3 gap-4">
                                            <motion.div variants={itemVariants}>
                                                <button
                                                    onClick={handleBuild}
                                                    disabled={isProcessing || !sitePath}
                                                    className="w-full glass-panel-tech p-6 rounded-sm text-left hover:border-accent/30 transition-all group disabled:opacity-50 relative overflow-hidden"
                                                >
                                                    <div className="absolute inset-0 bg-accent/5 translate-y-full group-hover:translate-y-0 transition-transform duration-500" />
                                                    <Cpu size={24} className="text-zinc-600 group-hover:text-accent mb-3 transition-colors relative z-10" />
                                                    <h3 className="text-base font-display font-medium text-white mb-1 relative z-10">BUILD</h3>
                                                    <p className="text-[10px] text-zinc-500 font-mono relative z-10">COMPILE</p>
                                                </button>
                                            </motion.div>

                                            <motion.div variants={itemVariants}>
                                                <button
                                                    onClick={handleDeploy}
                                                    disabled={isProcessing || !sitePath}
                                                    className="w-full glass-panel-tech p-6 rounded-sm text-left hover:border-accent/30 transition-all group disabled:opacity-50 relative overflow-hidden"
                                                >
                                                    <div className="absolute inset-0 bg-accent/5 translate-y-full group-hover:translate-y-0 transition-transform duration-500" />
                                                    <HardDrive size={24} className="text-zinc-600 group-hover:text-accent mb-3 transition-colors relative z-10" />
                                                    <h3 className="text-base font-display font-medium text-white mb-1 relative z-10">DEPLOY</h3>
                                                    <p className="text-[10px] text-zinc-500 font-mono relative z-10">UPLOAD</p>
                                                </button>
                                            </motion.div>

                                            <motion.div variants={itemVariants}>
                                                <button
                                                    onClick={handleUpdate}
                                                    disabled={isProcessing || !sitePath}
                                                    className="w-full glass-panel-tech p-6 rounded-sm text-left hover:border-accent/30 transition-all group disabled:opacity-50 relative overflow-hidden"
                                                >
                                                    <div className="absolute inset-0 bg-accent/5 translate-y-full group-hover:translate-y-0 transition-transform duration-500" />
                                                    <RefreshCw size={24} className="text-zinc-600 group-hover:text-accent mb-3 transition-colors relative z-10" />
                                                    <h3 className="text-base font-display font-medium text-white mb-1 relative z-10">UPDATE</h3>
                                                    <p className="text-[10px] text-zinc-500 font-mono relative z-10">REFRESH</p>
                                                </button>
                                            </motion.div>
                                        </div>

                                        <div className="grid grid-cols-2 gap-4 mt-6">
                                            <motion.div variants={itemVariants}>
                                                <button
                                                    onClick={handleServe}
                                                    disabled={isProcessing || !sitePath}
                                                    className="w-full glass-panel-tech p-6 rounded-sm text-left hover:border-accent/30 transition-all group disabled:opacity-50 relative overflow-hidden"
                                                >
                                                    <div className="absolute inset-0 bg-accent/5 translate-y-full group-hover:translate-y-0 transition-transform duration-500" />
                                                    <Play size={24} className="text-zinc-600 group-hover:text-accent mb-3 transition-colors relative z-10" />
                                                    <h3 className="text-base font-display font-medium text-white mb-1 relative z-10">SERVE</h3>
                                                    <p className="text-[10px] text-zinc-500 font-mono relative z-10">DEV_SERVER</p>
                                                </button>
                                            </motion.div>

                                            <motion.div variants={itemVariants}>
                                                <button
                                                    onClick={handleLaunch}
                                                    disabled={isProcessing || !sitePath}
                                                    className="w-full glass-panel-tech p-6 rounded-sm text-left hover:border-accent/30 transition-all group disabled:opacity-50 relative overflow-hidden"
                                                >
                                                    <div className="absolute inset-0 bg-accent/5 translate-y-full group-hover:translate-y-0 transition-transform duration-500" />
                                                    <Rocket size={24} className="text-zinc-600 group-hover:text-accent mb-3 transition-colors relative z-10" />
                                                    <h3 className="text-base font-display font-medium text-white mb-1 relative z-10">LAUNCH</h3>
                                                    <p className="text-[10px] text-zinc-500 font-mono relative z-10">WIZARD</p>
                                                </button>
                                            </motion.div>
                                        </div>
                                    </div>

                                    <motion.div variants={itemVariants} className="col-span-1 glass-panel-tech rounded-sm border-white/5 flex flex-col overflow-hidden h-full max-h-[500px]">
                                        <div className="p-3 border-b border-white/5 bg-black/20 flex justify-between items-center">
                                            <div className="flex items-center gap-2 text-[10px] font-mono text-zinc-400 uppercase tracking-wider">
                                                <Terminal size={12} />
                                                <span>System_Log</span>
                                            </div>
                                            {status && (
                                                <div className={cn(
                                                    "w-1.5 h-1.5 rounded-full animate-pulse",
                                                    status.type === 'success' ? "bg-accent" :
                                                        status.type === 'error' ? "bg-red-500" : "bg-blue-500"
                                                )} />
                                            )}
                                        </div>
                                        <div className="flex-1 p-4 font-mono text-[10px] text-zinc-500 overflow-y-auto space-y-1 scrollbar-thin scrollbar-thumb-zinc-800">
                                            <div className="text-zinc-700">{'>'} SYSTEM_READY...</div>
                                            {logs.map((log, i) => (
                                                <div key={i} className="text-zinc-400 border-l border-white/10 pl-2 hover:text-white transition-colors">
                                                    {log}
                                                </div>
                                            ))}
                                            {status && (
                                                <div className={cn(
                                                    "pt-2 font-bold",
                                                    status.type === 'success' ? "text-accent" :
                                                        status.type === 'error' ? "text-red-400" : "text-blue-400"
                                                )}>
                                                    {`> ${status.message}`}
                                                </div>
                                            )}
                                        </div>
                                    </motion.div>
                                </div>
                            )}

                            {activeTab === 'projects' && (
                                <div className="grid grid-cols-3 gap-6 h-full">
                                    <motion.div variants={itemVariants} className="col-span-2 h-full">
                                        <Card className="bg-zinc-900/20 h-full flex flex-col">
                                            <div className="flex items-center gap-3 mb-6 pb-4 border-b border-white/5">
                                                <Database size={20} className="text-zinc-500" />
                                                <div>
                                                    <h2 className="text-xl font-display text-white">PROJECTS</h2>
                                                    <p className="text-xs text-zinc-500 font-mono">{projects.length} deployed sites</p>
                                                </div>
                                                <button
                                                    onClick={loadProjects}
                                                    className="ml-auto p-2 hover:bg-white/5 rounded-sm transition-all"
                                                >
                                                    <RefreshCw size={16} className="text-zinc-500" />
                                                </button>
                                            </div>

                                            <div className="flex-1 overflow-y-auto space-y-3">
                                                {projects.length === 0 ? (
                                                    <div className="text-center py-10 text-zinc-500 font-mono text-sm">
                                                        <Database size={40} className="mx-auto mb-4 opacity-30" />
                                                        <p>No projects deployed yet.</p>
                                                        <button
                                                            onClick={() => setActiveTab('create')}
                                                            className="mt-4 text-accent hover:underline"
                                                        >
                                                            Create your first site
                                                        </button>
                                                    </div>
                                                ) : (
                                                    projects.map((project: any) => (
                                                        <div
                                                            key={project.id}
                                                            onClick={() => selectProjectForEdit(project)}
                                                            className={cn(
                                                                "p-4 bg-black/40 border rounded-sm transition-all cursor-pointer",
                                                                selectedProject?.id === project.id
                                                                    ? "border-accent/50 bg-accent/5"
                                                                    : "border-white/10 hover:border-accent/30"
                                                            )}
                                                        >
                                                            <div className="flex justify-between items-start mb-2">
                                                                <div>
                                                                    <h3 className="text-white font-display mb-1">{project.name}</h3>
                                                                    <div className="flex items-center gap-2">
                                                                        <span className="text-xs text-zinc-500 font-mono">{project.network}</span>
                                                                        {project.category && (
                                                                            <span className="text-[10px] px-2 py-0.5 bg-white/5 rounded-sm text-zinc-400 font-mono">
                                                                                {project.category}
                                                                            </span>
                                                                        )}
                                                                    </div>
                                                                </div>
                                                                <div className={cn(
                                                                    "text-xs font-mono px-2 py-1 rounded-sm",
                                                                    project.status === 'active'
                                                                        ? "text-green-400 bg-green-500/10"
                                                                        : "text-zinc-400 bg-zinc-500/10"
                                                                )}>
                                                                    {project.status}
                                                                </div>
                                                            </div>
                                                            <div className="grid grid-cols-2 gap-4 text-xs font-mono text-zinc-400 mt-3">
                                                                <div>
                                                                    <span className="text-zinc-600">Object ID:</span> {project.objectId?.slice(0, 16)}...
                                                                </div>
                                                                <div>
                                                                    <span className="text-zinc-600">Deploys:</span> {project.deploymentCount || 0}
                                                                </div>
                                                            </div>
                                                            {project.suinsDomain && (
                                                                <div className="mt-2 text-xs font-mono text-accent">
                                                                    {project.suinsDomain}
                                                                </div>
                                                            )}
                                                            <div className="flex gap-2 mt-3 pt-3 border-t border-white/5">
                                                                {project.objectId && (
                                                                    <button
                                                                        onClick={(e) => {
                                                                            e.stopPropagation();
                                                                            handleOpenInBrowser(`https://${project.objectId}.walrus.site`);
                                                                        }}
                                                                        className="text-[10px] text-accent hover:underline font-mono flex items-center gap-1"
                                                                    >
                                                                        <ExternalLink size={12} /> View Site
                                                                    </button>
                                                                )}
                                                                <button
                                                                    onClick={(e) => {
                                                                        e.stopPropagation();
                                                                        handleArchiveProject(project.id);
                                                                    }}
                                                                    className="text-[10px] text-zinc-500 hover:text-yellow-400 font-mono flex items-center gap-1"
                                                                >
                                                                    <Archive size={12} /> Archive
                                                                </button>
                                                                <button
                                                                    onClick={(e) => {
                                                                        e.stopPropagation();
                                                                        handleDeleteProject(project.id);
                                                                    }}
                                                                    className="text-[10px] text-zinc-500 hover:text-red-400 font-mono flex items-center gap-1"
                                                                >
                                                                    <Trash2 size={12} /> Delete
                                                                </button>
                                                            </div>
                                                        </div>
                                                    ))
                                                )}
                                            </div>
                                        </Card>
                                    </motion.div>

                                    <motion.div variants={itemVariants} className="h-full">
                                        <Card className="bg-zinc-900/20 h-full flex flex-col">
                                            <div className="flex items-center gap-3 mb-6 pb-4 border-b border-white/5">
                                                <Edit3 size={20} className="text-zinc-500" />
                                                <div>
                                                    <h2 className="text-lg font-display text-white">EDIT_PROJECT</h2>
                                                    <p className="text-xs text-zinc-500 font-mono">
                                                        {selectedProject ? selectedProject.name : 'Select a project'}
                                                    </p>
                                                </div>
                                            </div>

                                            {selectedProject ? (
                                                <div className="space-y-4 flex-1">
                                                    <div className="group">
                                                        <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Name</label>
                                                        <input
                                                            type="text"
                                                            value={editProjectName}
                                                            onChange={(e) => setEditProjectName(e.target.value)}
                                                            className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-2 text-sm text-white focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                        />
                                                    </div>
                                                    <div className="group">
                                                        <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Category</label>
                                                        <select
                                                            value={editProjectCategory}
                                                            onChange={(e) => setEditProjectCategory(e.target.value)}
                                                            className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-2 text-sm text-white focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                        >
                                                            <option value="">Select Category</option>
                                                            <option value="blog">Blog</option>
                                                            <option value="portfolio">Portfolio</option>
                                                            <option value="docs">Documentation</option>
                                                            <option value="business">Business</option>
                                                            <option value="personal">Personal</option>
                                                        </select>
                                                    </div>
                                                    <div className="group">
                                                        <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Description</label>
                                                        <textarea
                                                            value={editProjectDescription}
                                                            onChange={(e) => setEditProjectDescription(e.target.value)}
                                                            className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-2 text-sm text-white focus:outline-none focus:border-accent/50 transition-all font-mono h-20 resize-none"
                                                        />
                                                    </div>
                                                    <div className="group">
                                                        <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">SuiNS_Domain</label>
                                                        <input
                                                            type="text"
                                                            value={editProjectSuins}
                                                            onChange={(e) => setEditProjectSuins(e.target.value)}
                                                            className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-2 text-sm text-white focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                            placeholder="mysite.sui"
                                                        />
                                                    </div>
                                                    <div className="flex gap-2 pt-4 mt-auto">
                                                        <button
                                                            onClick={handleEditProject}
                                                            disabled={isProcessing}
                                                            className="flex-1 bg-accent text-black hover:bg-accent/80 py-2 rounded-sm font-medium transition-all flex items-center justify-center gap-2 disabled:opacity-50 text-sm"
                                                        >
                                                            {isProcessing ? <Loader2 className="animate-spin" size={16} /> : <Save size={16} />}
                                                            <span>Save</span>
                                                        </button>
                                                        <button
                                                            onClick={() => setSelectedProject(null)}
                                                            className="px-4 py-2 bg-white/5 hover:bg-white/10 rounded-sm transition-all text-sm text-zinc-400"
                                                        >
                                                            Cancel
                                                        </button>
                                                    </div>
                                                </div>
                                            ) : (
                                                <div className="flex-1 flex items-center justify-center text-zinc-600 font-mono text-sm">
                                                    Click a project to edit
                                                </div>
                                            )}
                                        </Card>
                                    </motion.div>
                                </div>
                            )}

                            {activeTab === 'ai' && (
                                <motion.div variants={itemVariants} className="h-full">
                                    <Card className="bg-zinc-900/20 h-full">
                                        <div className="flex items-center gap-4 mb-8 border-b border-white/5 pb-6">
                                            <div className="w-10 h-10 bg-white/10 rounded-sm flex items-center justify-center">
                                                <Sparkles size={20} className="text-zinc-300" />
                                            </div>
                                            <div>
                                                <h2 className="text-xl font-display font-medium text-white">AI_CONFIGURATION</h2>
                                                <p className="text-zinc-500 text-xs font-mono">AI_PROVIDER_SETUP</p>
                                            </div>
                                        </div>

                                        <div className="grid grid-cols-3 gap-6 mb-8">
                                            <div className="col-span-2 space-y-6">
                                                <div className="group">
                                                    <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Provider</label>
                                                    <select
                                                        value={aiProvider}
                                                        onChange={(e) => setAiProvider(e.target.value)}
                                                        className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                    >
                                                        <option value="">Select Provider</option>
                                                        <option value="openai">OpenAI</option>
                                                        <option value="openrouter">OpenRouter</option>
                                                    </select>
                                                </div>

                                                <div className="group">
                                                    <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">API_Key</label>
                                                    <input
                                                        type="password"
                                                        value={aiApiKey}
                                                        onChange={(e) => setAiApiKey(e.target.value)}
                                                        placeholder="sk-..."
                                                        className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                    />
                                                </div>

                                                <div className="group">
                                                    <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Base_URL</label>
                                                    <input
                                                        type="text"
                                                        value={aiBaseUrl}
                                                        onChange={(e) => setAiBaseUrl(e.target.value)}
                                                        placeholder={aiProvider === 'openai' ? 'https://api.openai.com/v1' : 'https://openrouter.ai/api/v1'}
                                                        className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                    />
                                                </div>

                                                <div className="group">
                                                    <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Model</label>
                                                    <input
                                                        type="text"
                                                        value={aiModel}
                                                        onChange={(e) => setAiModel(e.target.value)}
                                                        placeholder={aiProvider === 'openrouter' ? 'openai/gpt-4' : 'gpt-4'}
                                                        className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                    />
                                                </div>
                                            </div>

                                            <div className="col-span-1 space-y-4">
                                                <button
                                                    onClick={handleGetAIConfig}
                                                    disabled={isProcessing}
                                                    className="w-full glass-panel-tech p-4 rounded-sm text-left hover:border-accent/30 transition-all group disabled:opacity-50 relative overflow-hidden mb-4"
                                                >
                                                    <div className="absolute inset-0 bg-accent/5 translate-y-full group-hover:translate-y-0 transition-transform duration-500" />
                                                    <RefreshCw size={20} className="text-zinc-600 group-hover:text-accent mb-2 transition-colors relative z-10" />
                                                    <div className="text-sm font-display font-medium text-white relative z-10">Get Config</div>
                                                </button>

                                                {aiConfig && (
                                                    <div className="mb-4 p-4 bg-black/40 border border-white/10 rounded-sm">
                                                        <h4 className="text-sm font-display text-white mb-3">CURRENT_CONFIG</h4>
                                                        <div className="space-y-2 font-mono text-xs text-zinc-400">
                                                            <div className="flex justify-between">
                                                                <span>Enabled:</span>
                                                                <span className={aiConfig.enabled ? "text-accent" : "text-zinc-600"}>
                                                                    {aiConfig.enabled ? "YES" : "NO"}
                                                                </span>
                                                            </div>
                                                            {aiConfig.currentProvider && (
                                                                <>
                                                                    <div className="flex justify-between">
                                                                        <span>Provider:</span>
                                                                        <span className="text-zinc-300">{aiConfig.currentProvider}</span>
                                                                    </div>
                                                                    {aiConfig.currentModel && (
                                                                        <div className="flex justify-between">
                                                                            <span>Model:</span>
                                                                            <span className="text-zinc-300">{aiConfig.currentModel}</span>
                                                                        </div>
                                                                    )}
                                                                </>
                                                            )}
                                                            {aiConfig.configuredProviders && aiConfig.configuredProviders.length > 0 && (
                                                                <div className="mt-3 pt-3 border-t border-white/10">
                                                                    <div className="text-xs text-zinc-500 mb-1">CONFIGURED_PROVIDERS:</div>
                                                                    {aiConfig.configuredProviders.map((provider: string, i: number) => (
                                                                        <div key={i} className="text-xs text-zinc-300 px-2 py-1 bg-white/5 rounded">
                                                                            {provider}
                                                                        </div>
                                                                    ))}
                                                                </div>
                                                            )}
                                                        </div>
                                                    </div>
                                                )}

                                                <button
                                                    onClick={handleUpdateAIConfig}
                                                    disabled={isProcessing || !aiProvider || !aiApiKey}
                                                    className="w-full glass-panel-tech p-4 rounded-sm text-left hover:border-accent/30 transition-all group disabled:opacity-50 relative overflow-hidden"
                                                >
                                                    <div className="absolute inset-0 bg-accent/5 translate-y-full group-hover:translate-y-0 transition-transform duration-500" />
                                                    <Save size={20} className="text-zinc-600 group-hover:text-accent mb-2 transition-colors relative z-10" />
                                                    <div className="text-sm font-display font-medium text-white relative z-10">Save Config</div>
                                                </button>

                                                <button
                                                    onClick={handleCleanAIConfig}
                                                    disabled={isProcessing}
                                                    className="w-full glass-panel-tech p-4 rounded-sm text-left hover:border-red-500/30 transition-all group disabled:opacity-50 relative overflow-hidden"
                                                >
                                                    <div className="absolute inset-0 bg-red-500/5 translate-y-full group-hover:translate-y-0 transition-transform duration-500" />
                                                    <Trash2 size={20} className="text-zinc-600 group-hover:text-red-400 mb-2 transition-colors relative z-10" />
                                                    <div className="text-sm font-display font-medium text-white relative z-10">Clean Config</div>
                                                </button>
                                            </div>
                                        </div>

                                        <div className="mt-8 pt-6 border-t border-white/5">
                                            <h3 className="text-lg font-display text-white mb-6">AI_CAPABILITIES</h3>
                                            <div className="grid grid-cols-2 gap-4 font-mono text-sm">
                                                <div className="flex items-center gap-3 text-zinc-400">
                                                    <Check size={16} className="text-accent" />
                                                    <span>Generate blog posts and pages</span>
                                                </div>
                                                <div className="flex items-center gap-3 text-zinc-400">
                                                    <Check size={16} className="text-accent" />
                                                    <span>Update existing content</span>
                                                </div>
                                                <div className="flex items-center gap-3 text-zinc-400">
                                                    <Check size={16} className="text-accent" />
                                                    <span>Create complete sites with AI</span>
                                                </div>
                                                <div className="flex items-center gap-3 text-zinc-400">
                                                    <Check size={16} className="text-accent" />
                                                    <span>Multiple AI providers supported</span>
                                                </div>
                                                <div className="flex items-center gap-3 text-zinc-400">
                                                    <Key size={16} className="text-accent" />
                                                    <span>Credentials stored securely in ~/.walgo/</span>
                                                </div>
                                            </div>
                                        </div>
                                    </Card>
                                </motion.div>
                            )}

                            {activeTab === 'diagnostics' && (
                                <motion.div variants={itemVariants} className="h-full">
                                    <Card className="bg-zinc-900/20 h-full">
                                        <div className="flex items-center gap-4 mb-8 border-b border-white/5 pb-6">
                                            <div className="w-10 h-10 bg-white/10 rounded-sm flex items-center justify-center">
                                                <Stethoscope size={20} className="text-zinc-300" />
                                            </div>
                                            <div>
                                                <h2 className="text-xl font-display font-medium text-white">DOCTOR</h2>
                                                <p className="text-zinc-500 text-xs font-mono">ENVIRONMENT_DIAGNOSTICS</p>
                                            </div>
                                        </div>

                                        <div className="space-y-4">
                                            <button
                                                onClick={handleDoctor}
                                                disabled={isProcessing}
                                                className="w-full bg-accent/10 hover:bg-accent/20 text-accent border border-accent/30 py-4 rounded-sm font-medium transition-all flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed uppercase tracking-wider text-sm"
                                            >
                                                {isProcessing ? <Loader2 className="animate-spin" /> : <Stethoscope size={18} />}
                                                <span>Run_Diagnostics</span>
                                            </button>

                                            <div className="mt-6 p-4 bg-black/40 border border-white/10 rounded-sm">
                                                <h4 className="text-sm font-display text-white mb-3">Checks Performed:</h4>
                                                <div className="space-y-2 font-mono text-xs text-zinc-400">
                                                    <div>• Hugo installation</div>
                                                    <div>• site-builder availability</div>
                                                    <div>• walrus CLI presence</div>
                                                    <div>• Sui client configuration</div>
                                                    <div>• Wallet balance check</div>
                                                    <div>• Configuration files</div>
                                                </div>
                                            </div>
                                        </div>
                                    </Card>
                                </motion.div>
                            )}

                            {activeTab === 'import' && (
                                <motion.div variants={itemVariants} className="max-w-2xl mx-auto pt-10">
                                    <Card className="border-white/10 bg-zinc-900/20">
                                        <div className="flex items-center gap-4 mb-8 border-b border-white/5 pb-6">
                                            <div className="w-10 h-10 bg-purple-500/20 rounded-sm flex items-center justify-center">
                                                <Folder size={20} className="text-purple-400" />
                                            </div>
                                            <div>
                                                <h2 className="text-xl font-display font-medium text-white">IMPORT_OBSIDIAN</h2>
                                                <p className="text-zinc-500 text-xs font-mono">VAULT_TO_HUGO</p>
                                            </div>
                                        </div>

                                        <div className="space-y-6">
                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Hugo_Site</label>
                                                <div className="flex gap-2">
                                                    <input
                                                        type="text"
                                                        value={sitePath}
                                                        onChange={(e) => setSitePath(e.target.value)}
                                                        className="flex-1 bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                        placeholder="/path/to/hugo-site"
                                                    />
                                                    <button
                                                        onClick={handleSelectSitePath}
                                                        className="px-4 py-3 bg-white/5 border border-white/10 rounded-sm hover:bg-white/10 transition-all"
                                                    >
                                                        <Folder size={18} className="text-zinc-400" />
                                                    </button>
                                                </div>
                                            </div>

                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Obsidian_Vault</label>
                                                <div className="flex gap-2">
                                                    <input
                                                        type="text"
                                                        value={vaultPath}
                                                        onChange={(e) => setVaultPath(e.target.value)}
                                                        className="flex-1 bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                        placeholder="/path/to/obsidian-vault"
                                                    />
                                                    <button
                                                        onClick={handleSelectVaultPath}
                                                        className="px-4 py-3 bg-white/5 border border-white/10 rounded-sm hover:bg-white/10 transition-all"
                                                    >
                                                        <Folder size={18} className="text-zinc-400" />
                                                    </button>
                                                </div>
                                            </div>

                                            <div className="flex items-center gap-3">
                                                <input
                                                    type="checkbox"
                                                    id="includeDrafts"
                                                    checked={includeDrafts}
                                                    onChange={(e) => setIncludeDrafts(e.target.checked)}
                                                    className="w-4 h-4 bg-black/40 border border-white/10 rounded-sm text-accent focus:ring-accent"
                                                />
                                                <label htmlFor="includeDrafts" className="text-sm text-zinc-400 font-mono">
                                                    Include draft files
                                                </label>
                                            </div>

                                            <div className="p-4 bg-black/40 rounded-sm border border-white/10">
                                                <h4 className="text-sm font-display text-white mb-3">What Gets Imported:</h4>
                                                <div className="space-y-2 font-mono text-xs text-zinc-400">
                                                    <div className="flex items-center gap-2"><Check size={14} className="text-accent" /> Convert markdown files to Hugo format</div>
                                                    <div className="flex items-center gap-2"><Check size={14} className="text-accent" /> Preserve wikilinks and frontmatter</div>
                                                    <div className="flex items-center gap-2"><Check size={14} className="text-accent" /> Handle attachments and images</div>
                                                    <div className="flex items-center gap-2"><Check size={14} className="text-accent" /> Support drafts and published content</div>
                                                </div>
                                            </div>

                                            <button
                                                onClick={handleImportObsidian}
                                                disabled={isProcessing || !sitePath || !vaultPath}
                                                className="w-full bg-purple-500 text-white hover:bg-purple-400 py-4 rounded-sm font-medium transition-all flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed uppercase tracking-wider text-sm"
                                            >
                                                {isProcessing ? <Loader2 className="animate-spin" /> : <Folder size={18} />}
                                                <span>Import_Vault</span>
                                            </button>
                                        </div>
                                    </Card>
                                </motion.div>
                            )}

                            {activeTab === 'ai-create-site' && (
                                <motion.div variants={itemVariants} className="max-w-2xl mx-auto pt-10">
                                    <Card className="border-accent/20 bg-gradient-to-br from-zinc-900/40 to-accent/5">
                                        <div className="flex items-center gap-4 mb-8 border-b border-white/5 pb-6">
                                            <div className="w-10 h-10 bg-accent text-black rounded-sm flex items-center justify-center shadow-[0_0_20px_rgba(77,162,255,0.4)]">
                                                <Wand2 size={20} />
                                            </div>
                                            <div>
                                                <h2 className="text-xl font-display font-medium text-white">AI_CREATE_SITE</h2>
                                                <p className="text-zinc-500 text-xs font-mono">FULL_AI_WIZARD</p>
                                            </div>
                                            {!aiConfigured && (
                                                <div className="ml-auto px-3 py-1 bg-yellow-500/10 border border-yellow-500/30 rounded-sm text-xs text-yellow-500 font-mono">
                                                    AI_NOT_CONFIGURED
                                                </div>
                                            )}
                                        </div>

                                        <div className="space-y-5">
                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Parent_Directory</label>
                                                <div className="flex gap-2">
                                                    <input
                                                        type="text"
                                                        value={parentDir}
                                                        onChange={(e) => setParentDir(e.target.value)}
                                                        className="flex-1 bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                        placeholder="/path/to/parent"
                                                    />
                                                    <button
                                                        onClick={handleSelectParentDir}
                                                        className="px-4 py-3 bg-white/5 border border-white/10 rounded-sm hover:bg-white/10 transition-all"
                                                    >
                                                        <Folder size={18} className="text-zinc-400" />
                                                    </button>
                                                </div>
                                            </div>

                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Site_Name</label>
                                                <input
                                                    type="text"
                                                    value={siteName}
                                                    onChange={(e) => setSiteName(e.target.value)}
                                                    className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                    placeholder="my-awesome-blog"
                                                />
                                            </div>

                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Site_Type</label>
                                                <div className="grid grid-cols-4 gap-2">
                                                    {[
                                                        { id: 'blog', label: 'Blog', icon: BookOpen },
                                                        { id: 'portfolio', label: 'Portfolio', icon: Palette },
                                                        { id: 'docs', label: 'Docs', icon: FileCode },
                                                        { id: 'business', label: 'Business', icon: Briefcase },
                                                    ].map(type => (
                                                        <button
                                                            key={type.id}
                                                            onClick={() => setAiSiteType(type.id)}
                                                            className={cn(
                                                                "p-3 rounded-sm border transition-all flex flex-col items-center gap-2",
                                                                aiSiteType === type.id
                                                                    ? "bg-accent/20 border-accent/50 text-accent"
                                                                    : "bg-black/20 border-white/10 text-zinc-500 hover:border-white/20"
                                                            )}
                                                        >
                                                            <type.icon size={18} />
                                                            <span className="text-[10px] font-mono">{type.label}</span>
                                                        </button>
                                                    ))}
                                                </div>
                                            </div>

                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Site_Description</label>
                                                <textarea
                                                    value={aiSiteDescription}
                                                    onChange={(e) => setAiSiteDescription(e.target.value)}
                                                    className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono h-20 resize-none"
                                                    placeholder="A tech blog about web development and AI..."
                                                />
                                            </div>

                                            <div className="grid grid-cols-2 gap-4">
                                                <div className="group">
                                                    <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Target_Audience</label>
                                                    <input
                                                        type="text"
                                                        value={aiSiteAudience}
                                                        onChange={(e) => setAiSiteAudience(e.target.value)}
                                                        className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                        placeholder="Developers, tech enthusiasts"
                                                    />
                                                </div>
                                                <div className="group">
                                                    <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Theme</label>
                                                    <select
                                                        value={aiSiteTheme}
                                                        onChange={(e) => setAiSiteTheme(e.target.value)}
                                                        className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                    >
                                                        {themeOptions.map(theme => (
                                                            <option key={theme.name} value={theme.name}>{theme.label}</option>
                                                        ))}
                                                    </select>
                                                </div>
                                            </div>

                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Features_Pages (Optional)</label>
                                                <input
                                                    type="text"
                                                    value={aiSiteFeatures}
                                                    onChange={(e) => setAiSiteFeatures(e.target.value)}
                                                    className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                                                    placeholder="Projects page, Newsletter signup, Contact form"
                                                />
                                            </div>

                                            <button
                                                onClick={handleAICreateSite}
                                                disabled={isProcessing || !siteName || !parentDir || !aiConfigured}
                                                className="w-full bg-accent text-black hover:bg-accent/80 py-4 rounded-sm font-medium transition-all flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed uppercase tracking-wider text-sm"
                                            >
                                                {isProcessing ? <Loader2 className="animate-spin" /> : <Wand2 size={18} />}
                                                <span>{isProcessing ? 'Creating_Site...' : 'Create_With_AI'}</span>
                                            </button>

                                            {!aiConfigured && (
                                                <p className="text-center text-xs text-zinc-500 font-mono">
                                                    Configure AI credentials in the <button onClick={() => setActiveTab('ai')} className="text-accent hover:underline">AI tab</button> first
                                                </p>
                                            )}
                                        </div>
                                    </Card>
                                </motion.div>
                            )}
                        </motion.div>
                    </AnimatePresence>
                </main>
            </div>
        </div>
    )
}

export default App
