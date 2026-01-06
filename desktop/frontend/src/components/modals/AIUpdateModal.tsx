import React, { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { X, Sparkles, FileText, Loader2, AlertCircle } from "lucide-react";
import { buttonVariants, iconButtonVariants } from "../../utils/constants";
import { UpdateContent } from "../../../wailsjs/go/main/App";

interface AIUpdateModalProps {
  isOpen: boolean;
  onClose: () => void;
  sitePath: string;
  filePath: string;
  currentContent?: string;
  onSuccess?: () => void;
  onStatusChange?: (status: {
    type: "success" | "error" | "info";
    message: string;
  }) => void;
}

export const AIUpdateModal: React.FC<AIUpdateModalProps> = ({
  isOpen,
  onClose,
  sitePath,
  filePath,
  currentContent = "",
  onSuccess,
  onStatusChange,
}) => {
  const [instructions, setInstructions] = useState("");
  const [isProcessing, setIsProcessing] = useState(false);
  const [isMarkdown, setIsMarkdown] = useState(false);

  // Check if file is markdown
  useEffect(() => {
    if (filePath) {
      const ext = filePath.toLowerCase();
      setIsMarkdown(ext.endsWith('.md') || ext.endsWith('.markdown'));
    }
  }, [filePath]);

  const handleUpdate = async () => {
    if (!instructions.trim()) {
      if (onStatusChange) {
        onStatusChange({
          type: "error",
          message: "Instructions required",
        });
      }
      return;
    }

    if (!isMarkdown) {
      if (onStatusChange) {
        onStatusChange({
          type: "error",
          message: "AI update only supports Markdown files",
        });
      }
      return;
    }

    setIsProcessing(true);
    try {
      const result = await UpdateContent({
        filePath,
        instructions,
        sitePath,
      });

      if (result.success) {
        if (onStatusChange) {
          onStatusChange({
            type: "success",
            message: "Content updated successfully",
          });
        }
        if (onSuccess) {
          onSuccess();
        }
        handleClose();
      } else {
        if (onStatusChange) {
          onStatusChange({
            type: "error",
            message: `Update Failed: ${result.error}`,
          });
        }
      }
    } catch (err: any) {
      if (onStatusChange) {
        onStatusChange({
          type: "error",
          message: `Update Error: ${err?.toString()}`,
        });
      }
    } finally {
      setIsProcessing(false);
    }
  };

  const handleClose = () => {
    if (!isProcessing) {
      setInstructions("");
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4"
      onClick={(e) => {
        if (e.target === e.currentTarget && !isProcessing) {
          handleClose();
        }
      }}
    >
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.95 }}
        className="bg-zinc-900 border border-white/10 rounded-sm w-full max-w-2xl p-6 max-h-[90vh] overflow-y-auto"
      >
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-accent/20 rounded-sm flex items-center justify-center">
              <Sparkles size={20} className="text-accent" />
            </div>
            <div>
              <h2 className="text-lg font-display text-white">AI Update Content</h2>
              <p className="text-xs text-zinc-500 font-mono">Improve existing markdown content with AI</p>
            </div>
          </div>
          <motion.button
            onClick={handleClose}
            disabled={isProcessing}
            variants={iconButtonVariants}
            whileHover="hover"
            whileTap="tap"
            className="text-zinc-500 hover:text-white transition-colors disabled:opacity-50"
          >
            <X size={20} />
          </motion.button>
        </div>

        {/* File Info */}
        <div className="mb-4 p-3 bg-black/20 border border-white/10 rounded-sm">
          <div className="flex items-center gap-2 mb-1">
            <FileText size={14} className="text-zinc-400" />
            <span className="text-xs font-mono text-zinc-400">Current File:</span>
          </div>
          <p className="text-sm font-mono text-white break-all">{filePath}</p>
        </div>

        {/* Markdown Warning */}
        {!isMarkdown && (
          <div className="mb-4 p-3 bg-yellow-500/10 border border-yellow-500/30 rounded-sm flex items-start gap-2">
            <AlertCircle size={16} className="text-yellow-400 mt-0.5 flex-shrink-0" />
            <div>
              <p className="text-xs font-mono text-yellow-400 font-bold mb-1">
                ⚠️ NOT A MARKDOWN FILE
              </p>
              <p className="text-xs font-mono text-yellow-400/80">
                AI Update only works with markdown (.md) files. Please select a markdown file to use this feature.
              </p>
            </div>
          </div>
        )}

        {/* Form */}
        <div className="space-y-4">
          {/* Instructions */}
          <div>
            <label className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest mb-2 block">
              Update Instructions *
            </label>
            <textarea
              value={instructions}
              onChange={(e) => setInstructions(e.target.value)}
              placeholder="e.g., Add more details about the technical implementation, improve the introduction, add code examples..."
              autoComplete="off"
              autoCapitalize="off"
              autoCorrect="off"
              spellCheck="false"
              rows={6}
              className="w-full bg-black/20 border border-white/10 rounded-sm px-3 py-2.5 text-sm font-mono text-white placeholder-zinc-600 focus:outline-none focus:border-accent/50 transition-all resize-none"
              disabled={isProcessing || !isMarkdown}
            />
            <p className="text-xs text-zinc-500 font-mono mt-2">
              Describe what changes you want the AI to make to the content
            </p>
          </div>

          {/* Info Box */}
          <div className="p-3 bg-accent/5 border border-accent/20 rounded-sm">
            <p className="text-xs font-mono text-accent/80 leading-relaxed">
              💡 <span className="font-bold">Tip:</span> Be specific about what you want to change. 
              The AI will analyze the current content and apply your instructions while maintaining the document structure.
            </p>
          </div>
        </div>

        {/* Actions */}
        <div className="flex gap-3 mt-6">
          <motion.button
            onClick={handleClose}
            disabled={isProcessing}
            variants={buttonVariants}
            whileHover="hover"
            whileTap="tap"
            className="flex-1 px-4 py-2.5 bg-white/5 hover:bg-white/10 border border-white/10 text-sm font-mono text-zinc-300 rounded-sm transition-all disabled:opacity-50"
          >
            Cancel
          </motion.button>
          <motion.button
            onClick={handleUpdate}
            disabled={isProcessing || !instructions.trim() || !isMarkdown}
            variants={buttonVariants}
            whileHover="hover"
            whileTap="tap"
            className="flex-1 px-4 py-2.5 bg-accent/10 hover:bg-accent/20 text-accent border border-accent/30 rounded-sm text-sm font-mono transition-all disabled:opacity-50 flex items-center justify-center gap-2"
          >
            {isProcessing ? (
              <>
                <Loader2 size={14} className="animate-spin" />
                Updating...
              </>
            ) : (
              <>
                <Sparkles size={14} />
                Update Content
              </>
            )}
          </motion.button>
        </div>
      </motion.div>
    </div>
  );
};

