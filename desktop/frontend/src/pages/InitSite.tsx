import React from "react";
import { motion } from "framer-motion";
import { Plus, Folder, Loader2, Zap, AlertCircle, Check } from "lucide-react";
import { Card } from "../components/ui";
import { itemVariants } from "../utils/constants";
import { cn } from "../utils/helpers";
import { useSiteCreation } from "../hooks";

interface InitSiteProps {
  onSuccess?: (sitePath: string) => void;
  onStatusChange?: (status: { type: "success" | "error" | "info"; message: string }) => void;
}

export const InitSite: React.FC<InitSiteProps> = ({ onSuccess, onStatusChange }) => {
  const {
    siteName,
    setSiteName,
    parentDir,
    setParentDir,
    siteNameExists,
    siteNameCheckLoading,
    isProcessing,
    handleInit,
    handleSelectParentDir,
  } = useSiteCreation(onSuccess);

  const onSubmit = async () => {
    if (onStatusChange) {
      onStatusChange({ type: "info", message: "Initializing site..." });
    }

    const result = await handleInit();

    if (result.success && onStatusChange) {
      onStatusChange({
        type: "success",
        message: `Site Initialized: ${siteName}`,
      });
    } else if (!result.success && onStatusChange && result.error) {
      onStatusChange({ type: "error", message: `Init Failed: ${result.error}` });
    }
  };

  return (
    <motion.div variants={itemVariants} className="max-w-2xl mx-auto pt-10">
      <Card className="border-white/10 bg-zinc-900/20">
        <div className="flex items-center gap-4 mb-8 border-b border-white/5 pb-6">
          <div className="w-10 h-10 bg-white/10 rounded-sm flex items-center justify-center">
            <Plus size={20} className="text-zinc-300" />
          </div>
          <div>
            <h2 className="text-xl font-display font-medium text-white">
              INIT WALGO
            </h2>
            <p className="text-zinc-500 text-xs font-mono">INITIALIZE_SITE</p>
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
              type="text"
              value={siteName}
              onChange={(e) => setSiteName(e.target.value)}
              className={cn(
                "w-full bg-black/40 border rounded-sm px-6 py-4 text-lg text-white placeholder:text-zinc-800 focus:outline-none transition-all font-display tracking-wide",
                siteNameExists
                  ? "border-red-500/50 focus:border-red-500"
                  : "border-white/10 focus:border-accent/50"
              )}
              placeholder="my-blog"
              autoFocus
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
            <span>Init_Walgo_Site</span>
          </button>
        </div>
      </Card>
    </motion.div>
  );
};
