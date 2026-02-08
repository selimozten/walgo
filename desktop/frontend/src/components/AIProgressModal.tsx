import React from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  X,
  Minimize2,
  Loader2,
  CheckCircle2,
  AlertCircle,
  Sparkles,
} from "lucide-react";
import { useAIProgress } from "../contexts/AIProgressContext";

export const AIProgressModal: React.FC = () => {
  const { progressState, isModalOpen, closeModal, toggleMinimize, completionResult } =
    useAIProgress();

  if (!isModalOpen) return null;

  const progressPercentage = Math.round(progressState.progress * 100);

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4"
        onClick={closeModal}
      >
        <motion.div
          initial={{ scale: 0.9, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          exit={{ scale: 0.9, opacity: 0 }}
          onClick={(e) => e.stopPropagation()}
          className="bg-gradient-to-br from-zinc-900 to-black border border-accent/20 rounded-lg shadow-2xl max-w-2xl w-full overflow-hidden"
        >
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4 border-b border-white/5 bg-black/40">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 bg-accent/20 rounded-sm flex items-center justify-center">
                {progressState.isActive ? (
                  <Loader2 className="text-accent animate-spin" size={18} />
                ) : completionResult?.success === false ? (
                  <AlertCircle className="text-red-500" size={18} />
                ) : (
                  <CheckCircle2 className="text-green-500" size={18} />
                )}
              </div>
              <div>
                <h3 className="text-sm font-medium text-white">
                  {progressState.isActive
                    ? "Creating AI Site"
                    : completionResult?.success === false
                      ? "Site Creation Failed"
                      : "Site Created Successfully"}
                </h3>
                {progressState.siteName && (
                  <p className="text-xs text-zinc-500 font-mono">
                    {progressState.siteName}
                  </p>
                )}
              </div>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={toggleMinimize}
                className="p-2 hover:bg-white/5 rounded-sm transition-colors"
                title="Minimize"
              >
                <Minimize2 size={16} className="text-zinc-400" />
              </button>
              <button
                onClick={closeModal}
                className="p-2 hover:bg-white/5 rounded-sm transition-colors"
                title={progressState.isActive ? "Minimize" : "Close"}
              >
                <X size={16} className="text-zinc-400" />
              </button>
            </div>
          </div>

          {/* Content */}
          <div className="p-6 space-y-6">
            {/* Phase */}
            {progressState.phase && (
              <div className="flex items-center gap-3">
                <Sparkles className="text-accent" size={20} />
                <span className="text-lg font-mono text-accent">
                  {progressState.phase}
                </span>
              </div>
            )}

            {/* Progress Bar */}
            {progressState.total > 0 && (
              <div className="space-y-3">
                <div className="flex justify-between text-sm font-mono">
                  <span className="text-zinc-400">
                    {progressState.current} / {progressState.total} pages
                  </span>
                  <span className="text-accent font-medium">
                    {progressPercentage}%
                  </span>
                </div>
                <div className="relative w-full h-3 bg-zinc-800 rounded-full overflow-hidden">
                  <motion.div
                    initial={{ width: 0 }}
                    animate={{ width: `${progressPercentage}%` }}
                    transition={{ duration: 0.3 }}
                    className="absolute inset-y-0 left-0 bg-gradient-to-r from-accent to-blue-400 rounded-full"
                  />
                  {/* Shimmer effect */}
                  {progressState.isActive && (
                    <motion.div
                      animate={{
                        x: ["-100%", "200%"],
                      }}
                      transition={{
                        duration: 1.5,
                        repeat: Infinity,
                        ease: "linear",
                      }}
                      className="absolute inset-0 w-1/3 bg-gradient-to-r from-transparent via-white/20 to-transparent"
                    />
                  )}
                </div>
              </div>
            )}

            {/* Current File */}
            {progressState.currentFile && (
              <div className="p-4 bg-black/40 border border-white/5 rounded-sm">
                <div className="text-xs text-zinc-500 mb-1 font-mono">
                  Currently creating:
                </div>
                <div className="text-sm text-white font-mono break-all">
                  {progressState.currentFile}
                </div>
              </div>
            )}

            {/* Message */}
            {progressState.message &&
              progressState.message !== progressState.phase && (
                <div className="text-sm text-zinc-400 font-mono">
                  {progressState.message}
                </div>
              )}

            {/* Completion Message */}
            {!progressState.isActive && completionResult?.success === false && (
              <div className="flex items-center gap-2 p-4 bg-red-500/10 border border-red-500/20 rounded-sm">
                <AlertCircle className="text-red-500" size={20} />
                <span className="text-sm text-red-500 font-mono">
                  {completionResult.error || "Site creation failed."}
                </span>
              </div>
            )}
            {!progressState.isActive && completionResult?.success !== false && (
              <div className="flex items-center gap-2 p-4 bg-green-500/10 border border-green-500/20 rounded-sm">
                <CheckCircle2 className="text-green-500" size={20} />
                <span className="text-sm text-green-500 font-mono">
                  Site created successfully! You can now build and deploy it.
                </span>
              </div>
            )}
          </div>

          {/* Footer */}
          {progressState.isActive && (
            <div className="px-6 py-4 bg-black/40 border-t border-white/5">
              <p className="text-xs text-zinc-500 font-mono text-center">
                You can minimize this window and continue working. The creation
                will continue in the background.
              </p>
            </div>
          )}
        </motion.div>
      </motion.div>
    </AnimatePresence>
  );
};
