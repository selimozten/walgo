import React, { useState } from "react";
import { motion } from "framer-motion";
import {
  Rocket,
  Folder,
  Loader2,
  Zap,
  AlertCircle,
  Check,
  BookOpen,
  FileCode,
  Link,
  FileText,
} from "lucide-react";
import { Card } from "../components/ui";
import { itemVariants } from "../utils/constants";
import { cn } from "../utils/helpers";
import { useSiteCreation } from "../hooks";

interface QuickStartProps {
  onSuccess?: (sitePath: string) => void;
  onStatusChange?: (status: { type: "success" | "error" | "info"; message: string }) => void;
}

export const QuickStart: React.FC<QuickStartProps> = ({ onSuccess, onStatusChange }) => {
  const {
    siteName,
    setSiteName,
    parentDir,
    setParentDir,
    siteNameExists,
    siteNameCheckLoading,
    isProcessing,
    handleQuickStart,
    handleSelectParentDir,
  } = useSiteCreation(onSuccess);

  const [siteType, setSiteType] = useState("biolink");

  const siteTypes = [
    { id: "biolink", label: "Biolink", icon: Link },
    { id: "blog", label: "Blog", icon: BookOpen },
    { id: "docs", label: "Docs", icon: FileCode },
    { id: "whitepaper", label: "Whitepaper", icon: FileText },
  ];

  const onSubmit = async () => {
    if (onStatusChange) {
      onStatusChange({ type: "info", message: "Initializing quickstart..." });
    }

    const result = await handleQuickStart(siteType);

    if (result.success && onStatusChange) {
      onStatusChange({
        type: "success",
        message: `Quickstart Complete: ${siteName}`,
      });
    } else if (!result.success && onStatusChange && result.error) {
      onStatusChange({ type: "error", message: `Quickstart Failed: ${result.error}` });
    }
  };

  return (
    <motion.div variants={itemVariants} className="max-w-2xl mx-auto pt-10">
      <Card className="border-white/10 bg-zinc-900/20">
        <div className="flex items-center gap-4 mb-8 border-b border-white/5 pb-6">
          <div className="w-10 h-10 bg-accent text-black rounded-sm flex items-center justify-center shadow-[0_0_15px_rgba(77,162,255,0.3)]">
            <Rocket size={20} />
          </div>
          <div>
            <h2 className="text-xl font-display font-medium text-white">
              QUICKSTART
            </h2>
            <p className="text-zinc-500 text-xs font-mono">
              CREATE_BUILD_DEPLOY
            </p>
          </div>
        </div>

        <div className="space-y-6">
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
                key="quickstart-parent-dir"
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
              Project_Name
            </label>
            <input
              key="quickstart-site-name"
              type="text"
              value={siteName}
              onChange={(e) => setSiteName(e.target.value)}
              className={cn(
                "w-full bg-black/40 border rounded-sm px-6 py-4 text-lg text-white placeholder:text-zinc-800 focus:outline-none transition-all font-display tracking-wide",
                siteNameExists
                  ? "border-red-500/50 focus:border-red-500"
                  : "border-white/10 focus:border-accent/50"
              )}
              placeholder="my-site"
              autoFocus
              autoComplete="off"
              autoCapitalize="off"
              autoCorrect="off"
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
            <div className="grid grid-cols-4 gap-2">
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

          <button
            onClick={onSubmit}
            disabled={
              isProcessing ||
              !siteName ||
              siteNameExists ||
              siteNameCheckLoading
            }
            className="w-full bg-white text-black hover:bg-accent hover:text-black py-4 rounded-sm font-medium transition-all flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed uppercase tracking-wider text-sm"
          >
            {isProcessing ? (
              <Loader2 className="animate-spin" />
            ) : (
              <Zap size={18} />
            )}
            <span>Execute_Quickstart</span>
          </button>
        </div>
      </Card>
    </motion.div>
  );
};
