import React from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Loader2, CheckCircle2, Maximize2 } from "lucide-react";
import { useAIProgress } from "../contexts/AIProgressContext";

export const AIProgressIndicator: React.FC = () => {
  const { progressState, isMinimized, openModal } = useAIProgress();

  // Show indicator when minimized OR when active but modal is closed
  const shouldShow =
    (isMinimized || progressState.isActive) && progressState.isActive;

  if (!shouldShow) return null;

  const progressPercentage = Math.round(progressState.progress * 100);

  return (
    <AnimatePresence>
      <motion.div
        initial={{ scale: 0, opacity: 0, y: 100 }}
        animate={{ scale: 1, opacity: 1, y: 0 }}
        exit={{ scale: 0, opacity: 0, y: 100 }}
        className="fixed bottom-6 right-6 z-40"
      >
        <motion.button
          onClick={openModal}
          className="group relative bg-gradient-to-br from-zinc-900 to-black border border-accent/30 rounded-lg shadow-2xl overflow-hidden hover:border-accent/50 transition-all"
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
        >
          {/* Shimmer background effect */}
          <motion.div
            animate={{
              x: ["-100%", "200%"],
            }}
            transition={{
              duration: 2,
              repeat: Infinity,
              ease: "linear",
            }}
            className="absolute inset-0 w-1/2 bg-gradient-to-r from-transparent via-accent/10 to-transparent"
          />

          <div className="relative px-4 py-3 flex items-center gap-3">
            {/* Icon */}
            <div className="w-10 h-10 bg-accent/20 rounded-sm flex items-center justify-center">
              <Loader2 className="text-accent animate-spin" size={20} />
            </div>

            {/* Content */}
            <div className="flex flex-col items-start min-w-[200px]">
              <div className="flex items-center gap-2 mb-1">
                <span className="text-xs font-mono text-accent font-medium">
                  Creating AI Site
                </span>
                <Maximize2
                  size={12}
                  className="text-zinc-500 group-hover:text-accent transition-colors"
                />
              </div>

              {/* Mini Progress Bar */}
              <div className="w-full h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                <motion.div
                  initial={{ width: 0 }}
                  animate={{ width: `${progressPercentage}%` }}
                  transition={{ duration: 0.3 }}
                  className="h-full bg-gradient-to-r from-accent to-blue-400 rounded-full"
                />
              </div>

              {/* Stats */}
              <div className="flex items-center gap-2 mt-1.5">
                <span className="text-[10px] text-zinc-500 font-mono">
                  {progressState.current} / {progressState.total} pages
                </span>
                <span className="text-[10px] text-accent font-mono font-medium">
                  {progressPercentage}%
                </span>
              </div>
            </div>
          </div>

          {/* Glow effect on hover */}
          <div className="absolute inset-0 bg-accent/0 group-hover:bg-accent/5 transition-colors pointer-events-none" />
        </motion.button>
      </motion.div>
    </AnimatePresence>
  );
};
