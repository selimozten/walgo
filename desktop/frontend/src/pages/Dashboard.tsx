import { useState, useEffect, useRef, useMemo } from 'react';
import {
    Zap, Workflow, Wallet, ChevronDown, Check, Plus, Database, Sparkles, Activity, Import as ImportIcon, Copy, AlertCircle, RefreshCw
} from 'lucide-react';
import { Card } from '../components/ui/Card';
import { itemVariants, buttonVariants, iconButtonVariants } from '../utils/constants';
import { WalletInfo, SystemHealth } from '../types';
import { motion } from 'framer-motion';
import { cn } from '../utils/helpers';
import { useAIProgress } from '../contexts/AIProgressContext';

// Constants
const NETWORKS = [
    { id: 'testnet', label: 'Testnet', color: 'text-yellow-400' },
    { id: 'mainnet', label: 'Mainnet', color: 'text-green-400' },
] as const;

const COPY_FEEDBACK_DURATION = 5000; // ms - increased for better visibility
const DROPDOWN_OFFSET = 4; // px

interface DashboardProps {
    version: string;
    walletInfo: WalletInfo | null;
    addressList?: string[];
    aiConfigured?: boolean;
    systemHealth?: SystemHealth;
    onNavigate: (tab: string) => void;
    onSwitchNetwork?: (network: string) => void;
    onSwitchAccount?: (address: string) => void;
    onCreateAccount?: () => void;
    onImportAccount?: () => void;
    onStatusChange?: (status: { type: 'success' | 'error' | 'info'; message: string }) => void;
    onRefreshHealth?: () => Promise<void>;
}

