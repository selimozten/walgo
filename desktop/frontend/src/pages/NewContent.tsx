import React, { useState } from "react";
import { motion } from "framer-motion";
import { FileText, Folder, Loader2 } from "lucide-react";
import { Card } from "../components/ui";
import { itemVariants } from "../utils/constants";
import { NewContent as NewContentAPI, SelectDirectory } from "../../wailsjs/go/main/App";

interface NewContentProps {
  onSuccess?: () => void;
  onStatusChange?: (status: {
    type: "success" | "error" | "info";
    message: string;
  }) => void;
}

export const NewContent: React.FC<NewContentProps> = ({
  onSuccess,
  onStatusChange,
}) => {
  const [sitePath, setSitePath] = useState("");
  const [contentSlug, setContentSlug] = useState("");
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

  const handleCreateContent = async () => {
    if (!sitePath || !contentSlug) {
      if (onStatusChange) {
        onStatusChange({
          type: "error",
          message: "Site path and content slug required",
        });
      }
      return;
    }

    setIsProcessing(true);
    if (onStatusChange) {
      onStatusChange({ type: "info", message: "Creating content..." });
    }

    try {
      const result = await NewContentAPI({
        sitePath,
        slug: contentSlug,
        contentType: 'post',
        noBuild: false,
        serve: false,
      });

      if (result.success) {
        if (onStatusChange) {
          onStatusChange({
            type: "success",
            message: `Content Created: ${contentSlug}`,
          });
        }
        // Reset form
        setSitePath("");
        setContentSlug("");

        if (onSuccess) {
          onSuccess();
        }
      } else {
        if (onStatusChange) {
          onStatusChange({
            type: "error",
            message: `Create Failed: ${result.error || "Unknown error"}`,
          });
        }
      }
    } catch (err: any) {
      if (onStatusChange) {
        onStatusChange({
          type: "error",
          message: `Create Error: ${err?.toString() || "Unknown error"}`,
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
          <div className="w-10 h-10 bg-white/10 rounded-sm flex items-center justify-center">
            <FileText size={20} className="text-zinc-300" />
          </div>
          <div>
            <h2 className="text-xl font-display font-medium text-white">
              NEW_CONTENT
            </h2>
            <p className="text-zinc-500 text-xs font-mono">
              CREATE_PAGE_OR_POST
            </p>
          </div>
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
              Content_Slug
            </label>
            <input
              type="text"
              value={contentSlug}
              onChange={(e) => setContentSlug(e.target.value)}
              className="w-full bg-black/40 border border-white/10 rounded-sm px-6 py-4 text-lg text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-display tracking-wide"
              placeholder="posts/my-first-post.md"
              autoFocus
              autoComplete="off"
              autoCapitalize="off"
              autoCorrect="off"
            />
            <p className="text-[10px] font-mono text-zinc-600 mt-2 ml-1">
              Examples: posts/hello.md, pages/about.md, blog/intro.md
            </p>
          </div>

          <button
            onClick={handleCreateContent}
            disabled={isProcessing || !sitePath || !contentSlug}
            className="w-full bg-white text-black hover:bg-accent hover:text-black py-4 rounded-sm font-medium transition-all flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed uppercase tracking-wider text-sm"
          >
            {isProcessing ? (
              <Loader2 className="animate-spin" />
            ) : (
              <FileText size={18} />
            )}
            <span>Create_Content</span>
          </button>
        </div>
      </Card>
    </motion.div>
  );
};
