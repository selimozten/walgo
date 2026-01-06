import React from "react";
import { X, AlertCircle, Loader2 } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";

interface DeleteConfirmModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => Promise<void> | void;
  title?: string;
  message?: string;
  itemName?: string;
  isProcessing?: boolean;
}

export const DeleteConfirmModal: React.FC<DeleteConfirmModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  title = "Confirm Deletion",
  message = "Are you sure you want to delete this item? This action cannot be undone.",
  itemName,
  isProcessing = false,
}) => {
  if (!isOpen) return null;

  const handleConfirm = async () => {
    await onConfirm();
    if (!isProcessing) {
      onClose();
    }
  };

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50"
        onClick={onClose}
      >
        <motion.div
          initial={{ scale: 0.9, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          exit={{ scale: 0.9, opacity: 0 }}
          onClick={(e) => e.stopPropagation()}
          className="bg-zinc-900 border border-red-500/30 rounded-sm p-6 max-w-md w-full mx-4"
        >
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-red-500/20 rounded-sm flex items-center justify-center">
                <AlertCircle size={20} className="text-red-400" />
              </div>
              <h3 className="text-xl font-mono text-white uppercase tracking-wide">
                {title}
              </h3>
            </div>
            <button
              onClick={onClose}
              className="text-zinc-500 hover:text-white transition-colors"
              disabled={isProcessing}
            >
              <X size={20} />
            </button>
          </div>

          <div className="space-y-4">
            <p className="text-sm text-zinc-300 font-mono leading-relaxed">
              {message}
            </p>

            {itemName && (
              <div className="bg-red-500/10 border border-red-500/30 rounded-sm p-3">
                <p className="text-xs text-red-400 font-mono break-all">
                  {itemName}
                </p>
              </div>
            )}

            <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-sm p-3">
              <p className="text-[10px] text-yellow-400 font-mono leading-relaxed">
                <span className="font-bold">⚠️ WARNING:</span> This action is
                permanent and cannot be undone.
              </p>
            </div>

            <div className="flex gap-3 pt-2">
              <button
                onClick={onClose}
                disabled={isProcessing}
                className="flex-1 px-4 py-3 bg-white/5 hover:bg-white/10 border border-white/10 rounded-sm text-sm text-zinc-300 hover:text-white font-mono transition-all uppercase tracking-wide disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                onClick={handleConfirm}
                disabled={isProcessing}
                className="flex-1 px-4 py-3 bg-red-500/20 hover:bg-red-500/30 border border-red-500/40 rounded-sm text-sm text-red-400 font-mono transition-all uppercase tracking-wide disabled:opacity-50 flex items-center justify-center gap-2"
              >
                {isProcessing ? (
                  <>
                    <Loader2 size={14} className="animate-spin" />
                    Deleting...
                  </>
                ) : (
                  "Delete"
                )}
              </button>
            </div>
          </div>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  );
};
