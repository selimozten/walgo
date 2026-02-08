import React, { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  X,
  Rocket,
  RefreshCw,
  Loader2,
  Check,
  AlertCircle,
  ExternalLink,
  Copy,
  Globe,
  Zap,
  Package,
  Wallet,
  Lock,
} from "lucide-react";
import { cn } from "../../utils";
import { buttonVariants, iconButtonVariants } from "../../utils/constants";
import { useWallet, useVersionCheck } from "../../hooks";
import { LoadingOverlay } from "../ui/LoadingOverlay";

interface DeploymentModalProps {
  isOpen: boolean;
  isUpdate: boolean;
  projectName: string;
  sitePath?: string;
  network?: string;
  currentObjectId?: string;
  deployedWallet?: string; // Wallet address used for deployment
  currentEpochs?: number; // Current epochs from project
  onClose: () => void;
  onDeploy: (params: DeploymentParams) => Promise<DeploymentResult>;
}

export interface DeploymentParams {
  network: string;
  epochs: number;
  walletAddress?: string;
}

export interface DeploymentResult {
  success: boolean;
  objectId?: string;
  error?: string;
  logs?: string[];
}

interface GasEstimate {
  wal: number;
  sui: number;
  walRange: string;
  suiRange: string;
}

export const DeploymentModal: React.FC<DeploymentModalProps> = ({
  isOpen,
  isUpdate,
  projectName,
  sitePath,
  network: initialNetwork = "testnet",
  currentObjectId,
  deployedWallet,
  currentEpochs = 1,
  onClose,
  onDeploy,
}) => {
  // Use wallet hook internally
  const { walletInfo, addressList, switchAddress, switchNetwork } = useWallet();
  const { checkVersions, getSuiToolsWithUpdates } = useVersionCheck();
  const [step, setStep] = useState<"config" | "deploying" | "success">("config");
  const [network, setNetwork] = useState(initialNetwork);
  const [epochs, setEpochs] = useState("1");
  const [selectedAddress, setSelectedAddress] = useState(walletInfo?.address || "");
  const [gasEstimate, setGasEstimate] = useState<GasEstimate | null>(null);
  const [estimating, setEstimating] = useState(false);
  const [logs, setLogs] = useState<string[]>([]);
  const [deploymentResult, setDeploymentResult] = useState<DeploymentResult | null>(null);
  const [copiedField, setCopiedField] = useState<string | null>(null);
  const [hasChanges, setHasChanges] = useState(false);
  const [isSwitching, setIsSwitching] = useState(false);
  const [suiToolsNeedUpdate, setSuiToolsNeedUpdate] = useState<string[]>([]);

  // Determine if network/wallet are locked (site already deployed)
  const isLocked = !!currentObjectId;

  // Track changes
  useEffect(() => {
    const initialEpochsStr = isUpdate && currentEpochs > 0 ? currentEpochs.toString() : "1";
    const changed = network !== initialNetwork || epochs !== initialEpochsStr;
    setHasChanges(changed);
  }, [network, epochs, initialNetwork, currentEpochs, isUpdate]);

  // Reset on open - ONLY when modal opens
  useEffect(() => {
    if (isOpen) {
      setStep("config");
      // Load epochs from project if it's an update, otherwise default to 1
      const initialEpochs = isUpdate && currentEpochs > 0 ? currentEpochs.toString() : "1";
      setEpochs(initialEpochs);
      setLogs([]);
      setDeploymentResult(null);
      setHasChanges(false);
      setSuiToolsNeedUpdate([]);

      // If site is already deployed (locked mode), use deployed network and wallet
      if (isLocked && initialNetwork && deployedWallet) {
        setNetwork(initialNetwork);
        setSelectedAddress(deployedWallet);
        setIsSwitching(true);

        // Auto-switch network and wallet asynchronously
        const switchToDeployed = async () => {
          try {
            // Switch network if different
            if (walletInfo?.network !== initialNetwork) {
              const netResult = await switchNetwork(initialNetwork);
              if (!netResult.success) {
                console.error("Failed to switch network:", netResult.error);
              }
            }

            // Switch wallet if different
            if (walletInfo?.address !== deployedWallet) {
              const addrResult = await switchAddress(deployedWallet);
              if (!addrResult.success) {
                console.error("Failed to switch wallet:", addrResult.error);
              }
            }
          } catch (err) {
            console.error("Error during switching:", err);
          } finally {
            setIsSwitching(false);
          }
        };

        switchToDeployed();
      } else {
        // Normal mode: use current wallet info
        const currentNetwork = walletInfo?.network || initialNetwork;
        setNetwork(currentNetwork);
        setSelectedAddress(walletInfo?.address || "");
      }

      // Check versions (especially for mainnet)
      checkVersions().then(() => {
        const toolsNeedingUpdate = getSuiToolsWithUpdates();
        setSuiToolsNeedUpdate(toolsNeedingUpdate);
      });
    }
  }, [isOpen]); // REMOVED walletInfo, checkVersions, getSuiToolsWithUpdates to prevent infinite loop

  // Sync with wallet changes when modal is open - BUT NOT when locked
  useEffect(() => {
    if (isOpen && walletInfo && !isLocked) {
      // Only sync if NOT locked (not a deployed site)
      // Update selected address if wallet address changed
      if (walletInfo.address && walletInfo.address !== selectedAddress) {
        setSelectedAddress(walletInfo.address);
      }
      // Update network if wallet network changed
      if (walletInfo.network && walletInfo.network !== network) {
        setNetwork(walletInfo.network);
      }
    }
  }, [walletInfo?.address, walletInfo?.network, isOpen, isLocked]);

  // Re-estimate when params change - with debounce
  useEffect(() => {
    if (isOpen && step === "config") {
      const timer = setTimeout(() => {
        estimateGas();
      }, 800); // Increased debounce to 800ms
      return () => clearTimeout(timer);
    }
  }, [network, epochs, isOpen, step]);

  const estimateGas = async () => {
    setEstimating(true);
    try {
      // Skip estimation if no site path
      if (!sitePath) {
        throw new Error("Site path not provided");
      }

      const { EstimateGasFee } = await import('../../../wailsjs/go/main/App');
      
      // Call the real gas estimation API
      const result = await EstimateGasFee({
        sitePath: sitePath,
        network: network,
        epochs: parseInt(epochs) || 1,
      });
      
      if (result.success) {
        setGasEstimate({
          wal: result.wal,
          sui: result.sui,
          walRange: result.walRange,
          suiRange: result.suiRange,
        });
      } else {
        // Fallback to simple estimation if API fails
        console.warn("Gas estimation failed:", result.error);
        const epochCount = parseInt(epochs) || 1;
        const baseWal = 0.0015 * epochCount;
        const baseSui = 0.002;
        
        setGasEstimate({
          wal: baseWal,
          sui: baseSui,
          walRange: `${(baseWal * 0.9).toFixed(4)} - ${(baseWal * 1.1).toFixed(4)}`,
          suiRange: `${(baseSui * 0.9).toFixed(4)} - ${(baseSui * 1.1).toFixed(4)}`,
        });
      }
    } catch (err) {
      console.error("Failed to estimate gas:", err);
      // Fallback to simple estimation
      const epochCount = parseInt(epochs) || 1;
      const baseWal = 0.0015 * epochCount;
      const baseSui = 0.002;
      
      setGasEstimate({
        wal: baseWal,
        sui: baseSui,
        walRange: `${(baseWal * 0.9).toFixed(4)} - ${(baseWal * 1.1).toFixed(4)}`,
        suiRange: `${(baseSui * 0.9).toFixed(4)} - ${(baseSui * 1.1).toFixed(4)}`,
      });
    } finally {
      setEstimating(false);
    }
  };

  const handleAddressSwitch = async (address: string) => {
    if (address !== selectedAddress) {
      setIsSwitching(true);
      try {
        const result = await switchAddress(address);
        if (result.success) {
          // Local state will be updated by useEffect watching walletInfo.address
          // This ensures balance is also updated
        }
      } catch (err) {
        console.error("Failed to switch address:", err);
      } finally {
        setIsSwitching(false);
      }
    }
  };

  const parseAndFormatLogs = (rawLogs: string[]): string[] => {
    const formattedLogs: string[] = [];
    
    for (const log of rawLogs) {
      // Skip empty lines at the start
      if (!log.trim() && formattedLogs.length === 0) continue;
      
      // Format different types of log lines
      if (log.includes('📊 Deployment Plan:')) {
        formattedLogs.push('', '═══════════════════════════════════════', log, '═══════════════════════════════════════');
      } else if (log.includes('✨ Incremental deployment:')) {
        formattedLogs.push('', log);
      } else if (log.match(/^\s+📝 Added:|^\s+🔄 Modified:|^\s+🗑️\s+Deleted:|^\s+✅ Unchanged:/)) {
        formattedLogs.push(log);
      } else if (log.includes('💾 Space saved:')) {
        formattedLogs.push(log, '');
      } else if (log.match(/^\s+[+→-]/)) {
        // File operation lines
        formattedLogs.push(log);
      } else if (log.includes('ℹ️') || log.includes('✓') || log.includes('⚠️')) {
        formattedLogs.push(log);
      } else if (log.includes('[') && log.includes(']') && log.includes('/')) {
        // Progress indicators like [3/5]
        formattedLogs.push('', log);
      } else if (log.includes('INFO') || log.includes('WARN') || log.includes('ERROR')) {
        // Site-builder logs - simplify them
        const timestamp = log.match(/\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/)?.[0];
        const message = log.split('INFO')[1] || log.split('WARN')[1] || log.split('ERROR')[1];
        if (message) {
          formattedLogs.push(`  ${timestamp ? '⏱ ' : ''}${message.trim()}`);
        }
      } else if (log.includes('Parsing the directory') || log.includes('Applying the Walrus') || log.includes('Running the delete')) {
        formattedLogs.push(`  ⚙️  ${log.trim()}`);
      } else if (log.includes('Site object ID:')) {
        formattedLogs.push('', '═══════════════════════════════════════', log, '═══════════════════════════════════════');
      } else if (log.includes('To browse your')) {
        formattedLogs.push('', '🌐 Access Options:', log);
      } else if (log.match(/^\s+\d+\./)) {
        // Numbered list items
        formattedLogs.push(log);
      } else if (log.includes('✅') && (log.includes('successfully') || log.includes('completed'))) {
        formattedLogs.push('', log, '');
      } else if (log.includes('💾 Saving project') || log.includes('Project updated')) {
        formattedLogs.push(log);
      } else if (log.trim()) {
        // Other non-empty lines
        formattedLogs.push(log);
      }
    }
    
    return formattedLogs;
  };

  const handleDeploy = async () => {
    setStep("deploying");
    setLogs([]);
    
    try {
      // Add initial logs
      const initialLogs = [
        `${isUpdate ? "🔄 Updating" : "🚀 Deploying"} ${projectName}...`,
        `📡 Network: ${network}`,
        `⏱  Storage: ${epochs} epochs`,
        `💼 Wallet: ${selectedAddress.substring(0, 10)}...${selectedAddress.substring(selectedAddress.length - 6)}`,
        "",
        "⚙️  Initializing deployment...",
      ];
      setLogs(initialLogs);

      // Call the deployment function
      const result = await onDeploy({
        network,
        epochs: parseInt(epochs),
        walletAddress: selectedAddress,
      });

      // Parse and format logs from the result
      if (result.logs && result.logs.length > 0) {
        const formattedLogs = parseAndFormatLogs(result.logs);
        setLogs(prev => [...prev, "", ...formattedLogs]);
      }

      setDeploymentResult(result);
      
      if (result.success) {
        setLogs(prev => [...prev, "", "✅ Deployment completed successfully!", ""]);
        setStep("success");
      } else {
        setLogs(prev => [...prev, "", `❌ Deployment failed: ${result.error || "Unknown error"}`, ""]);
        setTimeout(() => setStep("config"), 3000);
      }
    } catch (err) {
      const error = err instanceof Error ? err.message : "Unknown error";
      setLogs(prev => [...prev, "", `❌ Error: ${error}`, ""]);
      setDeploymentResult({ success: false, error });
      setTimeout(() => setStep("config"), 3000);
    }
  };

  const handleCopy = (text: string, field: string) => {
    navigator.clipboard.writeText(text);
    setCopiedField(field);
    setTimeout(() => setCopiedField(null), 2000);
  };

  const getSuiscanURL = (objectId: string) => {
    const base = network === "mainnet" 
      ? "https://suiscan.xyz/mainnet/object/"
      : "https://suiscan.xyz/testnet/object/";
    return `${base}${objectId}`;
  };

  const getSuivisionURL = (objectId: string) => {
    const base = network === "mainnet"
      ? "https://suivision.xyz/object/"
      : "https://testnet.suivision.xyz/object/";
    return `${base}${objectId}`;
  };

  const openURL = async (url: string, e?: React.MouseEvent) => {
    // Prevent event propagation to avoid closing modal
    if (e) {
      e.preventDefault();
      e.stopPropagation();
    }
    const { BrowserOpenURL } = await import('../../../wailsjs/runtime/runtime');
    BrowserOpenURL(url);
  };

  if (!isOpen) return null;

  return (
    <div 
      className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4"
      onClick={(e) => {
        // Only close if clicking the backdrop, not during deployment
        if (e.target === e.currentTarget && step !== "deploying") {
          onClose();
        }
      }}
    >
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.95 }}
        className="bg-zinc-900 border border-white/10 rounded-sm max-w-2xl w-full max-h-[90vh] overflow-hidden flex flex-col"
      >
        {/* Header */}
        <div className="p-6 border-b border-white/5 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className={cn(
              "p-2 rounded-sm",
              isUpdate ? "bg-orange-500/10" : "bg-accent/10"
            )}>
              {isUpdate ? (
                <RefreshCw size={20} className="text-orange-400" />
              ) : (
                <Rocket size={20} className="text-accent" />
              )}
            </div>
            <div>
              <h2 className="text-lg font-display text-white">
                {isUpdate ? "Update Deployment" : "Deploy to Walrus"}
              </h2>
              <p className="text-xs text-zinc-500 font-mono">{projectName}</p>
            </div>
          </div>
          <motion.button
            onClick={onClose}
            variants={iconButtonVariants}
            whileHover="hover"
            whileTap="tap"
            className="p-2 hover:bg-white/10 rounded-sm transition-colors"
          >
            <X size={18} className="text-zinc-400" />
          </motion.button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          <AnimatePresence mode="wait">
            {step === "config" && (
              <motion.div
                key="config"
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 20 }}
                className="space-y-6"
              >
                {/* Locked Mode Banner */}
                {isLocked && (
                  <div className="bg-orange-500/10 border border-orange-500/30 rounded-sm p-4">
                    <div className="flex items-start gap-3">
                      <Lock size={16} className="text-orange-400 flex-shrink-0 mt-0.5" />
                      <div className="flex-1">
                        <h4 className="text-sm font-mono text-orange-400 mb-1 font-semibold">
                          Deployment Configuration Locked
                        </h4>
                        <p className="text-xs text-orange-400/70 font-mono">
                          Network and wallet are locked to match the existing deployment.
                        </p>
                        {deployedWallet && (
                          <p className="text-xs text-orange-400/70 font-mono mt-2">
                            <span className="font-semibold">Deployed Wallet:</span> {deployedWallet.substring(0, 20)}...{deployedWallet.substring(deployedWallet.length - 6)}
                          </p>
                        )}
                      </div>
                    </div>
                  </div>
                )}

                {/* Network Selection */}
                <div>
                  <label className="block text-sm font-mono text-zinc-400 mb-2 flex items-center gap-2">
                    Network
                    {isLocked && <Lock size={12} className="text-orange-400" />}
                    {isLocked && <span className="text-[10px] text-orange-400">(Locked)</span>}
                  </label>
                  <div className="grid grid-cols-2 gap-3">
                    {["testnet", "mainnet"].map((net) => (
                      <motion.button
                        key={net}
                        onClick={async () => {
                          if (isSwitching || isLocked) return; // Prevent changes when locked

                          setIsSwitching(true);
                          try {
                            // Switch network in SUI CLI
                            const result = await switchNetwork(net);
                            if (result.success) {
                              // Local state will be updated by useEffect watching walletInfo
                              setHasChanges(true);
                            } else {
                              console.error("Failed to switch network:", result.error);
                            }
                          } finally {
                            setIsSwitching(false);
                          }
                        }}
                        disabled={isSwitching || isLocked}
                        variants={buttonVariants}
                        whileHover={isLocked ? undefined : "hover"}
                        whileTap={isLocked ? undefined : "tap"}
                        className={cn(
                          "px-4 py-3 rounded-sm text-sm font-mono transition-all border relative",
                          network === net
                            ? net === "mainnet"
                              ? "bg-blue-500/10 border-blue-500/30 text-blue-400"
                              : "bg-purple-500/10 border-purple-500/30 text-purple-400"
                            : "bg-white/5 border-white/10 text-zinc-400 hover:bg-white/10",
                          (isSwitching || isLocked) && "opacity-50 cursor-not-allowed"
                        )}
                        title={isLocked ? "Network is locked because site is already deployed" : ""}
                      >
                        {isSwitching ? (
                          <span className="flex items-center gap-2">
                            <Loader2 size={14} className="animate-spin" />
                            Switching...
                          </span>
                        ) : (
                          net.charAt(0).toUpperCase() + net.slice(1)
                        )}
                      </motion.button>
                    ))}
                  </div>
                  {isLocked && (
                    <p className="text-xs text-orange-400/70 font-mono mt-2">
                      Network is locked to match the deployed site
                    </p>
                  )}
                </div>

                {/* Epochs */}
                <div>
                  <label className="block text-sm font-mono text-zinc-400 mb-2">
                    Storage Duration (Epochs)
                  </label>
                  <input
                    type="number"
                    value={epochs}
                    onChange={(e) => setEpochs(e.target.value)}
                    min="1"
                    max="53"
                    className={cn(
                      "w-full px-4 py-2 bg-zinc-900 border rounded-sm text-sm text-white placeholder-zinc-500 focus:outline-none transition-colors",
                      hasChanges && epochs !== "5"
                        ? "border-green-500 focus:border-green-400"
                        : "border-zinc-700 focus:border-accent"
                    )}
                  />
                  <p className="text-xs text-zinc-500 font-mono mt-1">
                    {network === "mainnet" 
                      ? `≈ ${parseInt(epochs) * 2} weeks`
                      : `≈ ${epochs} days`}
                  </p>
                </div>

                {/* Wallet Selection - Show when multiple addresses OR when locked */}
                {(addressList.length > 1 || isLocked) && (
                  <div>
                    <label className="block text-sm font-mono text-zinc-400 mb-2 flex items-center gap-2">
                      Wallet Address
                      {isLocked && <Lock size={12} className="text-orange-400" />}
                      {isLocked && <span className="text-[10px] text-orange-400">(Locked)</span>}
                      {isSwitching && !isLocked && (
                        <Loader2 size={12} className="animate-spin text-accent" />
                      )}
                    </label>
                    {addressList.length > 1 ? (
                      <select
                        value={selectedAddress}
                        onChange={(e) => handleAddressSwitch(e.target.value)}
                        disabled={isSwitching || isLocked}
                        className={cn(
                          "w-full px-4 py-2 bg-zinc-900 border border-zinc-700 rounded-sm text-sm text-white focus:outline-none focus:border-accent transition-colors",
                          (isSwitching || isLocked) && "opacity-50 cursor-not-allowed"
                        )}
                        title={isLocked ? "Wallet is locked because site is already deployed" : ""}
                      >
                        {addressList.map((addr) => (
                          <option key={addr} value={addr}>
                            {addr.substring(0, 20)}...
                          </option>
                        ))}
                      </select>
                    ) : (
                      <div
                        className={cn(
                          "w-full px-4 py-2 bg-zinc-900 border border-zinc-700 rounded-sm text-sm text-white font-mono",
                          isLocked && "opacity-50"
                        )}
                        title={isLocked ? "Wallet is locked because site is already deployed" : ""}
                      >
                        {selectedAddress ? `${selectedAddress.substring(0, 20)}...${selectedAddress.substring(selectedAddress.length - 6)}` : 'No address'}
                      </div>
                    )}
                    {isLocked && (
                      <p className="text-xs text-orange-400/70 font-mono mt-2">
                        Wallet is locked to match the deployed site
                      </p>
                    )}
                  </div>
                )}

                {/* Wallet Info */}
                {walletInfo && (
                  <div className="bg-zinc-800/50 rounded-sm p-4 border border-white/5">
                    <div className="flex items-center gap-2 mb-3">
                      <Wallet size={14} className="text-accent" />
                      <span className="text-xs font-mono text-zinc-400 uppercase">
                        Wallet Balance
                      </span>
                    </div>
                    <div className="space-y-2">
                      <div className="flex justify-between text-sm font-mono">
                        <span className="text-zinc-500">SUI:</span>
                        <span className="text-white">{walletInfo.suiBalance.toFixed(4)}</span>
                      </div>
                      <div className="flex justify-between text-sm font-mono">
                        <span className="text-zinc-500">WAL:</span>
                        <span className="text-white">{walletInfo.walBalance.toFixed(4)}</span>
                      </div>
                    </div>
                  </div>
                )}

                {/* Gas Estimate */}
                <div className="bg-zinc-800/50 rounded-sm p-4 border border-white/5">
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <Zap size={14} className="text-yellow-400" />
                      <span className="text-xs font-mono text-zinc-400 uppercase">
                        Estimated Cost
                      </span>
                    </div>
                    {estimating && <Loader2 size={14} className="text-zinc-500 animate-spin" />}
                  </div>
                  {gasEstimate && (
                    <div className="space-y-2">
                      <div className="flex justify-between text-sm font-mono">
                        <span className="text-zinc-500">WAL (storage):</span>
                        <span className="text-white">
                          {gasEstimate.wal.toFixed(4)} WAL
                        </span>
                      </div>
                      <div className="flex justify-between text-xs font-mono text-zinc-600">
                        <span>Range:</span>
                        <span>{gasEstimate.walRange} WAL</span>
                      </div>
                      <div className="flex justify-between text-sm font-mono mt-2">
                        <span className="text-zinc-500">SUI (gas):</span>
                        <span className="text-white">
                          {gasEstimate.sui.toFixed(4)} SUI
                        </span>
                      </div>
                      <div className="flex justify-between text-xs font-mono text-zinc-600">
                        <span>Range:</span>
                        <span>{gasEstimate.suiRange} SUI</span>
                      </div>
                    </div>
                  )}
                </div>

                {/* Version Update Warning for Mainnet */}
                {network === "mainnet" && suiToolsNeedUpdate.length > 0 && (
                  <div className="bg-red-500/10 border border-red-500/30 rounded-sm p-3">
                    <div className="flex items-start gap-2">
                      <AlertCircle size={16} className="text-red-400 flex-shrink-0 mt-0.5" />
                      <div className="text-xs text-red-400 font-mono">
                        <p className="font-semibold mb-1">⚠️ Tools Need Updates for Mainnet</p>
                        <p className="text-red-400/70 mb-2">
                          The following tools need to be updated: {suiToolsNeedUpdate.join(', ')}
                        </p>
                        <p className="text-red-400/70">
                          Please update them in System Health before deploying to mainnet.
                        </p>
                      </div>
                    </div>
                  </div>
                )}

                {isUpdate && currentObjectId && (
                  <div className="bg-blue-500/10 border border-blue-500/30 rounded-sm p-3">
                    <div className="flex items-start gap-2">
                      <AlertCircle size={16} className="text-blue-400 flex-shrink-0 mt-0.5" />
                      <div className="text-xs text-blue-400 font-mono">
                        <p className="font-semibold mb-1">Updating existing deployment</p>
                        <p className="text-blue-400/70">Object ID: {currentObjectId.substring(0, 20)}...</p>
                      </div>
                    </div>
                  </div>
                )}
              </motion.div>
            )}

            {step === "deploying" && (
              <motion.div
                key="deploying"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="space-y-4"
              >
                <div className="flex items-center justify-center gap-3 mb-6">
                  <Loader2 size={24} className="text-accent animate-spin" />
                  <span className="text-lg font-mono text-white">
                    {isUpdate ? "Updating..." : "Deploying..."}
                  </span>
                </div>

                <div className="bg-black/50 rounded-sm p-4 border border-white/5 max-h-96 overflow-y-auto font-mono text-xs scrollbar-thin scrollbar-thumb-zinc-700 scrollbar-track-transparent">
                  {logs.map((log, idx) => {
                    // Determine line styling based on content
                    let lineClass = "text-zinc-400 py-0.5";
                    
                    if (log.includes('✅') || log.includes('✓')) {
                      lineClass = "text-green-400 py-0.5 font-semibold";
                    } else if (log.includes('❌') || log.includes('✗') || log.includes('⚠️')) {
                      lineClass = "text-red-400 py-0.5 font-semibold";
                    } else if (log.includes('═══')) {
                      lineClass = "text-zinc-600 py-0.5";
                    } else if (log.includes('📊') || log.includes('🚀') || log.includes('🔄')) {
                      lineClass = "text-accent py-0.5 font-semibold";
                    } else if (log.includes('ℹ️') || log.includes('📡') || log.includes('⏱') || log.includes('💼')) {
                      lineClass = "text-blue-400 py-0.5";
                    } else if (log.includes('📝') || log.includes('🔄') || log.includes('🗑️') || log.includes('✨')) {
                      lineClass = "text-purple-400 py-0.5";
                    } else if (log.includes('💾')) {
                      lineClass = "text-yellow-400 py-0.5";
                    } else if (log.includes('🌐')) {
                      lineClass = "text-cyan-400 py-0.5 font-semibold";
                    } else if (log.match(/^\s+[+→-]/)) {
                      lineClass = "text-zinc-500 py-0.5 pl-4";
                    } else if (log.match(/^\s+\d+\./)) {
                      lineClass = "text-zinc-400 py-0.5 pl-4";
                    }
                    
                    return (
                      <div key={idx} className={lineClass}>
                        {log || '\u00A0'}
                      </div>
                    );
                  })}
                  {/* Auto-scroll indicator */}
                  <div ref={(el) => el?.scrollIntoView({ behavior: 'smooth' })} />
                </div>
              </motion.div>
            )}

            {step === "success" && deploymentResult?.success && (
              <motion.div
                key="success"
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.95 }}
                className="space-y-6"
              >
                <div className="text-center">
                  <div className="inline-flex items-center justify-center w-16 h-16 bg-green-500/10 rounded-full mb-4">
                    <Check size={32} className="text-green-400" />
                  </div>
                  <h3 className="text-xl font-display text-white mb-2">
                    {isUpdate ? "Update Successful!" : "Deployment Successful!"}
                  </h3>
                  <p className="text-sm text-zinc-500 font-mono">
                    Your site is now live on Walrus
                  </p>
                </div>

                {/* Object ID */}
                <div className="bg-zinc-800/50 rounded-sm p-4 border border-white/5">
                  <div className="flex items-center gap-2 mb-2">
                    <Package size={14} className="text-accent" />
                    <span className="text-xs font-mono text-zinc-400 uppercase">
                      Object ID
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 text-sm font-mono text-white bg-black/50 px-3 py-2 rounded-sm overflow-x-auto">
                      {deploymentResult.objectId}
                    </code>
                    <motion.button
                      onClick={() => handleCopy(deploymentResult.objectId!, "objectId")}
                      variants={iconButtonVariants}
                      whileHover="hover"
                      whileTap="tap"
                      className="p-2 hover:bg-white/10 rounded-sm transition-colors"
                    >
                      {copiedField === "objectId" ? (
                        <Check size={16} className="text-green-400" />
                      ) : (
                        <Copy size={16} className="text-zinc-400" />
                      )}
                    </motion.button>
                  </div>
                </div>

                {/* Explorer Links */}
                <div className="space-y-2">
                  <div className="flex items-center gap-2 mb-2">
                    <ExternalLink size={14} className="text-accent" />
                    <span className="text-xs font-mono text-zinc-400 uppercase">
                      View on Explorers
                    </span>
                  </div>
                  <motion.button
                    onClick={(e) => openURL(getSuiscanURL(deploymentResult.objectId!), e)}
                    variants={buttonVariants}
                    whileHover="hover"
                    whileTap="tap"
                    className="w-full px-4 py-2 bg-white/5 hover:bg-white/10 border border-white/10 rounded-sm text-sm font-mono text-zinc-300 transition-all flex items-center justify-between"
                  >
                    <span>Suiscan</span>
                    <ExternalLink size={14} />
                  </motion.button>
                  <motion.button
                    onClick={(e) => openURL(getSuivisionURL(deploymentResult.objectId!), e)}
                    variants={buttonVariants}
                    whileHover="hover"
                    whileTap="tap"
                    className="w-full px-4 py-2 bg-white/5 hover:bg-white/10 border border-white/10 rounded-sm text-sm font-mono text-zinc-300 transition-all flex items-center justify-between"
                  >
                    <span>Suivision</span>
                    <ExternalLink size={14} />
                  </motion.button>
                </div>

                {/* SuiNS Info */}
                {network === "mainnet" && (
                  <div className="bg-purple-500/10 border border-purple-500/30 rounded-sm p-4">
                    <div className="flex items-start gap-3">
                      <Globe size={16} className="text-purple-400 flex-shrink-0 mt-0.5" />
                      <div className="flex-1">
                        <h4 className="text-sm font-mono text-purple-400 mb-2 font-semibold">
                          Configure SuiNS Domain (Optional)
                        </h4>
                        <p className="text-xs text-purple-400/70 font-mono mb-3">
                          Link a SuiNS domain to access your site at https://your-domain.wal.app
                        </p>
                        
                        <div className="space-y-2 mb-3 text-xs text-purple-400/70 font-mono">
                          <p className="font-semibold text-purple-400">Step 1: Get a SuiNS domain</p>
                          <ul className="list-disc list-inside space-y-1 ml-2">
                            <li>Visit suins.io and connect your wallet</li>
                            <li>Purchase a domain (names use only letters a-z and numbers 0-9)</li>
                          </ul>
                          
                          <p className="font-semibold text-purple-400 mt-3">Step 2: Link to your Walrus Site</p>
                          <ul className="list-disc list-inside space-y-1 ml-2">
                            <li>Go to "Names You Own" on suins.io</li>
                            <li>Click the three dots menu on your domain</li>
                            <li>Select "Link To Walrus Site"</li>
                            <li>Paste your Object ID and approve the transaction</li>
                          </ul>
                          
                          <p className="font-semibold text-purple-400 mt-3">Step 3: Access your site</p>
                          <ul className="list-disc list-inside space-y-1 ml-2">
                            <li>Your site will be live at: https://your-domain.wal.app</li>
                            <li>DNS propagation may take a few minutes</li>
                          </ul>
                        </div>
                        
                        <div className="flex gap-2">
                          <motion.button
                            onClick={(e) => openURL("https://suins.io", e)}
                            variants={buttonVariants}
                            whileHover="hover"
                            whileTap="tap"
                            className="flex-1 px-3 py-1.5 bg-purple-500/20 hover:bg-purple-500/30 border border-purple-500/30 rounded-sm text-xs font-mono text-purple-400 transition-all flex items-center justify-center gap-2"
                          >
                            <Globe size={12} />
                            Go to SuiNS
                            <ExternalLink size={12} />
                          </motion.button>
                          <motion.button
                            onClick={(e) => openURL("https://docs.wal.app/docs/walrus-sites/tutorial-suins", e)}
                            variants={buttonVariants}
                            whileHover="hover"
                            whileTap="tap"
                            className="flex-1 px-3 py-1.5 bg-purple-500/20 hover:bg-purple-500/30 border border-purple-500/30 rounded-sm text-xs font-mono text-purple-400 transition-all flex items-center justify-center gap-2"
                          >
                            View Guide
                            <ExternalLink size={12} />
                          </motion.button>
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </motion.div>
            )}
          </AnimatePresence>
        </div>

        {/* Footer */}
        <div className="p-6 border-t border-white/5 flex gap-3">
          {step === "config" && (
            <>
              <motion.button
                onClick={onClose}
                variants={buttonVariants}
                whileHover="hover"
                whileTap="tap"
                className="flex-1 px-4 py-2 bg-white/5 hover:bg-white/10 text-zinc-400 hover:text-white rounded-sm text-sm font-mono transition-all"
              >
                Cancel
              </motion.button>
              <motion.button
                onClick={handleDeploy}
                variants={buttonVariants}
                whileHover="hover"
                whileTap="tap"
                disabled={estimating || (network === "mainnet" && suiToolsNeedUpdate.length > 0)}
                className={cn(
                  "flex-1 px-4 py-2 rounded-sm text-sm font-mono transition-all flex items-center justify-center gap-2 border",
                  isUpdate
                    ? "bg-orange-500/10 hover:bg-orange-500/20 text-orange-400 border-orange-500/30"
                    : "bg-accent/10 hover:bg-accent/20 text-accent border-accent/30",
                  (estimating || (network === "mainnet" && suiToolsNeedUpdate.length > 0)) && "opacity-50 cursor-not-allowed"
                )}
                title={network === "mainnet" && suiToolsNeedUpdate.length > 0 ? "Update tools in System Health before deploying to mainnet" : ""}
              >
                {isUpdate ? <RefreshCw size={14} /> : <Rocket size={14} />}
                {isUpdate ? "Update Site" : "Deploy Site"}
              </motion.button>
            </>
          )}
          {step === "success" && (
            <motion.button
              onClick={onClose}
              variants={buttonVariants}
              whileHover="hover"
              whileTap="tap"
              className="w-full px-4 py-2 bg-accent/10 hover:bg-accent/20 text-accent border border-accent/30 rounded-sm text-sm font-mono transition-all"
            >
              Done
            </motion.button>
          )}
        </div>
      </motion.div>

      {/* Loading Overlay for Network/Wallet Switching */}
      {isSwitching && (
        <LoadingOverlay 
          message="Switching..." 
          fullScreen={false}
        />
      )}
    </div>
  );
};

