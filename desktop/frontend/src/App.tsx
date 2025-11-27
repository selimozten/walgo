import { useState, useEffect } from 'react';
import { CreateSite, BuildSite, DeploySite } from '../wailsjs/go/main/App';
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
    X
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
        setStatus({ type: 'info', message: 'INITIALIZING_SEQUENCE...' });
        addLog(`Initializing project: ${siteName}`);
        try {
            const cwd = ".";
            await CreateSite(cwd, siteName);
            setStatus({ type: 'success', message: `PROJECT_CREATED: ${siteName}` });
            addLog(`Success: Project created at ${cwd}/${siteName}`);
            setSitePath(`${cwd}/${siteName}`);
            setTimeout(() => setActiveTab('manage'), 1000);
        } catch (err) {
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
        } catch (err) {
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
            } else {
                setStatus({ type: 'error', message: `DEPLOY_FAILED: ${result.error}` });
                addLog(`Deployment Failed: ${result.error}`);
            }
        } catch (err) {
            setStatus({ type: 'error', message: `SYSTEM_ERROR: ${err}` });
            addLog(`System Error: ${err}`);
        } finally {
            setIsProcessing(false);
        }
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
                    <NavItem id="dashboard" icon={LayoutGrid} label="Overview" />
                    <NavItem id="create" icon={Plus} label="New" />
                    <NavItem id="manage" icon={Settings} label="Manage" />
                </nav>

                <div className="mt-auto">
                    <div className="w-8 h-8 rounded-sm bg-zinc-900 border border-white/10 flex items-center justify-center text-[10px] font-mono text-zinc-500">
                        v1
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
                            <span>SYSTEMS</span>
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
                                                    <Plus size={20} className="text-accent" />
                                                </div>
                                                <ArrowUpRight className="text-zinc-600 group-hover:text-accent transition-colors" />
                                            </div>

                                            <div className="relative z-10">
                                                <h2 className="text-3xl font-display font-light text-white mb-2">INITIALIZE_PROJECT</h2>
                                                <p className="text-zinc-500 max-w-md font-light">Begin new static site sequence. Configure for decentralized distribution.</p>
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
                                                    <div className="text-4xl font-display font-light text-white mb-1">0</div>
                                                    <div className="text-[10px] text-zinc-600 uppercase tracking-widest">Active_Deployments</div>
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
                                <motion.div variants={itemVariants} className="max-w-2xl mx-auto pt-10">
                                    <Card className="border-white/10 bg-zinc-900/20">
                                        <div className="flex items-center gap-4 mb-8 border-b border-white/5 pb-6">
                                            <div className="w-10 h-10 bg-accent text-black rounded-sm flex items-center justify-center shadow-[0_0_15px_rgba(77,162,255,0.3)]">
                                                <Plus size={20} />
                                            </div>
                                            <div>
                                                <h2 className="text-xl font-display font-medium text-white">PROJECT_INIT</h2>
                                                <p className="text-zinc-500 text-xs font-mono">CONFIGURE_WORKSPACE_PARAMETERS</p>
                                            </div>
                                        </div>

                                        <div className="space-y-8">
                                            <div className="group">
                                                <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">Project_Designator</label>
                                                <input
                                                    type="text"
                                                    value={siteName}
                                                    onChange={(e) => setSiteName(e.target.value)}
                                                    className="w-full bg-black/40 border border-white/10 rounded-sm px-6 py-4 text-lg text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-display tracking-wide"
                                                    placeholder="PROJECT_ALPHA"
                                                    autoFocus
                                                />
                                            </div>

                                            <button
                                                onClick={handleCreate}
                                                disabled={isProcessing || !siteName}
                                                className="w-full bg-white text-black hover:bg-accent hover:text-black py-4 rounded-sm font-medium transition-all flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed uppercase tracking-wider text-sm"
                                            >
                                                {isProcessing ? <Loader2 className="animate-spin" /> : <Zap size={18} />}
                                                <span>Execute_Init</span>
                                            </button>
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
                                                </div>
                                                <input
                                                    type="text"
                                                    value={sitePath}
                                                    onChange={(e) => setSitePath(e.target.value)}
                                                    className="w-full bg-black/40 border border-white/10 rounded-sm px-5 py-3 text-sm font-mono text-zinc-300 focus:outline-none focus:border-accent/50 transition-all"
                                                    placeholder="/PATH/TO/SITE"
                                                />
                                            </Card>
                                        </motion.div>

                                        <div className="grid grid-cols-2 gap-6">
                                            <motion.div variants={itemVariants}>
                                                <button
                                                    onClick={handleBuild}
                                                    disabled={isProcessing || !sitePath}
                                                    className="w-full glass-panel-tech p-8 rounded-sm text-left hover:border-accent/30 transition-all group disabled:opacity-50 relative overflow-hidden"
                                                >
                                                    <div className="absolute inset-0 bg-accent/5 translate-y-full group-hover:translate-y-0 transition-transform duration-500" />
                                                    <Cpu size={32} className="text-zinc-600 group-hover:text-accent mb-4 transition-colors relative z-10" />
                                                    <h3 className="text-lg font-display font-medium text-white mb-1 relative z-10">BUILD</h3>
                                                    <p className="text-[10px] text-zinc-500 font-mono relative z-10">COMPILE_ASSETS</p>
                                                </button>
                                            </motion.div>

                                            <motion.div variants={itemVariants}>
                                                <button
                                                    onClick={handleDeploy}
                                                    disabled={isProcessing || !sitePath}
                                                    className="w-full glass-panel-tech p-8 rounded-sm text-left hover:border-accent/30 transition-all group disabled:opacity-50 relative overflow-hidden"
                                                >
                                                    <div className="absolute inset-0 bg-accent/5 translate-y-full group-hover:translate-y-0 transition-transform duration-500" />
                                                    <HardDrive size={32} className="text-zinc-600 group-hover:text-accent mb-4 transition-colors relative z-10" />
                                                    <h3 className="text-lg font-display font-medium text-white mb-1 relative z-10">DEPLOY</h3>
                                                    <p className="text-[10px] text-zinc-500 font-mono relative z-10">UPLOAD_TO_WALRUS</p>
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
                        </motion.div>
                    </AnimatePresence>
                </main>
            </div>
        </div>
    )
}

export default App
