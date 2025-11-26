import { useState, useEffect } from 'react';
import { CreateSite, BuildSite, DeploySite } from '../wailsjs/go/main/App';
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
    Zap
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

    useEffect(() => {
        setStatus(null);
    }, [activeTab]);

    const addLog = (msg: string) => {
        setLogs(prev => [...prev, `[${new Date().toLocaleTimeString()}] ${msg}`]);
    };

    const handleCreate = async () => {
        if (!siteName) return;
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'Initializing...' });
        addLog(`Initializing project: ${siteName}`);
        try {
            const cwd = ".";
            await CreateSite(cwd, siteName);
            setStatus({ type: 'success', message: `Project '${siteName}' created` });
            addLog(`Success: Project created at ${cwd}/${siteName}`);
            setSitePath(`${cwd}/${siteName}`);
            setTimeout(() => setActiveTab('manage'), 1000);
        } catch (err) {
            setStatus({ type: 'error', message: `Failed: ${err}` });
            addLog(`Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleBuild = async () => {
        if (!sitePath) return;
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'Building...' });
        addLog(`Starting build for ${sitePath}`);
        try {
            await BuildSite(sitePath);
            setStatus({ type: 'success', message: 'Build completed' });
            addLog('Build completed successfully');
        } catch (err) {
            setStatus({ type: 'error', message: `Build failed: ${err}` });
            addLog(`Build Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const handleDeploy = async () => {
        if (!sitePath) return;
        setIsProcessing(true);
        setStatus({ type: 'info', message: 'Deploying...' });
        addLog('Initiating deployment to Walrus...');
        try {
            const result = await DeploySite(sitePath, 1);
            if (result.success) {
                setStatus({ type: 'success', message: `Deployed: ${result.objectId}` });
                addLog(`Deployment Success: ${result.objectId}`);
            } else {
                setStatus({ type: 'error', message: `Failed: ${result.error}` });
                addLog(`Deployment Failed: ${result.error}`);
            }
        } catch (err) {
            setStatus({ type: 'error', message: `Error: ${err}` });
            addLog(`System Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
    };

    const NavItem = ({ id, icon: Icon, label }: { id: string, icon: any, label: string }) => (
        <button
            onClick={() => setActiveTab(id)}
            className={cn(
                "group flex flex-col items-center justify-center w-16 h-16 rounded-2xl transition-all duration-300 relative",
                activeTab === id
                    ? "bg-white text-black shadow-[0_0_20px_rgba(255,255,255,0.3)]"
                    : "text-zinc-500 hover:text-zinc-200 hover:bg-zinc-900"
            )}
        >
            <Icon size={24} strokeWidth={1.5} />
            {activeTab === id && (
                <motion.div
                    layoutId="active-dot"
                    className="absolute -right-2 w-1 h-8 bg-white rounded-full"
                />
            )}
        </button>
    );

    const Card = ({ children, className, onClick }: { children: React.ReactNode, className?: string, onClick?: () => void }) => (
        <div
            onClick={onClick}
            className={cn(
                "glass-panel rounded-3xl p-6 transition-all duration-300 group relative overflow-hidden",
                onClick && "cursor-pointer hover:border-white/20 hover:bg-zinc-900/60",
                className
            )}
        >
            <div className="absolute inset-0 bg-gradient-to-br from-white/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
            <div className="relative z-10">{children}</div>
        </div>
    );

    return (
        <div className="min-h-screen bg-black flex font-sans text-zinc-200 selection:bg-white/20 overflow-hidden bg-grid-pattern">
            {/* Navigation Rail */}
            <div className="w-24 border-r border-white/5 flex flex-col items-center py-8 z-20 bg-black/50 backdrop-blur-xl">
                <div className="mb-12">
                    <div className="w-12 h-12 bg-zinc-900 rounded-xl flex items-center justify-center border border-white/10">
                        <Globe className="text-white" size={24} />
                    </div>
                </div>

                <nav className="space-y-6 flex-1">
                    <NavItem id="dashboard" icon={LayoutGrid} label="Overview" />
                    <NavItem id="create" icon={Plus} label="New" />
                    <NavItem id="manage" icon={Settings} label="Manage" />
                </nav>

                <div className="mt-auto">
                    <div className="w-10 h-10 rounded-full bg-zinc-900 border border-white/10 flex items-center justify-center text-xs font-medium text-zinc-400">
                        W
                    </div>
                </div>
            </div>

            {/* Main Content */}
            <div className="flex-1 flex flex-col h-screen overflow-hidden relative">
                <header className="px-10 py-8 flex justify-between items-end z-10">
                    <div>
                        <h1 className="text-4xl font-light tracking-tight text-white mb-2">
                            {activeTab === 'dashboard' && 'Overview'}
                            {activeTab === 'create' && 'Initialize'}
                            {activeTab === 'manage' && 'Operations'}
                        </h1>
                        <div className="flex items-center gap-2 text-zinc-500 text-sm font-mono">
                            <span>WALGO</span>
                            <ChevronRight size={12} />
                            <span className="uppercase">{activeTab}</span>
                        </div>
                    </div>

                    <div className="flex items-center gap-4">
                        <div className="px-4 py-2 rounded-full bg-zinc-900/50 border border-white/5 flex items-center gap-2 text-xs font-mono text-zinc-400">
                            <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                            SYSTEM ONLINE
                        </div>
                    </div>
                </header>

                <main className="flex-1 overflow-y-auto px-10 pb-10">
                    <AnimatePresence mode="wait">
                        <motion.div
                            key={activeTab}
                            initial={{ opacity: 0, scale: 0.98 }}
                            animate={{ opacity: 1, scale: 1 }}
                            exit={{ opacity: 0, scale: 1.02 }}
                            transition={{ duration: 0.2, ease: "easeOut" }}
                            className="h-full"
                        >
                            {activeTab === 'dashboard' && (
                                <div className="grid grid-cols-3 gap-6 h-full max-h-[600px]">
                                    <Card onClick={() => setActiveTab('create')} className="col-span-2 row-span-2 flex flex-col justify-between bg-zinc-900/20">
                                        <div className="flex justify-between items-start">
                                            <div className="p-3 bg-white/5 rounded-xl">
                                                <Plus size={32} className="text-white" />
                                            </div>
                                            <ArrowUpRight className="text-zinc-600 group-hover:text-white transition-colors" />
                                        </div>
                                        <div>
                                            <h2 className="text-3xl font-light text-white mb-2">New Project</h2>
                                            <p className="text-zinc-500 max-w-md">Initialize a new static site structure. Configured for high-performance decentralized hosting.</p>
                                        </div>
                                    </Card>

                                    <Card className="col-span-1 bg-zinc-900/20">
                                        <div className="flex items-center gap-3 mb-4 text-zinc-400">
                                            <Activity size={20} />
                                            <span className="text-sm font-mono uppercase">Status</span>
                                        </div>
                                        <div className="space-y-4">
                                            <div className="flex justify-between items-center">
                                                <span className="text-sm text-zinc-500">Walrus Network</span>
                                                <span className="text-xs font-mono text-emerald-500">CONNECTED</span>
                                            </div>
                                            <div className="flex justify-between items-center">
                                                <span className="text-sm text-zinc-500">Local Daemon</span>
                                                <span className="text-xs font-mono text-emerald-500">ACTIVE</span>
                                            </div>
                                            <div className="h-px bg-white/5 my-4" />
                                            <div className="text-center py-4">
                                                <div className="text-4xl font-light text-white mb-1">0</div>
                                                <div className="text-xs text-zinc-600 uppercase tracking-widest">Active Deployments</div>
                                            </div>
                                        </div>
                                    </Card>

                                    <Card onClick={() => setActiveTab('manage')} className="col-span-1 bg-zinc-900/20">
                                        <div className="flex justify-between items-start mb-8">
                                            <div className="p-3 bg-white/5 rounded-xl">
                                                <Command size={24} className="text-white" />
                                            </div>
                                            <ArrowUpRight className="text-zinc-600 group-hover:text-white transition-colors" />
                                        </div>
                                        <h3 className="text-xl text-white mb-1">Manage</h3>
                                        <p className="text-sm text-zinc-500">Build & Deploy existing sites</p>
                                    </Card>
                                </div>
                            )}

                            {activeTab === 'create' && (
                                <div className="max-w-2xl mx-auto pt-10">
                                    <Card className="border-white/10">
                                        <div className="flex items-center gap-4 mb-8">
                                            <div className="w-12 h-12 bg-white text-black rounded-xl flex items-center justify-center">
                                                <Plus size={24} />
                                            </div>
                                            <div>
                                                <h2 className="text-xl font-medium text-white">Project Initialization</h2>
                                                <p className="text-zinc-500 text-sm">Configure your new workspace</p>
                                            </div>
                                        </div>

                                        <div className="space-y-8">
                                            <div className="group">
                                                <label className="block text-xs font-mono text-zinc-500 uppercase mb-3 ml-1">Project Name</label>
                                                <input
                                                    type="text"
                                                    value={siteName}
                                                    onChange={(e) => setSiteName(e.target.value)}
                                                    className="w-full bg-black/50 border border-white/10 rounded-xl px-6 py-4 text-lg text-white placeholder:text-zinc-800 focus:outline-none focus:border-white/30 transition-all font-light"
                                                    placeholder="project-alpha"
                                                    autoFocus
                                                />
                                            </div>

                                            <button
                                                onClick={handleCreate}
                                                disabled={isProcessing || !siteName}
                                                className="w-full bg-white text-black hover:bg-zinc-200 py-4 rounded-xl font-medium transition-all flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed"
                                            >
                                                {isProcessing ? <Loader2 className="animate-spin" /> : <Zap size={20} />}
                                                <span>Initialize System</span>
                                            </button>
                                        </div>
                                    </Card>
                                </div>
                            )}

                            {activeTab === 'manage' && (
                                <div className="grid grid-cols-3 gap-6 h-full">
                                    <div className="col-span-2 space-y-6">
                                        <Card className="border-white/10">
                                            <div className="flex items-center gap-3 mb-6">
                                                <Folder size={18} className="text-zinc-500" />
                                                <span className="text-sm font-mono text-zinc-500 uppercase">Target Directory</span>
                                            </div>
                                            <input
                                                type="text"
                                                value={sitePath}
                                                onChange={(e) => setSitePath(e.target.value)}
                                                className="w-full bg-black/50 border border-white/10 rounded-xl px-5 py-3 text-sm font-mono text-zinc-300 focus:outline-none focus:border-white/30 transition-all"
                                                placeholder="/path/to/site"
                                            />
                                        </Card>

                                        <div className="grid grid-cols-2 gap-6">
                                            <button
                                                onClick={handleBuild}
                                                disabled={isProcessing || !sitePath}
                                                className="glass-panel p-8 rounded-3xl text-left hover:bg-zinc-900/60 transition-all group border-white/5 hover:border-white/20 disabled:opacity-50"
                                            >
                                                <Cpu size={32} className="text-zinc-600 group-hover:text-white mb-4 transition-colors" />
                                                <h3 className="text-lg font-medium text-white mb-1">Build</h3>
                                                <p className="text-xs text-zinc-500 font-mono">COMPILE ASSETS</p>
                                            </button>

                                            <button
                                                onClick={handleDeploy}
                                                disabled={isProcessing || !sitePath}
                                                className="glass-panel p-8 rounded-3xl text-left hover:bg-zinc-900/60 transition-all group border-white/5 hover:border-white/20 disabled:opacity-50"
                                            >
                                                <HardDrive size={32} className="text-zinc-600 group-hover:text-white mb-4 transition-colors" />
                                                <h3 className="text-lg font-medium text-white mb-1">Deploy</h3>
                                                <p className="text-xs text-zinc-500 font-mono">UPLOAD TO WALRUS</p>
                                            </button>
                                        </div>
                                    </div>

                                    <div className="col-span-1 glass-panel rounded-3xl border-white/5 flex flex-col overflow-hidden">
                                        <div className="p-4 border-b border-white/5 bg-white/[0.02] flex justify-between items-center">
                                            <div className="flex items-center gap-2 text-xs font-mono text-zinc-400">
                                                <Terminal size={12} />
                                                <span>OUTPUT_LOG</span>
                                            </div>
                                            {status && (
                                                <div className={cn(
                                                    "w-2 h-2 rounded-full animate-pulse",
                                                    status.type === 'success' ? "bg-emerald-500" :
                                                        status.type === 'error' ? "bg-red-500" : "bg-blue-500"
                                                )} />
                                            )}
                                        </div>
                                        <div className="flex-1 p-4 font-mono text-xs text-zinc-500 overflow-y-auto space-y-2">
                                            <div className="text-zinc-700">System ready...</div>
                                            {logs.map((log, i) => (
                                                <div key={i} className="text-zinc-400 border-l-2 border-white/10 pl-2">
                                                    {log}
                                                </div>
                                            ))}
                                            {status && (
                                                <div className={cn(
                                                    "pt-2 font-bold",
                                                    status.type === 'success' ? "text-emerald-400" :
                                                        status.type === 'error' ? "text-red-400" : "text-blue-400"
                                                )}>
                                                    {`> ${status.message}`}
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                </div>
                            )}
                        </motion.div>
                    </AnimatePresence>
                </main>
            </div>
        </div>
    )
}

export default App
