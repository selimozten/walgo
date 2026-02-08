import React, { useState, useEffect, useRef } from "react";
import { motion } from "framer-motion";
import {
  Wand2,
  Folder,
  Loader2,
  AlertCircle,
  Check,
  BookOpen,
  FileCode,
} from "lucide-react";
import { Card } from "../components/ui";
import { itemVariants } from "../utils/constants";
import { cn } from "../utils/helpers";
import { useSiteCreation } from "../hooks";
import { useAIConfig } from "../hooks";
import { useAIProgress } from "../contexts/AIProgressContext";
import { AICreateSite as AICreateSiteAPI } from "../../wailsjs/go/main/App";

interface AICreateSiteProps {
  onSuccess?: (sitePath: string) => void;
  onStatusChange?: (status: {
    type: "success" | "error" | "info";
    message: string;
  }) => void;
  onNavigateToAI?: () => void;
}

export const AICreateSite: React.FC<AICreateSiteProps> = ({
  onSuccess,
  onStatusChange,
  onNavigateToAI,
}) => {
  const { configured: aiConfigured } = useAIConfig();
  const { startProgress, completionResult, clearCompletionResult } = useAIProgress();
  const {
    siteName,
    setSiteName,
    parentDir,
    setParentDir,
    siteNameExists,
    siteNameCheckLoading,
    handleSelectParentDir,
  } = useSiteCreation();

  const [siteType, setSiteType] = useState("blog");
  const [siteDescription, setSiteDescription] = useState("");
  const [siteAudience, setSiteAudience] = useState("");
  const [isProcessing, setIsProcessing] = useState(false);

  // Store callbacks in refs so effect can access latest values
  const onSuccessRef = useRef(onSuccess);
  const onStatusChangeRef = useRef(onStatusChange);
  const siteNameRef = useRef(siteName);
  onSuccessRef.current = onSuccess;
  onStatusChangeRef.current = onStatusChange;
  siteNameRef.current = siteName;

  const siteTypes = [
    { id: "blog", label: "Blog", icon: BookOpen },
    { id: "docs", label: "Docs", icon: FileCode },
  ];

  // Handle completion via polling result
  useEffect(() => {
    if (!isProcessing || !completionResult) return;

    setIsProcessing(false);
    clearCompletionResult();

    if (completionResult.success) {
      if (onStatusChangeRef.current) {
        onStatusChangeRef.current({
          type: "success",
          message: `AI Site Created: ${siteNameRef.current}`,
        });
      }

      // Reset form
      setSiteName("");
      setParentDir("");
      setSiteDescription("");
      setSiteAudience("");
      setSiteType("blog");

      if (onSuccessRef.current && completionResult.sitePath) {
        onSuccessRef.current(completionResult.sitePath);
      }
    } else {
      if (onStatusChangeRef.current) {
        onStatusChangeRef.current({
          type: "error",
          message: `AI Create Failed: ${completionResult.error || "Unknown error"}`,
        });
      }
    }
  }, [isProcessing, completionResult, clearCompletionResult, setSiteName, setParentDir]);

  const handleCreate = () => {
    if (!siteName) {
      if (onStatusChange) {
        onStatusChange({ type: "error", message: "Site name required" });
      }
      return;
    }

    if (!aiConfigured) {
      if (onStatusChange) {
        onStatusChange({ type: "error", message: "AI not configured" });
      }
      return;
    }

    setIsProcessing(true);

    // Start progress tracking and open modal
    startProgress(siteName);

    if (onStatusChange) {
      onStatusChange({ type: "info", message: "AI creating site..." });
    }

    // Fire and forget - Go method runs in background goroutine
    // Results come back via polling
    AICreateSiteAPI({
      parentDir: parentDir || "",
      siteName,
      siteType,
      description: siteDescription,
      audience: siteAudience,
    });
  };

  return (
    <motion.div variants={itemVariants} className="max-w-2xl mx-auto pt-10">
      <Card className="border-accent/20 bg-gradient-to-br from-zinc-900/40 to-accent/5">
        <div className="flex items-center gap-4 mb-8 border-b border-white/5 pb-6">
          <div className="w-10 h-10 bg-accent text-black rounded-sm flex items-center justify-center shadow-[0_0_20px_rgba(77,162,255,0.4)]">
            <Wand2 size={20} />
          </div>
          <div>
            <h2 className="text-xl font-display font-medium text-white">
              AI_CREATE_SITE
            </h2>
            <p className="text-zinc-500 text-xs font-mono">FULL_AI_WIZARD</p>
          </div>
          {!aiConfigured && (
            <div className="ml-auto px-3 py-1 bg-yellow-500/10 border border-yellow-500/30 rounded-sm text-xs text-yellow-500 font-mono">
              AI_NOT_CONFIGURED
            </div>
          )}
        </div>

        <div className="space-y-5">
          <div className="group">
            <div className="flex items-center justify-between mb-2">
              <label className="block text-[10px] font-mono text-accent uppercase ml-1 tracking-widest">
                Parent_Directory
              </label>
              <span className="text-[10px] font-mono text-zinc-500">
                (Optional - defaults to ~/walgo-sites)
              </span>
            </div>
            <div className="flex gap-2">
              <input
                type="text"
                value={parentDir}
                onChange={(e) => setParentDir(e.target.value)}
                className="flex-1 bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                placeholder="/path/to/parent (leave empty for ~/walgo-sites)"
                autoComplete="off"
                autoCapitalize="off"
                autoCorrect="off"
              />
              <button
                onClick={handleSelectParentDir}
                className="px-4 py-3 bg-white/5 border border-white/10 rounded-sm hover:bg-white/10 transition-all"
              >
                <Folder size={18} className="text-zinc-400" />
              </button>
            </div>
          </div>

          <div className="group">
            <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">
              Site_Name
            </label>
            <input
              type="text"
              value={siteName}
              onChange={(e) => setSiteName(e.target.value)}
              className={cn(
                "w-full bg-black/40 border rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none transition-all font-mono",
                siteNameExists
                  ? "border-red-500/50 focus:border-red-500"
                  : "border-white/10 focus:border-accent/50"
              )}
              placeholder="my-awesome-blog"
              autoComplete="off"
              autoCapitalize="off"
              autoCorrect="off"
              spellCheck="false"
            />
            {siteNameCheckLoading && siteName.length >= 2 && (
              <p className="text-xs text-zinc-500 font-mono mt-2 ml-1">
                Checking availability...
              </p>
            )}
            {siteNameExists && !siteNameCheckLoading && (
              <p className="text-xs text-red-400 font-mono mt-2 ml-1 flex items-center gap-1">
                <AlertCircle size={12} />
                This project name already exists. Please choose a different
                name.
              </p>
            )}
            {!siteNameExists && !siteNameCheckLoading && siteName.length >= 2 && (
              <p className="text-xs text-green-400 font-mono mt-2 ml-1 flex items-center gap-1">
                <Check size={12} />
                Name available
              </p>
            )}
          </div>

          <div className="group">
            <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">
              Site_Type
            </label>
            <div className="grid grid-cols-2 gap-2">
              {siteTypes.map((type) => (
                <button
                  key={type.id}
                  onClick={() => setSiteType(type.id)}
                  className={cn(
                    "p-3 rounded-sm border transition-all flex flex-col items-center gap-2",
                    siteType === type.id
                      ? "bg-accent/20 border-accent/50 text-accent"
                      : "bg-black/20 border-white/10 text-zinc-500 hover:border-white/20"
                  )}
                >
                  <type.icon size={18} />
                  <span className="text-[10px] font-mono">{type.label}</span>
                </button>
              ))}
            </div>
          </div>

          <div className="group">
            <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">
              Site_Description
            </label>
            <textarea
              value={siteDescription}
              onChange={(e) => setSiteDescription(e.target.value)}
              className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono h-20 resize-none"
              placeholder="A tech blog about web development and AI..."
              autoComplete="off"
              autoCapitalize="off"
              autoCorrect="off"
              spellCheck="false"
            />
          </div>

          <div className="group">
            <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">
              Target_Audience
            </label>
            <input
              type="text"
              value={siteAudience}
              onChange={(e) => setSiteAudience(e.target.value)}
              className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
              placeholder="Developers, tech enthusiasts"
              autoComplete="off"
              autoCapitalize="off"
              autoCorrect="off"
              spellCheck="false"
            />
          </div>

          <button
            onClick={handleCreate}
            disabled={
              isProcessing ||
              !siteName ||
              !aiConfigured ||
              siteNameExists ||
              siteNameCheckLoading
            }
            className="w-full bg-accent text-black hover:bg-accent/80 py-4 rounded-sm font-medium transition-all flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed uppercase tracking-wider text-sm"
          >
            {isProcessing ? (
              <Loader2 className="animate-spin" />
            ) : (
              <Wand2 size={18} />
            )}
            <span>
              {isProcessing ? "Creating_Site..." : "Create_With_AI"}
            </span>
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
