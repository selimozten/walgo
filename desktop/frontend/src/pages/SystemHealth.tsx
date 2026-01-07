import React, { useEffect, useState } from 'react';
import {
    Check, X, AlertCircle, Activity, Zap, Stethoscope, RefreshCw, Download, ArrowUp, Terminal
} from 'lucide-react';
import { Card } from '../components/ui/Card';
import { itemVariants, buttonVariants } from '../utils/constants';
import { SystemHealth as SystemHealthType } from '../types';
import { motion } from 'framer-motion';
import { useVersionCheck } from '../hooks';
import { InstallInstructionsModal } from '../components/modals';

interface SystemHealthProps {
    systemHealth: SystemHealthType | null;
    onCheckDeps: () => void;
    onRefresh?: () => void;
    installing?: any;
}

export const SystemHealth: React.FC<SystemHealthProps> = ({
    systemHealth,
    onCheckDeps,
    onRefresh,
    installing
}) => {
    const [showInstallModal, setShowInstallModal] = useState(false);
    const [missingTools, setMissingTools] = useState<string[]>([]);
    
    const { 
        versions, 
        loading: versionLoading, 
        updating, 
        checkVersions, 
        updateTool,
        hasUpdates,
        getSuiToolsWithUpdates
    } = useVersionCheck();

    // Load versions on mount
    useEffect(() => {
        if (systemHealth?.netOnline) {
            checkVersions();
        }
    }, [systemHealth?.netOnline, checkVersions]);

    const getStatusColor = (installed?: boolean) =>
        installed ? "text-accent" : "text-red-500";

    // Check if all dependencies are installed
    const allInstalled = systemHealth?.netOnline && 
                         systemHealth?.suiInstalled && 
                         systemHealth?.walrusInstalled && 
                         systemHealth?.siteBuilder && 
                         systemHealth?.hugoInstalled;

    // Get version info for a tool
    const getVersionInfo = (toolName: string) => {
        return versions.find(v => v.tool === toolName);
    };

    // Check if tool is being updated
    const isUpdating = (toolName: string) => {
        return updating.includes(toolName);
    };

    // Sui tools that need updates
    const suiToolsNeedingUpdate = getSuiToolsWithUpdates();

    return (
        <motion.div variants={itemVariants} className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-white mb-1">System Health</h1>
                    <p className="text-zinc-400 text-sm">Monitor and manage your Walgo dependencies</p>
                </div>
                <motion.button
                    onClick={() => {
                        onRefresh?.();
                        checkVersions();
                    }}
                    className="flex items-center gap-2 px-3 py-1.5 bg-zinc-800 hover:bg-zinc-700 rounded-sm text-sm text-zinc-300 transition-colors"
                    variants={buttonVariants}
                    whileHover="hover"
                    whileTap="tap"
                >
                    <RefreshCw size={14} />
                    Refresh
                </motion.button>
            </div>

            {/* Warning if Sui tools need updates */}
            {suiToolsNeedingUpdate.length > 0 && (
                <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-sm p-4">
                    <div className="flex items-start gap-3">
                        <AlertCircle size={20} className="text-yellow-500 flex-shrink-0 mt-0.5" />
                        <div className="flex-1">
                            <h3 className="text-sm font-semibold text-yellow-500 mb-1">
                                Updates Available for Deployment Tools
                            </h3>
                            <p className="text-xs text-zinc-300 mb-3">
                                The following tools have updates available: {suiToolsNeedingUpdate.join(', ')}. 
                                It's recommended to update before deploying to mainnet.
                            </p>
                            <button
                                onClick={async () => {
                                    for (const tool of suiToolsNeedingUpdate) {
                                        await updateTool(tool, 'mainnet');
                                    }
                                }}
                                disabled={updating.length > 0}
                                className="px-3 py-1.5 bg-yellow-500/20 hover:bg-yellow-500/30 text-yellow-500 border border-yellow-500/30 rounded-sm text-xs font-mono uppercase tracking-wider transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                            >
                                {updating.length > 0 ? (
                                    <>
                                        <Activity size={14} className="animate-spin" />
                                        Updating...
                                    </>
                                ) : (
                                    <>
                                        <Download size={14} />
                                        Update All Sui Tools
                                    </>
                                )}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* System Status */}
            <Card>
                <div className="flex items-center gap-2 mb-4">
                    <Stethoscope size={20} className="text-accent" />
                    <h2 className="text-lg font-semibold text-white">Dependencies Status</h2>
                </div>

                <div className="space-y-3">
                    {/* Network Status */}
                    <div className="bg-black/20 p-3 rounded-sm">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                                {systemHealth?.netOnline ? (
                                    <Check size={14} className="text-accent" />
                                ) : (
                                    <X size={14} className="text-red-500" />
                                )}
                                <span className="text-[10px] font-mono text-zinc-400 uppercase tracking-widest">Network</span>
                            </div>
                            <span className={`font-mono text-[10px] ${getStatusColor(systemHealth?.netOnline)}`}>
                                {systemHealth?.netOnline ? "ONLINE" : "OFFLINE"}
                            </span>
                        </div>
                    </div>

                    {/* Sui Status with Version */}
                    {renderToolStatus('sui', 'Sui CLI', systemHealth?.suiInstalled, getVersionInfo('sui'))}

                    {/* Walrus Status with Version */}
                    {renderToolStatus('walrus', 'Walrus CLI', systemHealth?.walrusInstalled, getVersionInfo('walrus'))}

                    {/* Site Builder Status with Version */}
                    {renderToolStatus('site-builder', 'Site Builder', systemHealth?.siteBuilder, getVersionInfo('site-builder'))}

                    <div className="h-px bg-white/5" />

                    {/* Hugo Status - No version check/update (managed via package manager) */}
                    <div className="bg-black/20 p-3 rounded-sm">
                        <div className="flex items-center justify-between mb-2">
                            <div className="flex items-center gap-2">
                                {systemHealth?.hugoInstalled ? (
                                    <Check size={14} className="text-accent" />
                                ) : (
                                    <X size={14} className="text-red-500" />
                                )}
                                <span className="text-[10px] font-mono text-zinc-400 uppercase tracking-widest">Hugo (Optional)</span>
                            </div>
                            <span className={`font-mono text-[10px] ${systemHealth?.hugoInstalled ? 'text-accent' : 'text-red-500'}`}>
                                {systemHealth?.hugoInstalled ? 'INSTALLED' : 'NOT INSTALLED'}
                            </span>
                        </div>
                        <p className="text-[10px] text-zinc-500 font-mono">
                            Install via package manager: brew install hugo (macOS) | apt install hugo (Linux)
                        </p>
                    </div>

                    <div className="h-px bg-white/5" />

                    {/* Overall Status */}
                    <div className="bg-accent/5 border border-accent/20 p-3 rounded-sm">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                                {allInstalled ? (
                                    <Check size={14} className="text-accent" />
                                ) : (
                                    <AlertCircle size={14} className="text-yellow-500" />
                                )}
                                <span className="text-[10px] font-mono text-zinc-400 uppercase tracking-widest">
                                    Overall Status
                                </span>
                            </div>
                            <span className={`font-mono text-[10px] ${
                                allInstalled ? "text-accent" : "text-yellow-500"
                            }`}>
                                {allInstalled ? "READY" : "SETUP REQUIRED"}
                            </span>
                        </div>
                    </div>

                    {/* Installation Instructions Button - Only show if something is missing */}
                    {!allInstalled && (
                        <motion.button
                            onClick={() => {
                                // Collect missing tools
                                const missing: string[] = [];
                                if (!systemHealth?.suiInstalled) missing.push('sui');
                                if (!systemHealth?.walrusInstalled) missing.push('walrus');
                                if (!systemHealth?.siteBuilder) missing.push('site-builder');
                                if (!systemHealth?.hugoInstalled) missing.push('hugo');
                                setMissingTools(missing);
                                setShowInstallModal(true);
                            }}
                            disabled={allInstalled}
                            className="w-full mt-4 px-4 py-3 bg-accent/10 hover:bg-accent/20 text-accent border border-accent/30 rounded-sm text-sm font-mono uppercase tracking-wider transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                            variants={buttonVariants}
                            whileHover="hover"
                            whileTap="tap"
                        >
                            <Terminal size={16} />
                            View Installation Instructions
                        </motion.button>
                    )}
                </div>
            </Card>

            {/* Installation Instructions Modal */}
            <InstallInstructionsModal
                isOpen={showInstallModal}
                onClose={() => setShowInstallModal(false)}
                missingDeps={missingTools}
                onRefreshStatus={onRefresh}
            />
        </motion.div>
    );

    // Helper function to render tool status with version info
    function renderToolStatus(
        toolName: string, 
        displayName: string, 
        installed?: boolean, 
        versionInfo?: any,
        isOptional: boolean = false
    ) {
        const hasUpdate = versionInfo?.updateRequired;
        const isToolUpdating = isUpdating(toolName);

        return (
            <div className={`bg-black/20 p-3 rounded-sm ${hasUpdate ? 'border border-yellow-500/20' : ''}`}>
                <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                        {installed ? (
                            <Check size={14} className="text-accent" />
                        ) : (
                            <X size={14} className="text-red-500" />
                        )}
                        <span className="text-[10px] font-mono text-zinc-400 uppercase tracking-widest">{displayName}</span>
                    </div>
                    <span className={`font-mono text-[10px] ${getStatusColor(installed)}`}>
                        {installed ? "INSTALLED" : "MISSING"}
                    </span>
                </div>

                {/* Version Info */}
                {installed && versionInfo && (
                    <div className="mt-2 pt-2 border-t border-white/5">
                        <div className="flex items-center justify-between text-[9px] font-mono">
                            <span className="text-zinc-500">
                                Current: <span className="text-zinc-400">{versionInfo.currentVersion}</span>
                            </span>
                            {versionInfo.latestVersion !== 'unknown' && (
                                <span className="text-zinc-500">
                                    Latest: <span className={hasUpdate ? "text-yellow-500" : "text-zinc-400"}>
                                        {versionInfo.latestVersion}
                                    </span>
                                </span>
                            )}
                        </div>

                        {/* Update Button */}
                        {hasUpdate && (
                            <button
                                onClick={() => updateTool(toolName, 'mainnet')}
                                disabled={isToolUpdating}
                                className={`w-full mt-2 px-2 py-1.5 rounded-sm text-[9px] font-mono uppercase tracking-wider transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-1.5 ${
                                    isOptional 
                                        ? 'bg-blue-500/10 hover:bg-blue-500/20 text-blue-400 border border-blue-500/30'
                                        : 'bg-yellow-500/10 hover:bg-yellow-500/20 text-yellow-500 border border-yellow-500/30'
                                }`}
                            >
                                {isToolUpdating ? (
                                    <>
                                        <Activity size={12} className="animate-spin" />
                                        Updating...
                                    </>
                                ) : (
                                    <>
                                        <ArrowUp size={12} />
                                        Update to {versionInfo.latestVersion}
                                        {isOptional && ' (Optional)'}
                                    </>
                                )}
                            </button>
                        )}
                    </div>
                )}
            </div>
        );
    }
};
