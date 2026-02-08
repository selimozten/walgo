import React, { useState } from "react";
import { motion } from "framer-motion";
import { Sparkles, Folder, Loader2 } from "lucide-react";
import { Card } from "../components/ui";
import { itemVariants } from "../utils/constants";
import { useAIConfig } from "../hooks";
import {
  GenerateContent,
  SelectDirectory,
} from "../../wailsjs/go/main/App";

interface AIGenerateProps {
  onSuccess?: () => void;
  onStatusChange?: (status: {
    type: "success" | "error" | "info";
    message: string;
  }) => void;
  onNavigateToAI?: () => void;
}

export const AIGenerate: React.FC<AIGenerateProps> = ({
  onSuccess,
  onStatusChange,
  onNavigateToAI,
}) => {
  const { configured: aiConfigured } = useAIConfig();
  const [sitePath, setSitePath] = useState("");
  const [contentType, setContentType] = useState("post");
  const [topic, setTopic] = useState("");
  const [context, setContext] = useState("");
  const [isProcessing, setIsProcessing] = useState(false);

  const handleSelectSitePath = async () => {
    try {
      const dir = await SelectDirectory("Select Hugo Site");
      if (dir) {
        setSitePath(dir);
      }
    } catch (err) {
      console.error("Error selecting site path:", err);
    }
  };

  const handleGenerate = async () => {
    if (!sitePath || !topic) {
      if (onStatusChange) {
        onStatusChange({
          type: "error",
          message: "Site path and topic required",
        });
      }
      return;
    }

    if (!aiConfigured) {
      if (onStatusChange) {
        onStatusChange({
          type: "error",
          message: "AI not configured",
        });
      }
      return;
    }

    setIsProcessing(true);
    if (onStatusChange) {
      onStatusChange({ type: "info", message: "Generating content with AI..." });
    }

    try {
      const result = await GenerateContent({
        sitePath,
        filePath: '',
        contentType,
        topic,
        context: context || '',
        instructions: '', // Empty for legacy mode
      });

      if (result.success) {
        if (onStatusChange) {
          onStatusChange({
            type: "success",
            message: `Content Generated: ${result.filePath || topic}`,
          });
        }
        // Reset form
        setSitePath("");
        setTopic("");
        setContext("");

        if (onSuccess) {
          onSuccess();
        }
      } else {
        if (onStatusChange) {
          onStatusChange({
            type: "error",
            message: `Generation Failed: ${result.error || "Unknown error"}`,
          });
        }
      }
    } catch (err) {
      if (onStatusChange) {
        onStatusChange({
          type: "error",
          message: `AI Error: ${err?.toString() || "Unknown error"}`,
        });
      }
    } finally {
      setIsProcessing(false);
    }
  };

  return (
    <motion.div variants={itemVariants} className="max-w-2xl mx-auto pt-10">
      <Card className="border-white/10 bg-zinc-900/20">
        <div className="flex items-center gap-4 mb-8 border-b border-white/5 pb-6">
          <div className="w-10 h-10 bg-accent/20 rounded-sm flex items-center justify-center">
            <Sparkles size={20} className="text-accent" />
          </div>
          <div>
            <h2 className="text-xl font-display font-medium text-white">
              AI_GENERATE
            </h2>
            <p className="text-zinc-500 text-xs font-mono">
              AI_POWERED_CONTENT
            </p>
          </div>
          {!aiConfigured && (
            <div className="ml-auto px-3 py-1 bg-yellow-500/10 border border-yellow-500/30 rounded-sm text-xs text-yellow-500 font-mono">
              AI_NOT_CONFIGURED
            </div>
          )}
        </div>

        <div className="space-y-6">
          <div className="group">
            <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">
              Hugo_Site
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={sitePath}
                onChange={(e) => setSitePath(e.target.value)}
                className="flex-1 bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                placeholder="/path/to/hugo-site"
                autoComplete="off"
                autoCapitalize="off"
                autoCorrect="off"
              />
              <button
                onClick={handleSelectSitePath}
                className="px-4 py-3 bg-white/5 border border-white/10 rounded-sm hover:bg-white/10 transition-all"
              >
                <Folder size={18} className="text-zinc-400" />
              </button>
            </div>
          </div>

          <div className="group">
            <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">
              Content_Type
            </label>
            <select
              value={contentType}
              onChange={(e) => setContentType(e.target.value)}
              className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white focus:outline-none focus:border-accent/50 transition-all font-mono"
            >
              <option value="post">Blog Post</option>
              <option value="page">Page</option>
            </select>
          </div>

          <div className="group">
            <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">
              Topic
            </label>
            <input
              type="text"
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              className="w-full bg-black/40 border border-white/10 rounded-sm px-6 py-4 text-lg text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-display tracking-wide"
              placeholder="Getting started with blockchain"
              autoFocus
              autoComplete="off"
              autoCapitalize="off"
              autoCorrect="off"
            />
          </div>

          <div className="group">
            <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">
              Additional_Context (Optional)
            </label>
            <textarea
              value={context}
              onChange={(e) => setContext(e.target.value)}
              className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono h-24 resize-none"
              placeholder="Target audience, tone, specific points to cover..."
            />
          </div>

          <button
            onClick={handleGenerate}
            disabled={isProcessing || !sitePath || !topic || !aiConfigured}
            className="w-full bg-accent text-black hover:bg-accent/80 py-4 rounded-sm font-medium transition-all flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed uppercase tracking-wider text-sm"
          >
            {isProcessing ? (
              <Loader2 className="animate-spin" />
            ) : (
              <Sparkles size={18} />
            )}
            <span>Generate_Content</span>
          </button>

          {!aiConfigured && (
            <p className="text-center text-xs text-zinc-500 font-mono">
              Configure AI credentials in the{" "}
              <button
                onClick={onNavigateToAI}
                className="text-accent hover:underline"
              >
                AI tab
              </button>{" "}
              first
            </p>
          )}
        </div>
      </Card>
    </motion.div>
  );
};