export const Dashboard: React.FC<DashboardProps> = ({
    version,
    walletInfo,
    addressList = [],
    aiConfigured = false,
    systemHealth,
    onNavigate,
    onSwitchNetwork,
    onSwitchAccount,
    onCreateAccount,
    onImportAccount,
    onRefreshHealth,
    onStatusChange
}) => {
    const { progressState } = useAIProgress();
    const hugoInstalled = systemHealth?.hugoInstalled ?? false;
    const [showNetworkMenu, setShowNetworkMenu] = useState(false);
    const [showAccountMenu, setShowAccountMenu] = useState(false);
    const [copiedAddress, setCopiedAddress] = useState(false);
    const [accountDropdownPosition, setAccountDropdownPosition] = useState({ top: 0, left: 0, width: 0 });
    const [networkDropdownPosition, setNetworkDropdownPosition] = useState({ top: 0, left: 0, width: 0 });
    const networkButtonRef = useRef<HTMLButtonElement>(null);
    const accountButtonRef = useRef<HTMLButtonElement>(null);
    const networkDropdownRef = useRef<HTMLDivElement>(null);
    const accountDropdownRef = useRef<HTMLDivElement>(null);

    const currentNetwork = NETWORKS.find(n => n.id === walletInfo?.network) || NETWORKS[0];

    const sortedAddressList = useMemo(() => {
        return [...addressList].sort((a, b) => {
            if (a === walletInfo?.address) return -1;
            if (b === walletInfo?.address) return 1;
            return 0;
        });
    }, [addressList, walletInfo?.address]);

    // Close dropdowns when clicking outside
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (
                networkButtonRef.current && 
                !networkButtonRef.current.contains(event.target as Node) &&
                networkDropdownRef.current &&
                !networkDropdownRef.current.contains(event.target as Node)
            ) {
                setShowNetworkMenu(false);
            }
            if (
                accountButtonRef.current && 
                !accountButtonRef.current.contains(event.target as Node) &&
                accountDropdownRef.current &&
                !accountDropdownRef.current.contains(event.target as Node)
            ) {
                setShowAccountMenu(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const updateDropdownPosition = (buttonRef: React.RefObject<HTMLButtonElement>, setPosition: (pos: { top: number; left: number; width: number }) => void) => {
        if (buttonRef.current) {
            const rect = buttonRef.current.getBoundingClientRect();
            setPosition({
                top: rect.bottom + window.scrollY + DROPDOWN_OFFSET,
                left: rect.left + window.scrollX,
                width: rect.width,
            });
        }
    };

    const handleNetworkMenuToggle = () => {
        if (!showNetworkMenu) {
            updateDropdownPosition(networkButtonRef, setNetworkDropdownPosition);
        }
        setShowNetworkMenu(!showNetworkMenu);
    };

    const handleAccountMenuToggle = () => {
        if (!showAccountMenu) {
            updateDropdownPosition(accountButtonRef, setAccountDropdownPosition);
        }
        setShowAccountMenu(!showAccountMenu);
    };

    const handleCopyAddress = async () => {
        if (walletInfo?.address) {
            try {
                await navigator.clipboard.writeText(walletInfo.address);
                setCopiedAddress(true);
                setTimeout(() => setCopiedAddress(false), COPY_FEEDBACK_DURATION);
            } catch (err) {
                console.error('Failed to copy address:', err);
            }
        }
    };

    return (
        <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{
                duration: 0.4,
                ease: [0.4, 0, 0.2, 1]
            }}
            className="space-y-6"
        >
            {/* Header */}
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-white mb-1">Dashboard</h1>
                    <p className="text-zinc-400 text-sm">Welcome to Walgo Desktop</p>
                </div>

                {/* Refresh Button */}
                <button
                    onClick={() => window.location.reload()}
                    className="p-2.5 bg-white/5 hover:bg-white/10 border border-white/10 hover:border-accent/30 rounded-sm transition-all group"
                    title="Refresh Dashboard"
                >
                    <RefreshCw size={18} className="text-zinc-400 group-hover:text-accent transition-colors" />
                </button>
            </div>

            {/* Wallet Info Section */}
            <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.1, duration: 0.4, ease: [0.4, 0, 0.2, 1] }}
            >
            <Card className="bg-gradient-to-br from-accent/5 to-transparent border-accent/20">
                <div className="flex items-center gap-2 mb-4 border-b border-white/5 pb-3">
                    <Wallet size={20} className="text-accent" />
                    <h2 className="text-lg font-semibold text-white">Wallet Overview</h2>
                </div>

                {walletInfo?.address ? (
                    <div className="space-y-4">
                        {/* Current Address with Copy */}
                        <div>
                            <div className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest mb-2">
                                Current Address
                            </div>
                            <div className="bg-black/20 border border-white/10 rounded-sm px-3 py-2.5 flex items-center justify-between gap-2">
                                <span className="text-sm font-mono text-white truncate">
                                    {walletInfo?.address}
                                </span>
                                <motion.button
                                    onClick={handleCopyAddress}
                                    className="flex-shrink-0 p-1.5 hover:bg-white/5 rounded-sm transition-colors"
                                    title="Copy address"
                                    variants={iconButtonVariants}
                                    whileHover="hover"
                                    whileTap="tap"
                                >
                                    {copiedAddress ? (
                                        <Check size={16} className="text-green-400" />
                                    ) : (
                                        <Copy size={16} className="text-zinc-400" />
                                    )}
                                </motion.button>
                            </div>
                        </div>

                        {/* Network and Account Switchers */}
                        <div className="grid grid-cols-2 gap-4">
                            {/* Network Switcher */}
                            <div>
                                <div className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest mb-2">
                                    Network
                                </div>
                                <motion.button
                                    ref={networkButtonRef}
                                    onClick={handleNetworkMenuToggle}
                                    className="w-full bg-black/20 hover:bg-black/30 border border-white/10 rounded-sm px-3 py-2.5 text-sm font-mono text-white flex items-center justify-between transition-all"
                                    variants={buttonVariants}
                                    whileHover="hover"
                                    whileTap="tap"
                                >
                                    <span className={currentNetwork.color}>{currentNetwork.label}</span>
                                    <ChevronDown size={14} className={`transition-transform ${showNetworkMenu ? 'rotate-180' : ''}`} />
                                </motion.button>
                            </div>

                            {/* Account Switcher */}
                            <div>
                                <div className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest mb-2">
                                    Switch Account
                                </div>
                                <motion.button
                                    ref={accountButtonRef}
                                    onClick={handleAccountMenuToggle}
                                    className="w-full bg-black/20 hover:bg-black/30 border border-white/10 rounded-sm px-3 py-2.5 text-sm font-mono text-white flex items-center justify-between transition-all"
                                    variants={buttonVariants}
                                    whileHover="hover"
                                    whileTap="tap"
                                >
                                    <span className="text-white">
                                        {walletInfo?.address?.slice(0, 8)}...{walletInfo?.address?.slice(-6)}
                                    </span>
                                    <ChevronDown size={14} className={`transition-transform ${showAccountMenu ? 'rotate-180' : ''}`} />
                                </motion.button>
                            </div>
                        </div>

                        {/* Balances */}
                        <div className="grid grid-cols-2 gap-4">
                            {/* SUI Balance */}
                            <div>
                                <div className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest mb-2">
                                    SUI Balance
                                </div>
                                <div className="bg-black/20 p-3 rounded-sm">
                                    <div className="text-sm font-mono text-accent font-semibold">
                                        {walletInfo.suiBalance || 0} SUI
                                    </div>
                                </div>
                            </div>

                            {/* WAL Balance */}
                            <div>
                                <div className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest mb-2">
                                    WAL Balance
                                </div>
                                <div className="bg-black/20 p-3 rounded-sm">
                                    <div className="text-sm font-mono text-accent font-semibold">
                                        {walletInfo.walBalance || 0} WAL
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                ) : (
                    <div className="text-center py-8">
                        <Wallet size={48} className="mx-auto mb-3 text-zinc-700" />
                        <p className="text-zinc-500 text-sm font-mono">No wallet connected</p>
                        <p className="text-zinc-600 text-xs font-mono mt-1">
                            Configure your SUI wallet in settings
                        </p>
                    </div>
                )}
            </Card>
            </motion.div>

            {/* Quick Actions */}
            <motion.div 
                className="grid grid-cols-2 gap-4"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.2, duration: 0.4, ease: [0.4, 0, 0.2, 1] }}
            >
                {/* QuickStart Card */}
                <Card
                    className={cn(
                        "relative overflow-hidden transition-all group",
                        !hugoInstalled ? "opacity-60 cursor-not-allowed" : "cursor-pointer hover:bg-white/5"
                    )}
                    onClick={() => {
                        if (!hugoInstalled) {
                            if (onStatusChange) {
                                onStatusChange({
                                    type: 'error',
                                    message: 'Hugo is required to create sites. Please install Hugo from System Health page.'
                                });
                            }
                            return;
                        }
                        onNavigate('quickstart');
                    }}
                >
                    {!hugoInstalled && (
                        <div className="absolute top-2 right-2 z-10">
                            <AlertCircle size={20} className="text-red-400" />
                        </div>
                    )}
                    <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                        <Zap size={64} className={cn(hugoInstalled ? "text-accent" : "text-zinc-600")} />
                    </div>
                    <div className="flex items-start gap-3">
                        <div className={cn(
                            "p-2 rounded-sm transition-colors",
                            hugoInstalled ? "bg-accent/10 group-hover:bg-accent/20" : "bg-zinc-800/50"
                        )}>
                            <Zap size={24} className={cn(hugoInstalled ? "text-accent" : "text-zinc-600")} />
                        </div>
                        <div className="flex-1">
                            <div className="text-xs text-zinc-500 font-mono uppercase mb-1">Quick Action</div>
                            <div className={cn(
                                "text-lg font-semibold transition-colors",
                                hugoInstalled ? "text-white group-hover:text-accent" : "text-zinc-600"
                            )}>
                                QuickStart
                            </div>
                            <div className="text-xs text-zinc-500 font-mono mt-1">
                                {!hugoInstalled ? "Hugo Required" : "Create site instantly"}
                            </div>
                        </div>
                    </div>

                    <div className="mt-4 pt-4 border-t border-white/5">
                        <div className={cn(
                            "text-[10px] font-mono uppercase tracking-wider",
                            hugoInstalled ? "text-accent" : "text-zinc-600"
                        )}>
                            {!hugoInstalled ? "Install Hugo first" : "Click to start →"}
                        </div>
                    </div>
                </Card>

                {/* AI Pipeline Card */}
                <Card
                    className={cn(
                        "relative overflow-hidden transition-all group",
                        (!aiConfigured || progressState.isActive || !hugoInstalled)
                            ? "opacity-60 cursor-not-allowed"
                            : "cursor-pointer hover:bg-white/5"
                    )}
                    onClick={() => {
                        if (!hugoInstalled) {
                            if (onStatusChange) {
                                onStatusChange({
                                    type: 'error',
                                    message: 'Hugo is required to create sites. Please install Hugo from System Health page.'
                                });
                            }
                            return;
                        }
                        if (progressState.isActive) {
                            if (onStatusChange) {
                                onStatusChange({
                                    type: 'info',
                                    message: 'AI in progress: AI is currently creating a site. Please wait until it finishes.'
                                });
                            }
                        } else if (aiConfigured) {
                            onNavigate('ai-create-site');
                        } else if (onStatusChange) {
                            onStatusChange({
                                type: 'error',
                                message: 'AI not configured: Please configure AI in Settings to use AI features.'
                            });
                        }
                    }}
                >
                    {!hugoInstalled && (
                        <div className="absolute top-2 right-2 z-10">
                            <AlertCircle size={20} className="text-red-400" />
                        </div>
                    )}
                    <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                        <Workflow size={64} className={
                            aiConfigured && !progressState.isActive && hugoInstalled ? "text-purple-400" : "text-zinc-600"
                        } />
                    </div>
                    <div className="flex items-start gap-3">
                        <div className={cn(
                            "p-2 rounded-sm transition-colors",
                            aiConfigured && !progressState.isActive && hugoInstalled
                                ? "bg-purple-500/10 group-hover:bg-purple-500/20"
                                : "bg-zinc-800/50"
                        )}>
                            <Workflow size={24} className={
                                aiConfigured && !progressState.isActive && hugoInstalled ? "text-purple-400" : "text-zinc-600"
                            } />
                        </div>
                        <div className="flex-1">
                            <div className="text-xs text-zinc-500 font-mono uppercase mb-1">AI Powered</div>
                            <div className={cn(
                                "text-lg font-semibold transition-colors",
                                aiConfigured && !progressState.isActive
                                    ? "text-white group-hover:text-purple-400"
                                    : "text-zinc-600"
                            )}>
                                AI Pipeline
                            </div>
                            <div className="text-xs text-zinc-500 font-mono mt-1">
                                {progressState.isActive
                                    ? "AI is busy creating a site..."
                                    : "Full AI site creation"
                                }
                            </div>
                        </div>
                    </div>

                    <div className="mt-4 pt-4 border-t border-white/5">
                        <div className={cn(
                            "text-[10px] font-mono uppercase tracking-wider",
                            progressState.isActive ? "text-zinc-600" : "text-purple-400"
                        )}>
                            {progressState.isActive ? "Please wait..." : "Click to start →"}
                        </div>
                    </div>
                </Card>
            </motion.div>

            {/* Additional Actions */}
            <motion.div 
                className="grid grid-cols-3 gap-4"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.3, duration: 0.4, ease: [0.4, 0, 0.2, 1] }}
            >
                {/* Projects Card */}
                <Card
                    className="relative overflow-hidden cursor-pointer hover:bg-white/5 transition-all group"
                    onClick={() => onNavigate('projects')}
                >
                    <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                        <Database size={64} className="text-blue-400" />
                    </div>
                    <div className="flex items-start gap-3">
                        <div className="p-2 rounded-sm bg-blue-500/10 group-hover:bg-blue-500/20 transition-colors">
                            <Database size={24} className="text-blue-400" />
                        </div>
                        <div className="flex-1">
                            <div className="text-xs text-zinc-500 font-mono uppercase mb-1">Manage</div>
                            <div className="text-lg font-semibold text-white group-hover:text-blue-400 transition-colors">
                                Projects
                            </div>
                            <div className="text-xs text-zinc-500 font-mono mt-1">
                                View all projects
                            </div>
                        </div>
                    </div>

                    <div className="mt-4 pt-4 border-t border-white/5">
                        <div className="text-[10px] font-mono text-blue-400 uppercase tracking-wider">
                            View Projects →
                        </div>
                    </div>
                </Card>

                {/* AI Config Card */}
                <Card
                    className="relative overflow-hidden cursor-pointer hover:bg-white/5 transition-all group"
                    onClick={() => onNavigate('ai')}
                >
                    <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                        <Sparkles size={64} className="text-pink-400" />
                    </div>
                    <div className="flex items-start gap-3">
                        <div className="p-2 rounded-sm bg-pink-500/10 group-hover:bg-pink-500/20 transition-colors">
                            <Sparkles size={24} className="text-pink-400" />
                        </div>
                        <div className="flex-1">
                            <div className="text-xs text-zinc-500 font-mono uppercase mb-1">Configure</div>
                            <div className="text-lg font-semibold text-white group-hover:text-pink-400 transition-colors">
                                AI Config
                            </div>
                            <div className="text-xs text-zinc-500 font-mono mt-1">
                                Manage AI settings
                            </div>
                        </div>
                    </div>

                    <div className="mt-4 pt-4 border-t border-white/5">
                        <div className="text-[10px] font-mono text-pink-400 uppercase tracking-wider">
                            Configure →
                        </div>
                    </div>
                </Card>

                {/* System Health Card */}
                <Card
                    className="relative overflow-hidden cursor-pointer hover:bg-white/5 transition-all group"
                    onClick={() => onNavigate('system-health')}
                >
                    <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                        <Activity size={64} className="text-green-400" />
                    </div>
                    <div className="flex items-start gap-3">
                        <div className="p-2 rounded-sm bg-green-500/10 group-hover:bg-green-500/20 transition-colors">
                            <Activity size={24} className="text-green-400" />
                        </div>
                        <div className="flex-1">
                            <div className="text-xs text-zinc-500 font-mono uppercase mb-1">System</div>
                            <div className="text-lg font-semibold text-white group-hover:text-green-400 transition-colors">
                                System Health
                            </div>
                            <div className="text-xs text-zinc-500 font-mono mt-1">
                                Check dependencies
                            </div>
                        </div>
                    </div>

                    <div className="mt-4 pt-4 border-t border-white/5">
                        <div className="text-[10px] font-mono text-green-400 uppercase tracking-wider">
                            Check Now →
                        </div>
                    </div>
                </Card>
            </motion.div>

        {/* Fixed positioned dropdowns */}
        {showNetworkMenu && (
            <div
                ref={networkDropdownRef}
                style={{
                    position: 'fixed',
                    top: `${networkDropdownPosition.top}px`,
                    left: `${networkDropdownPosition.left}px`,
                    width: `${networkDropdownPosition.width}px`,
                    zIndex: 1000,
                }}
            >
                <div className="bg-zinc-900 border border-white/10 rounded-sm shadow-xl">
                    {NETWORKS.map((network) => (
                        <motion.button
                            key={network.id}
                            onClick={() => {
                                onSwitchNetwork?.(network.id);
                                setShowNetworkMenu(false);
                            }}
                            className="w-full px-3 py-2 text-sm font-mono text-left hover:bg-white/5 transition-colors flex items-center justify-between"
                            variants={buttonVariants}
                            whileHover="hover"
                            whileTap="tap"
                        >
                            <span className={network.color}>{network.label}</span>
                            {network.id === walletInfo?.network && (
                                <Check size={14} className="text-accent" />
                            )}
                        </motion.button>
                    ))}
                </div>
            </div>
        )}

        {showAccountMenu && (
            <div
                ref={accountDropdownRef}
                style={{
                    position: 'fixed',
                    top: `${accountDropdownPosition.top}px`,
                    left: `${accountDropdownPosition.left}px`,
                    width: `${accountDropdownPosition.width}px`,
                    zIndex: 1000,
                }}
            >
                <div className="bg-zinc-900 border border-white/10 rounded-sm shadow-xl">
                    {onCreateAccount && (
                        <motion.button
                            onClick={() => {
                                onCreateAccount?.();
                                setShowAccountMenu(false);
                            }}
                            className="w-full px-3 py-2 text-sm font-mono text-left hover:bg-white/5 transition-colors flex items-center gap-2 text-accent border-b border-white/5"
                            variants={buttonVariants}
                            whileHover="hover"
                            whileTap="tap"
                        >
                            <Plus size={14} />
                            <span>Create New Account</span>
                        </motion.button>
                    )}
                    {onImportAccount && (
                        <motion.button
                            onClick={() => {
                                onImportAccount?.();
                                setShowAccountMenu(false);
                            }}
                            className="w-full px-3 py-2 text-sm font-mono text-left hover:bg-white/5 transition-colors flex items-center gap-2 text-accent border-b border-white/5"
                            variants={buttonVariants}
                            whileHover="hover"
                            whileTap="tap"
                        >
                            <ImportIcon size={14} />
                            <span>Import Account</span>
                        </motion.button>
                    )}
                    <div className="h-px bg-white/10 my-1" />
                    <div className="max-h-64 overflow-y-auto">
                        {sortedAddressList.map((address) => (
                            <motion.button
                                key={address}
                                onClick={() => {
                                    onSwitchAccount?.(address);
                                    setShowAccountMenu(false);
                                }}
                                className={`w-full px-3 py-2 text-sm font-mono text-left hover:bg-white/5 transition-colors flex items-center justify-between ${
                                    address === walletInfo?.address ? 'bg-accent/10' : ''
                                }`}
                                variants={buttonVariants}
                                whileHover="hover"
                                whileTap="tap"
                            >
                                <span className={address === walletInfo?.address ? 'text-accent font-semibold' : 'text-white'}>
                                    {address.slice(0, 8)}...{address.slice(-6)}
                                </span>
                                {address === walletInfo?.address && (
                                    <Check size={14} className="text-accent" />
                                )}
                            </motion.button>
                        ))}
                    </div>
                </div>
            </div>
        )}
        </motion.div>
    );
};
