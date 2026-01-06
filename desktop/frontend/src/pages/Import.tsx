import React, { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { Folder, Loader2, Check } from "lucide-react";
import { Card } from "../components/ui";
import { itemVariants } from "../utils/constants";
import {
  ImportObsidian,
  SelectDirectory,
  GetDefaultSitesDir,
} from "../../wailsjs/go/main/App";

interface ImportProps {
  onSuccess?: () => void;
  onStatusChange?: (status: {
    type: "success" | "error" | "info";
    message: string;
  }) => void;
}

export const Import: React.FC<ImportProps> = ({
  onSuccess,
  onStatusChange,
}) => {
  const [parentDir, setParentDir] = useState("");
  const [vaultPath, setVaultPath] = useState("");
  const [siteName, setSiteName] = useState("");
  const [linkStyle, setLinkStyle] = useState<"markdown" | "relref">("markdown");
  const [includeDrafts, setIncludeDrafts] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);

  // Load default parent directory on mount
  useEffect(() => {
    const loadDefaultDir = async () => {
      try {
        const dir = await GetDefaultSitesDir();
        if (dir && !parentDir) {
          setParentDir(dir);
        }
      } catch (err) {
        console.error("Failed to load default sites directory:", err);
      }
    };
    loadDefaultDir();
  }, []);

  const handleSelectParentDir = async () => {
    try {
      const dir = await SelectDirectory("Select Parent Directory for New Site");
      if (dir) {
        setParentDir(dir);
      }
    } catch (err) {
      console.error("Error selecting parent directory:", err);
    }
  };

  const handleSelectVaultPath = async () => {
    try {
      const dir = await SelectDirectory("Select Obsidian Vault");
      if (dir) {
        setVaultPath(dir);
        // Auto-set site name from vault directory name if not already set
        if (!siteName) {
          const vaultName = dir.split('/').pop() || '';
          setSiteName(vaultName);
        }
      }
    } catch (err) {
      console.error("Error selecting vault path:", err);
    }
  };

  const handleImport = async () => {
    if (!vaultPath) {
      if (onStatusChange) {
        onStatusChange({
          type: "error",
          message: "Vault path required: Please select an Obsidian vault",
        });
      }
      return;
    }

    setIsProcessing(true);
    if (onStatusChange) {
      onStatusChange({ type: "info", message: "Creating site and importing..." });
    }

    try {
      const result = await ImportObsidian({
        vaultPath,
        siteName: siteName || "", // Defaults to vault name in backend
        parentDir: parentDir || "", // Defaults to current dir in backend
        outputDir: '',
        dryRun: false,
        convertLinks: true,
        linkStyle: linkStyle,
        includeDrafts: includeDrafts,
      });

      if (result.success) {
        if (onStatusChange) {
          onStatusChange({
            type: "success",
            message: `Import Complete: Created site at ${result.sitePath}. ${result.filesImported || 0} files imported.`,
          });
        }
        // Reset form
        setParentDir("");
        setVaultPath("");
        setSiteName("");
        setLinkStyle("markdown");
        setIncludeDrafts(false);

        if (onSuccess) {
          onSuccess();
        }
      } else {
        if (onStatusChange) {
          onStatusChange({
            type: "error",
            message: `Import Failed: ${result.error || "Unknown error"}`,
          });
        }
      }
    } catch (err: any) {
      if (onStatusChange) {
        onStatusChange({
          type: "error",
          message: `Import Error: ${err?.toString() || "Unknown error"}`,
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
          <div className="w-10 h-10 bg-purple-500/20 rounded-sm flex items-center justify-center">
            <Folder size={20} className="text-purple-400" />
          </div>
          <div>
            <h2 className="text-xl font-display font-medium text-white">
              IMPORT_OBSIDIAN
            </h2>
            <p className="text-zinc-500 text-xs font-mono">CREATE_SITE_FROM_VAULT</p>
          </div>
        </div>

        <div className="space-y-6">
          {/* Parent Directory - First */}
          <div className="group">
            <div className="flex items-center justify-between mb-2">
              <label className="block text-[10px] font-mono text-zinc-400 uppercase ml-1 tracking-widest">
                Parent_Directory
              </label>
              <span className="text-[10px] font-mono text-zinc-500">
                (Defaults to current directory)
              </span>
            </div>
            <div className="flex gap-2">
              <input
                type="text"
                value={parentDir}
                onChange={(e) => setParentDir(e.target.value)}
                className="flex-1 bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                placeholder="/path/to/parent (leave empty for current directory)"
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

          {/* Obsidian Vault - Required */}
          <div className="group">
            <label className="block text-[10px] font-mono text-accent uppercase mb-2 ml-1 tracking-widest">
              Obsidian_Vault *
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={vaultPath}
                onChange={(e) => setVaultPath(e.target.value)}
                className="flex-1 bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
                placeholder="/path/to/obsidian-vault"
                autoComplete="off"
                autoCapitalize="off"
                autoCorrect="off"
                spellCheck="false"
              />
              <button
                onClick={handleSelectVaultPath}
                className="px-4 py-3 bg-white/5 border border-white/10 rounded-sm hover:bg-white/10 transition-all"
              >
                <Folder size={18} className="text-zinc-400" />
              </button>
            </div>
          </div>

          {/* Site Name */}
          <div className="group">
            <label className="block text-[10px] font-mono text-zinc-400 uppercase mb-2 ml-1 tracking-widest">
              Site_Name <span className="text-zinc-600">(defaults to vault name)</span>
            </label>
            <input
              type="text"
              value={siteName}
              onChange={(e) => setSiteName(e.target.value)}
              className="w-full bg-black/40 border border-white/10 rounded-sm px-4 py-3 text-sm text-white placeholder:text-zinc-800 focus:outline-none focus:border-accent/50 transition-all font-mono"
              placeholder="my-site"
              autoComplete="off"
              autoCapitalize="off"
              autoCorrect="off"
            />
          </div>

          {/* Link Style Selection */}
          <div className="group">
            <div className="flex items-center justify-between mb-2">
              <label className="block text-[10px] font-mono text-zinc-400 uppercase ml-1 tracking-widest">
                Wikilink_Conversion_Style
              </label>
              <span className="text-[10px] font-mono text-zinc-500">
                (How to convert [[wikilinks]])
              </span>
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setLinkStyle("markdown")}
                className={`flex-1 px-4 py-3 rounded-sm border transition-all font-mono text-xs ${
                  linkStyle === "markdown"
                    ? "bg-accent/20 border-accent/50 text-accent"
                    : "bg-black/40 border-white/10 text-zinc-400 hover:bg-white/5"
                }`}
              >
                <div className="font-medium">Markdown</div>
                <div className="text-[10px] mt-1 opacity-70">
                  [Page](page.md) - Permissive
                </div>
              </button>
              <button
                onClick={() => setLinkStyle("relref")}
                className={`flex-1 px-4 py-3 rounded-sm border transition-all font-mono text-xs ${
                  linkStyle === "relref"
                    ? "bg-accent/20 border-accent/50 text-accent"
                    : "bg-black/40 border-white/10 text-zinc-400 hover:bg-white/5"
                }`}
              >
                <div className="font-medium">Relref</div>
                <div className="text-[10px] mt-1 opacity-70">
                  Hugo shortcode - Strict
                </div>
              </button>
            </div>
            <p className="text-[10px] font-mono text-zinc-600 mt-2 ml-1">
              {linkStyle === "markdown"
                ? "Plain markdown links - works even if target missing"
                : "Hugo relref - ensures all links are valid (may fail build if target missing)"}
            </p>
          </div>

          <div className="flex items-center gap-3">
            <input
              type="checkbox"
              id="includeDrafts"
              checked={includeDrafts}
              onChange={(e) => setIncludeDrafts(e.target.checked)}
              className="w-4 h-4 bg-black/40 border border-white/10 rounded-sm text-accent focus:ring-accent"
            />
            <label
              htmlFor="includeDrafts"
              className="text-sm text-zinc-400 font-mono"
            >
              Include draft files
            </label>
          </div>

          <div className="p-4 bg-black/40 rounded-sm border border-white/10">
            <h4 className="text-sm font-display text-white mb-3">
              What Gets Imported:
            </h4>
            <div className="space-y-2 font-mono text-xs text-zinc-400">
              <div className="flex items-center gap-2">
                <Check size={14} className="text-accent" /> Convert markdown
                files to Hugo format
              </div>
              <div className="flex items-center gap-2">
                <Check size={14} className="text-accent" /> Preserve wikilinks
                and frontmatter
              </div>
              <div className="flex items-center gap-2">
                <Check size={14} className="text-accent" /> Support drafts and
                published content
              </div>
            </div>
          </div>

          <button
            onClick={handleImport}
            disabled={isProcessing || !vaultPath}
            className="w-full bg-purple-500 text-white hover:bg-purple-400 py-4 rounded-sm font-medium transition-all flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed uppercase tracking-wider text-sm"
          >
            {isProcessing ? (
              <Loader2 className="animate-spin" />
            ) : (
              <Folder size={18} />
            )}
            <span>Create_Site_and_Import</span>
          </button>
        </div>
      </Card>
    </motion.div>
  );
};
