import React, { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { X, Palette, Loader2, Download, ExternalLink, Check, AlertCircle } from "lucide-react";
import { buttonVariants, iconButtonVariants } from "../../utils/constants";
import {
  InstallTheme,
  GetInstalledThemes,
} from "../../../wailsjs/go/main/App";

interface ThemeModalProps {
  isOpen: boolean;
  onClose: () => void;
  sitePath?: string;
  onSuccess?: (themeName: string) => void;
  onStatusChange?: (status: {
    type: "success" | "error" | "info";
    message: string;
  }) => void;
}

// Walgo's own Hugo themes
const WALGO_THEMES = [
  {
    name: "Walgo Biolink",
    url: "https://github.com/ganbitlabs/walgo-biolink",
    description: "Minimal link-in-bio theme for personal landing pages",
  },
  {
    name: "Walgo Whitepaper",
    url: "https://github.com/ganbitlabs/walgo-whitepaper",
    description: "Clean, professional theme for whitepapers and reports",
  },
];

export const ThemeModal: React.FC<ThemeModalProps> = ({
  isOpen,
  onClose,
  sitePath = "",
  onSuccess,
  onStatusChange,
}) => {
  const [githubUrl, setGithubUrl] = useState("");
  const [isInstalling, setIsInstalling] = useState(false);
  const [currentThemes, setCurrentThemes] = useState<string[]>([]);
  const [isLoadingThemes, setIsLoadingThemes] = useState(false);
  const [installSuccess, setInstallSuccess] = useState(false);
  const [installedThemeName, setInstalledThemeName] = useState("");

  // Load current themes when modal opens
  useEffect(() => {
    if (isOpen && sitePath) {
      loadCurrentThemes();
      setInstallSuccess(false);
      setInstalledThemeName("");
    }
  }, [isOpen, sitePath]);

  const loadCurrentThemes = async () => {
    if (!sitePath) return;

    setIsLoadingThemes(true);
    try {
      const result = await GetInstalledThemes(sitePath);
      if (result.success) {
        setCurrentThemes(result.themes || []);
      }
    } catch (err) {
      console.error("Error loading themes:", err);
    } finally {
      setIsLoadingThemes(false);
    }
  };

  const handleInstall = async (url?: string) => {
    const themeUrl = url || githubUrl.trim();

    if (!sitePath) {
      onStatusChange?.({
        type: "error",
        message: "Site path is required",
      });
      return;
    }

    if (!themeUrl) {
      onStatusChange?.({
        type: "error",
        message: "Please enter a GitHub URL",
      });
      return;
    }

    // Validate URL format
    if (!themeUrl.includes("github.com")) {
      onStatusChange?.({
        type: "error",
        message: "Only GitHub URLs are supported",
      });
      return;
    }

    setIsInstalling(true);
    setInstallSuccess(false);
    try {
      const result = await InstallTheme({
        sitePath,
        githubUrl: themeUrl,
      });

      if (result.success) {
        setInstallSuccess(true);
        setInstalledThemeName(result.themeName);
        setCurrentThemes([result.themeName]);
        setGithubUrl("");

        onStatusChange?.({
          type: "success",
          message: `Theme '${result.themeName}' installed successfully!`,
        });

        onSuccess?.(result.themeName);
      } else {
        onStatusChange?.({
          type: "error",
          message: `Installation failed: ${result.error}`,
        });
      }
    } catch (err) {
      onStatusChange?.({
        type: "error",
        message: `Installation error: ${err?.toString()}`,
      });
    } finally {
      setIsInstalling(false);
    }
  };

  const handleClose = () => {
    setGithubUrl("");
    setInstallSuccess(false);
    setInstalledThemeName("");
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/80 backdrop-blur-sm"
        onClick={handleClose}
      />

      {/* Modal */}
      <motion.div
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.95, y: 20 }}
        className="relative w-full max-w-lg mx-4 bg-zinc-900 border border-white/10 rounded-lg shadow-2xl overflow-hidden"
      >
        {/* Header */}
        <div className="px-6 py-4 border-b border-white/10 flex items-center justify-between bg-gradient-to-r from-purple-500/10 via-transparent to-pink-500/10">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-purple-500/20 rounded-lg">
              <Palette size={20} className="text-purple-400" />
            </div>
            <div>
              <h2 className="text-lg font-display text-white">Theme Manager</h2>
              <p className="text-xs font-mono text-zinc-500">Install Hugo themes from GitHub</p>
            </div>
          </div>
          <motion.button
            onClick={handleClose}
            variants={iconButtonVariants}
            whileHover="hover"
            whileTap="tap"
            className="p-2 hover:bg-white/10 rounded-lg transition-colors"
          >
            <X size={18} className="text-zinc-400" />
          </motion.button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-6">
          {/* Current Theme */}
          <div className="space-y-2">
            <label className="text-xs font-mono text-zinc-400 uppercase tracking-wider">
              Current Theme
            </label>
            <div className="p-3 bg-zinc-800/50 border border-white/10 rounded-lg">
              {isLoadingThemes ? (
                <div className="flex items-center gap-2 text-zinc-500">
                  <Loader2 size={14} className="animate-spin" />
                  <span className="text-sm font-mono">Loading...</span>
                </div>
              ) : currentThemes.length > 0 ? (
                <div className="flex items-center gap-2">
                  <Check size={14} className="text-green-400" />
                  <span className="text-sm font-mono text-white">{currentThemes[0]}</span>
                </div>
              ) : (
                <div className="flex items-center gap-2 text-zinc-500">
                  <AlertCircle size={14} />
                  <span className="text-sm font-mono">No theme installed</span>
                </div>
              )}
            </div>
          </div>

          {/* Success Message */}
          {installSuccess && (
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              className="p-4 bg-green-500/10 border border-green-500/30 rounded-lg"
            >
              <div className="flex items-center gap-3">
                <Check size={20} className="text-green-400" />
                <div>
                  <p className="text-sm font-mono text-green-400">Theme installed successfully!</p>
                  <p className="text-xs font-mono text-green-400/70 mt-1">
                    Theme '{installedThemeName}' is now active. Click 'Serve Site' to preview.
                  </p>
                </div>
              </div>
            </motion.div>
          )}

          {/* Install from URL */}
          <div className="space-y-3">
            <label className="text-xs font-mono text-zinc-400 uppercase tracking-wider">
              Install from GitHub URL
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={githubUrl}
                onChange={(e) => setGithubUrl(e.target.value)}
                placeholder="https://github.com/user/hugo-theme-name"
                disabled={isInstalling}
                className="flex-1 px-4 py-3 bg-zinc-800/50 border border-white/10 rounded-lg text-sm font-mono text-white placeholder-zinc-600 focus:outline-none focus:border-purple-500/50 transition-colors disabled:opacity-50"
              />
              <motion.button
                onClick={() => handleInstall()}
                disabled={isInstalling || !githubUrl.trim()}
                variants={buttonVariants}
                whileHover="hover"
                whileTap="tap"
                className="px-4 py-3 bg-purple-500/20 hover:bg-purple-500/30 text-purple-400 border border-purple-500/30 rounded-lg transition-colors disabled:opacity-50 flex items-center gap-2"
              >
                {isInstalling ? (
                  <Loader2 size={16} className="animate-spin" />
                ) : (
                  <Download size={16} />
                )}
              </motion.button>
            </div>
          </div>

          {/* Walgo Themes */}
          <div className="space-y-3">
            <label className="text-xs font-mono text-zinc-400 uppercase tracking-wider">
              Walgo Themes
            </label>
            <div className="space-y-2 max-h-48 overflow-y-auto pr-2">
              {WALGO_THEMES.map((theme) => (
                <div
                  key={theme.name}
                  className="p-3 bg-zinc-800/30 border border-white/5 rounded-lg hover:border-purple-500/30 transition-all group"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-mono text-white">{theme.name}</span>
                        <a
                          href={theme.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-zinc-500 hover:text-purple-400 transition-colors"
                          onClick={(e) => e.stopPropagation()}
                        >
                          <ExternalLink size={12} />
                        </a>
                      </div>
                      <p className="text-xs font-mono text-zinc-500 truncate mt-0.5">
                        {theme.description}
                      </p>
                    </div>
                    <motion.button
                      onClick={() => handleInstall(theme.url)}
                      disabled={isInstalling}
                      variants={buttonVariants}
                      whileHover="hover"
                      whileTap="tap"
                      className="ml-3 px-3 py-1.5 bg-purple-500/20 text-purple-400 text-xs font-mono rounded hover:bg-purple-500/30 transition-colors disabled:opacity-50 flex items-center gap-1.5"
                    >
                      {isInstalling ? (
                        <Loader2 size={12} className="animate-spin" />
                      ) : (
                        <Download size={12} />
                      )}
                      Install
                    </motion.button>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Info */}
          <div className="p-3 bg-zinc-800/30 border border-white/5 rounded-lg">
            <p className="text-xs font-mono text-zinc-500">
              <span className="text-zinc-400">Note:</span> Installing a new theme will remove any existing themes.
              The hugo.toml file will be automatically updated with the new theme name.
            </p>
          </div>
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-white/10 flex justify-end gap-3 bg-black/20">
          <motion.button
            onClick={handleClose}
            variants={buttonVariants}
            whileHover="hover"
            whileTap="tap"
            className="px-4 py-2.5 bg-white/5 hover:bg-white/10 border border-white/10 text-sm font-mono text-zinc-300 rounded-lg transition-all"
          >
            Close
          </motion.button>
        </div>
      </motion.div>
    </div>
  );
};
