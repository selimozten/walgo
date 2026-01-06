import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { X, Download, Check, AlertCircle, Loader2, RefreshCw } from 'lucide-react';
import { cn } from '../../utils';

interface InstallStep {
  tool: string;
  status: 'pending' | 'installing' | 'success' | 'error';
  message?: string;
  error?: string;
}

interface AutoInstallModalProps {
  isOpen: boolean;
  onClose: () => void;
  missingDeps: string[];
  onInstall: (tools: string[]) => Promise<void>;
  installing: boolean;
}

export const AutoInstallModal: React.FC<AutoInstallModalProps> = ({
  isOpen,
  onClose,
  missingDeps,
  onInstall,
  installing
}) => {
  const [steps, setSteps] = useState<InstallStep[]>([]);
  const [isComplete, setIsComplete] = useState(false);
  const [hasErrors, setHasErrors] = useState(false);

  // Initialize steps when modal opens
  useEffect(() => {
    if (isOpen && missingDeps.length > 0) {
      const initialSteps: InstallStep[] = missingDeps.map(dep => ({
        tool: dep,
        status: 'pending'
      }));
      setSteps(initialSteps);
      setIsComplete(false);
      setHasErrors(false);
    }
  }, [isOpen, missingDeps]);

  // Handle installation
  const handleInstall = async () => {
    try {
      // Convert display names to tool names
      const toolMap: Record<string, string> = {
        'Sui CLI': 'sui',
        'Walrus CLI': 'walrus',
        'Site Builder': 'site-builder',
        'Hugo': 'hugo'
      };
      
      const tools = missingDeps.map(dep => toolMap[dep] || dep.toLowerCase());
      
      // Start installation
      await onInstall(tools);
      
      // Mark all as success (actual status will come from parent)
      setSteps(prev => prev.map(step => ({
        ...step,
        status: 'success'
      })));
      setIsComplete(true);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Installation failed';
      
      // Mark as error
      setSteps(prev => prev.map(step => ({
        ...step,
        status: 'error',
        error: errorMessage
      })));
      setHasErrors(true);
      setIsComplete(true);
    }
  };

  const handleClose = () => {
    if (!installing) {
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <AnimatePresence>
      <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.95 }}
          className="bg-zinc-900 border border-accent/30 rounded-sm p-6 max-w-2xl w-full max-h-[80vh] overflow-y-auto"
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-accent/20 rounded-sm flex items-center justify-center">
                {installing ? (
                  <Loader2 className="text-accent animate-spin" size={20} />
                ) : isComplete && !hasErrors ? (
                  <Check className="text-green-400" size={20} />
                ) : hasErrors ? (
                  <AlertCircle className="text-red-400" size={20} />
                ) : (
                  <Download className="text-accent" size={20} />
                )}
              </div>
              <div>
                <h2 className="text-xl font-display font-bold text-white">
                  {installing ? 'Installing Dependencies' : isComplete ? 'Installation Complete' : 'Missing Dependencies'}
                </h2>
                <p className="text-sm text-zinc-400 font-mono">
                  {installing ? 'Please wait...' : isComplete ? 'All done!' : `${missingDeps.length} tool(s) required`}
                </p>
              </div>
            </div>
            {!installing && (
              <button
                onClick={handleClose}
                className="p-2 hover:bg-white/10 rounded-sm transition-colors"
              >
                <X size={18} className="text-zinc-400" />
              </button>
            )}
          </div>

          {/* Installation Steps */}
          <div className="space-y-3 mb-6">
            {steps.map((step, index) => (
              <motion.div
                key={step.tool}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: index * 0.1 }}
                className={cn(
                  "p-4 rounded-sm border transition-all",
                  step.status === 'pending' && "bg-zinc-800/50 border-zinc-700",
                  step.status === 'installing' && "bg-accent/10 border-accent/30",
                  step.status === 'success' && "bg-green-500/10 border-green-500/30",
                  step.status === 'error' && "bg-red-500/10 border-red-500/30"
                )}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    {step.status === 'pending' && (
                      <div className="w-5 h-5 rounded-full border-2 border-zinc-600" />
                    )}
                    {step.status === 'installing' && (
                      <Loader2 className="text-accent animate-spin" size={20} />
                    )}
                    {step.status === 'success' && (
                      <Check className="text-green-400" size={20} />
                    )}
                    {step.status === 'error' && (
                      <AlertCircle className="text-red-400" size={20} />
                    )}
                    <div>
                      <p className="text-sm font-mono text-white">{step.tool}</p>
                      {step.message && (
                        <p className="text-xs text-zinc-400 mt-1">{step.message}</p>
                      )}
                      {step.error && (
                        <p className="text-xs text-red-400 mt-1">{step.error}</p>
                      )}
                    </div>
                  </div>
                  <span className={cn(
                    "text-xs font-mono uppercase tracking-wider",
                    step.status === 'pending' && "text-zinc-500",
                    step.status === 'installing' && "text-accent",
                    step.status === 'success' && "text-green-400",
                    step.status === 'error' && "text-red-400"
                  )}>
                    {step.status}
                  </span>
                </div>
              </motion.div>
            ))}
          </div>

          {/* Info Message */}
          {!installing && !isComplete && (
            <div className="bg-blue-500/10 border border-blue-500/30 rounded-sm p-4 mb-6">
              <div className="flex gap-3">
                <AlertCircle className="text-blue-400 flex-shrink-0" size={20} />
                <div className="text-sm text-blue-200">
                  <p className="font-semibold mb-1">Installation Required</p>
                  <p className="text-blue-300/80">
                    These tools are required to create and launch sites. The installation process will:
                  </p>
                  <ul className="list-disc list-inside mt-2 space-y-1 text-blue-300/80">
                    <li>Automatically install suiup (if not already installed)</li>
                    <li>Download and install missing dependencies</li>
                    <li>Configure tools for optimal performance</li>
                    <li>Verify installation success</li>
                  </ul>
                </div>
              </div>
            </div>
          )}

          {/* Error Message */}
          {hasErrors && (
            <div className="bg-red-500/10 border border-red-500/30 rounded-sm p-4 mb-6">
              <div className="flex gap-3">
                <AlertCircle className="text-red-400 flex-shrink-0" size={20} />
                <div className="text-sm text-red-200">
                  <p className="font-semibold mb-1">Installation Failed</p>
                  <p className="text-red-300/80 mb-2">
                    Some tools failed to install. Please check the error messages above and try again.
                  </p>
                  {steps.some(s => s.error) && (
                    <div className="mt-2 space-y-1">
                      {steps.filter(s => s.error).map(step => (
                        <p key={step.tool} className="text-xs text-red-300/60">
                          • {step.tool}: {step.error}
                        </p>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}

          {/* Actions */}
          <div className="flex gap-3">
            {!isComplete && !installing && (
              <>
                <button
                  onClick={handleClose}
                  className="flex-1 px-4 py-2 bg-white/5 hover:bg-white/10 text-zinc-400 hover:text-white rounded-sm text-sm font-mono transition-all"
                >
                  Cancel
                </button>
                <button
                  onClick={handleInstall}
                  disabled={installing}
                  className="flex-1 px-4 py-2 bg-accent/10 hover:bg-accent/20 text-accent border border-accent/30 rounded-sm text-sm font-mono transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                >
                  <Download size={14} />
                  Install All
                </button>
              </>
            )}
            {isComplete && (
              <button
                onClick={() => {
                  if (hasErrors) {
                    // Reset and try again
                    setIsComplete(false);
                    setHasErrors(false);
                    setSteps(prev => prev.map(step => ({
                      ...step,
                      status: 'pending',
                      error: undefined
                    })));
                  } else {
                    handleClose();
                  }
                }}
                className="flex-1 px-4 py-2 bg-accent/10 hover:bg-accent/20 text-accent border border-accent/30 rounded-sm text-sm font-mono transition-all flex items-center justify-center gap-2"
              >
                {hasErrors ? (
                  <>
                    <RefreshCw size={14} />
                    Try Again
                  </>
                ) : (
                  <>
                    <Check size={14} />
                    Done
                  </>
                )}
              </button>
            )}
            {installing && (
              <div className="flex-1 px-4 py-2 bg-accent/5 border border-accent/20 rounded-sm text-sm font-mono text-accent flex items-center justify-center gap-2">
                <Loader2 size={14} className="animate-spin" />
                Installing...
              </div>
            )}
          </div>
        </motion.div>
      </div>
    </AnimatePresence>
  );
};

