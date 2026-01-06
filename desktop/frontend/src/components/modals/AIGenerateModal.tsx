import React, { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { X, Sparkles, Loader2, FolderOpen, FileText } from "lucide-react";
import { buttonVariants, iconButtonVariants } from "../../utils/constants";
import {
  GenerateContent,
  GetContentStructure,
} from "../../../wailsjs/go/main/App";

interface AIGenerateModalProps {
  isOpen: boolean;
  onClose: () => void;
  sitePath?: string;
  onSuccess?: (filePath: string) => void;
  onStatusChange?: (status: {
    type: "success" | "error" | "info";
    message: string;
  }) => void;
}

export const AIGenerateModal: React.FC<AIGenerateModalProps> = ({
  isOpen,
  onClose,
  sitePath = "",
  onSuccess,
  onStatusChange,
}) => {
  const [instructions, setInstructions] = useState("");
  const [isProcessing, setIsProcessing] = useState(false);
  const [contentStructure, setContentStructure] = useState<any>(null);
  const [isLoadingStructure, setIsLoadingStructure] = useState(false);

  // Load content structure when modal opens
  useEffect(() => {
    if (isOpen && sitePath) {
      loadContentStructure();
    }
  }, [isOpen, sitePath]);

  const loadContentStructure = async () => {
    if (!sitePath) return;
    
    setIsLoadingStructure(true);
    try {
      const structure = await GetContentStructure(sitePath);
      setContentStructure(structure);
    } catch (err) {
      console.error("Error loading content structure:", err);
    } finally {
      setIsLoadingStructure(false);
    }
  };

  const handleGenerate = async () => {
    if (!sitePath || !instructions.trim()) {
      if (onStatusChange) {
        onStatusChange({
          type: "error",
          message: "Instructions required",
        });
      }
      return;
    }

    setIsProcessing(true);
    try {
      const result = await GenerateContent({
        sitePath,
        filePath: "",
        contentType: "",
        topic: "",
        context: "",
        instructions: instructions.trim(),
      });

      if (result.success) {
        if (onStatusChange) {
          onStatusChange({
            type: "success",
            message: `Content Generated: ${result.filePath}`,
          });
        }
        if (onSuccess) {
          onSuccess(result.filePath);
        }
        handleClose();
      } else {
        if (onStatusChange) {
          onStatusChange({
            type: "error",
            message: `Generation Failed: ${result.error}`,
          });
        }
      }
    } catch (err: any) {
      if (onStatusChange) {
        onStatusChange({
          type: "error",
          message: `Generation Error: ${err?.toString()}`,
        });
      }
    } finally {
      setIsProcessing(false);
    }
  };

  const handleClose = () => {
    if (!isProcessing) {
      setInstructions("");
      setContentStructure(null);
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
        className="bg-zinc-900 border border-white/10 rounded-sm w-full max-w-3xl p-6 max-h-[90vh] overflow-y-auto"
      >
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-accent/20 rounded-sm flex items-center justify-center">
              <Sparkles size={20} className="text-accent" />
            </div>
            <div>
              <h2 className="text-lg font-display text-white">AI Generate Content</h2>
              <p className="text-xs text-zinc-500 font-mono">Just describe what you want - AI handles the rest</p>
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

        {/* Content Structure Display */}
        {isLoadingStructure ? (
          <div className="mb-6 p-4 bg-black/20 border border-white/5 rounded-sm">
            <div className="flex items-center gap-2 text-zinc-500 text-sm font-mono">
              <Loader2 size={14} className="animate-spin" />
              Loading content structure...
            </div>
          </div>
        ) : contentStructure && contentStructure.contentTypes && contentStructure.contentTypes.length > 0 ? (
          <div className="mb-6 p-4 bg-black/20 border border-white/5 rounded-sm">
            <div className="flex items-center gap-2 mb-3">
              <FolderOpen size={16} className="text-accent" />
              <h3 className="text-sm font-mono text-zinc-300">Current Content Structure</h3>
            </div>
            <div className="space-y-2">
              {contentStructure.contentTypes.map((ct: any, idx: number) => (
                <div key={idx} className="flex items-start gap-2 text-xs font-mono">
                  <span className="text-accent">📁</span>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="text-white">{ct.name}/</span>
                      <span className="text-zinc-500">({ct.fileCount} files)</span>
                      {ct.name === contentStructure.defaultType && (
                        <span className="px-1.5 py-0.5 bg-accent/20 text-accent text-[10px] rounded">DEFAULT</span>
                      )}
                    </div>
                    {ct.files && ct.files.length > 0 && ct.files.length <= 5 && (
                      <div className="ml-4 mt-1 space-y-0.5">
                        {ct.files.map((file: string, fidx: number) => (
                          <div key={fidx} className="flex items-center gap-2 text-zinc-600">
                            <FileText size={12} />
                            <span>{file}</span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
            <div className="mt-3 pt-3 border-t border-white/5">
              <p className="text-[10px] text-zinc-500 font-mono">
                💡 AI will automatically choose the right folder and filename based on your instructions
              </p>
            </div>
          </div>
        ) : (
          <div className="mb-6 p-4 bg-black/20 border border-white/5 rounded-sm">
            <p className="text-xs text-zinc-500 font-mono">
              No content structure found. AI will create appropriate folders.
            </p>
          </div>
        )}

        {/* Instructions Input */}
        <div className="mb-6">
          <label className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest mb-2 block">
            What do you want to create? *
          </label>
          <textarea
            value={instructions}
            onChange={(e) => setInstructions(e.target.value)}
            placeholder={`Examples:
• Create a blog post about blockchain technology for beginners
• Write a tutorial on deploying Hugo sites to Walrus
• Generate an about page for my portfolio
• Create a post comparing different decentralized storage solutions

The AI will determine the best location and filename automatically.`}
            autoComplete="off"
            autoCapitalize="off"
            autoCorrect="off"
            spellCheck="false"
            rows={8}
            className="w-full bg-black/20 border border-white/10 rounded-sm px-3 py-2.5 text-sm font-mono text-white placeholder-zinc-600 focus:outline-none focus:border-accent/50 transition-all resize-none"
            disabled={isProcessing}
            autoFocus
          />
          <p className="text-[10px] text-zinc-500 font-mono mt-2">
            Be specific about the topic, tone, and any requirements. The AI will handle content type, filename, and location.
          </p>
        </div>

        {/* Actions */}
        <div className="flex gap-3">
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
            onClick={handleGenerate}
            disabled={isProcessing || !instructions.trim()}
            variants={buttonVariants}
            whileHover="hover"
            whileTap="tap"
            className="flex-1 px-4 py-2.5 bg-accent/10 hover:bg-accent/20 text-accent border border-accent/30 rounded-sm text-sm font-mono transition-all disabled:opacity-50 flex items-center justify-center gap-2"
          >
            {isProcessing ? (
              <>
                <Loader2 size={14} className="animate-spin" />
                Generating...
              </>
            ) : (
              <>
                <Sparkles size={14} />
                Generate Content
              </>
            )}
          </motion.button>
        </div>
      </motion.div>
    </div>
  );
};
