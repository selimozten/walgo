import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { X, Copy, Check, ExternalLink, Terminal, AlertCircle } from 'lucide-react';
import { cn } from '../../utils';

interface InstallInstructionsModalProps {
  isOpen: boolean;
  onClose: () => void;
  missingDeps: string[];
  onRefreshStatus?: () => void;
}

interface ToolInstallInfo {
  name: string;
  description: string;
  releaseUrl: string;
  commands: string[];
  notes?: string[];
}

export const InstallInstructionsModal: React.FC<InstallInstructionsModalProps> = ({
  isOpen,
  onClose,
  missingDeps,
  onRefreshStatus,
}) => {
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null);

  const handleClose = () => {
    // Refresh status when closing to detect newly installed tools
    if (onRefreshStatus) {
      onRefreshStatus();
    }
    onClose();
  };

  const toolsInfo: Record<string, ToolInstallInfo> = {
    'All Tools': {
      name: 'All Tools (Recommended)',
      description: 'Install all required tools with one command',
      releaseUrl: 'https://github.com/selimozten/walgo',
      commands: [
        '# One-command installation (Recommended)',
        'curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash',
      ],
      notes: [
        'This will install: suiup, Sui CLI, Walrus CLI, and Site Builder',
        'Interactive installation - you can choose testnet/mainnet',
        'After installation, restart your terminal',
        'Verify with: sui --version && walrus --version && site-builder --version',
      ],
    },
    'sui': {
      name: 'Sui CLI',
      description: 'Sui blockchain command-line interface',
      releaseUrl: 'https://github.com/MystenLabs/sui/releases',
      commands: [
        '# Step 1: Install suiup (Sui version manager)',
        'curl -sSfL https://raw.githubusercontent.com/Mystenlabs/suiup/main/install.sh | sh',
        '',
        '# Step 2: Add to PATH (restart terminal after this)',
        'export PATH="$HOME/.local/bin:$PATH"',
        '',
        '# Step 3: Install Sui CLI',
        'suiup install sui@testnet',
        'suiup default set sui@testnet',
        '',
        '# Optional: Install mainnet version too',
        'suiup install sui@mainnet',
      ],
      notes: [
        'After installation, restart your terminal',
        'Verify with: sui --version',
        'You can switch networks with: suiup default set sui@mainnet',
      ],
    },
    'walrus': {
      name: 'Walrus CLI',
      description: 'Walrus decentralized storage command-line interface',
      releaseUrl: 'https://github.com/MystenLabs/walrus-docs/releases',
      commands: [
        '# Step 1: Ensure suiup is installed (see Sui CLI instructions)',
        '',
        '# Step 2: Install Walrus CLI',
        'suiup install walrus@testnet',
        'suiup default set walrus@testnet',
        '',
        '# Optional: Install mainnet version too',
        'suiup install walrus@mainnet',
      ],
      notes: [
        'Requires suiup to be installed first',
        'Verify with: walrus --version',
        'You can switch networks with: suiup default set walrus@mainnet',
      ],
    },
    'site-builder': {
      name: 'Site Builder',
      description: 'Walrus site builder for deploying static sites',
      releaseUrl: 'https://github.com/MystenLabs/walrus-sites/releases',
      commands: [
        '# Step 1: Ensure suiup is installed (see Sui CLI instructions)',
        '',
        '# Step 2: Install site-builder',
        'suiup install site-builder@mainnet',
        'suiup default set site-builder@mainnet',
      ],
      notes: [
        'Requires suiup to be installed first',
        'Only mainnet version available (works for all networks)',
        'Verify with: site-builder --version',
      ],
    },
    'hugo': {
      name: 'Hugo Extended',
      description: 'Static site generator (Extended version required)',
      releaseUrl: 'https://github.com/gohugoio/hugo/releases',
      commands: [
        '# macOS (Homebrew)',
        'brew install hugo',
        '',
        '# Linux (Snap)',
        'snap install hugo',
        '',
        '# Linux (Debian/Ubuntu)',
        'sudo apt-get install hugo',
        '',
        '# Windows (Chocolatey)',
        'choco install hugo-extended',
        '',
        '# Or download directly from GitHub',
        'https://github.com/gohugoio/hugo/releases',
      ],
      notes: [
        'IMPORTANT: Install Hugo Extended, not standard Hugo',
        'Extended version is required for SCSS/SASS support',
        'Verify with: hugo version (should show "extended")',
        'On macOS, brew install hugo automatically installs Extended',
        'On Windows, use hugo-extended package, not hugo',
      ],
    },
  };

  const handleCopy = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedCommand(text);
      setTimeout(() => setCopiedCommand(null), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  const handleOpenUrl = (url: string) => {
    window.open(url, '_blank');
  };

  if (!isOpen) return null;

  // Always show "All Tools" first if there are missing dependencies
  const allToolsInfo = missingDeps.length > 0 ? [toolsInfo['All Tools']] : [];
  const individualTools = missingDeps.map(dep => toolsInfo[dep]).filter(Boolean);
  const missingTools = [...allToolsInfo, ...individualTools];

  return (
    <AnimatePresence>
      <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.95 }}
          className="bg-zinc-900 border border-accent/30 rounded-sm p-6 max-w-4xl w-full max-h-[85vh] overflow-y-auto"
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-accent/20 rounded-sm flex items-center justify-center">
                <Terminal className="text-accent" size={20} />
              </div>
              <div>
                <h2 className="text-xl font-display font-bold text-white">
                  Installation Instructions
                </h2>
                <p className="text-sm text-zinc-400 font-mono">
                  Manual installation for {missingDeps.length} tool(s)
                </p>
              </div>
            </div>
            <button
              onClick={handleClose}
              className="p-2 hover:bg-white/10 rounded-sm transition-colors"
            >
              <X size={18} className="text-zinc-400" />
            </button>
          </div>

          {/* Info Banner */}
          <div className="bg-blue-500/10 border border-blue-500/30 rounded-sm p-4 mb-6">
            <div className="flex gap-3">
              <AlertCircle className="text-blue-400 flex-shrink-0 mt-0.5" size={18} />
              <div className="text-sm text-blue-200">
                <p className="font-semibold mb-1">Why Manual Installation?</p>
                <p className="text-blue-300/80">
                  Manual installation is more reliable and gives you full control. 
                  Simply copy the commands below and paste them into your terminal.
                </p>
              </div>
            </div>
          </div>

          {/* Installation Instructions for Each Tool */}
          <div className="space-y-6">
            {missingTools.map((tool, index) => (
              <motion.div
                key={tool.name}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.1 }}
                className="bg-zinc-800/50 border border-zinc-700 rounded-sm p-5"
              >
                {/* Tool Header */}
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h3 className="text-lg font-display font-bold text-white mb-1">
                      {tool.name}
                    </h3>
                    <p className="text-sm text-zinc-400">{tool.description}</p>
                  </div>
                  <button
                    onClick={() => handleOpenUrl(tool.releaseUrl)}
                    className="flex items-center gap-2 px-3 py-1.5 bg-accent/10 hover:bg-accent/20 text-accent border border-accent/30 rounded-sm text-xs font-mono transition-all"
                  >
                    <ExternalLink size={14} />
                    GitHub Releases
                  </button>
                </div>

                {/* Commands */}
                <div className="bg-black/40 rounded-sm p-4 mb-4 relative group">
                  <button
                    onClick={() => {
                      const commandText = tool.commands
                        .filter(cmd => !cmd.startsWith('#') && cmd.trim() !== '')
                        .join('\n');
                      handleCopy(commandText);
                    }}
                    className="absolute top-2 right-2 p-2 bg-zinc-800 hover:bg-zinc-700 border border-zinc-600 rounded-sm transition-all opacity-0 group-hover:opacity-100"
                    title="Copy all commands"
                  >
                    {copiedCommand === tool.commands.join('\n') ? (
                      <Check size={14} className="text-green-400" />
                    ) : (
                      <Copy size={14} className="text-zinc-400" />
                    )}
                  </button>
                  <pre className="text-xs font-mono text-zinc-300 overflow-x-auto">
                    {tool.commands.map((cmd, i) => (
                      <div
                        key={i}
                        className={cn(
                          "py-0.5",
                          cmd.startsWith('#') ? "text-zinc-500" : "text-green-400"
                        )}
                      >
                        {cmd || '\u00A0'}
                      </div>
                    ))}
                  </pre>
                </div>

                {/* Individual Command Copy Buttons */}
                <div className="space-y-2 mb-4">
                  {tool.commands
                    .filter(cmd => !cmd.startsWith('#') && cmd.trim() !== '')
                    .map((cmd, i) => (
                      <button
                        key={i}
                        onClick={() => handleCopy(cmd)}
                        className="w-full flex items-center justify-between px-3 py-2 bg-zinc-900/50 hover:bg-zinc-900 border border-zinc-700 hover:border-accent/30 rounded-sm text-xs font-mono text-left transition-all group"
                      >
                        <span className="text-zinc-300 truncate flex-1 mr-2">{cmd}</span>
                        {copiedCommand === cmd ? (
                          <Check size={14} className="text-green-400 flex-shrink-0" />
                        ) : (
                          <Copy size={14} className="text-zinc-500 group-hover:text-accent flex-shrink-0" />
                        )}
                      </button>
                    ))}
                </div>

                {/* Notes */}
                {tool.notes && tool.notes.length > 0 && (
                  <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-sm p-3">
                    <p className="text-xs font-semibold text-yellow-200 mb-2">Important Notes:</p>
                    <ul className="space-y-1">
                      {tool.notes.map((note, i) => (
                        <li key={i} className="text-xs text-yellow-200/80 flex items-start gap-2">
                          <span className="text-yellow-400 mt-0.5">•</span>
                          <span>{note}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </motion.div>
            ))}
          </div>

          {/* Footer */}
          <div className="mt-6 pt-6 border-t border-white/10">
            <div className="flex items-center justify-between">
              <p className="text-xs text-zinc-500 font-mono">
                After installation, restart Walgo to detect the new tools
              </p>
              <button
                onClick={handleClose}
                className="px-4 py-2 bg-accent/10 hover:bg-accent/20 text-accent border border-accent/30 rounded-sm text-sm font-mono transition-all"
              >
                Close
              </button>
            </div>
          </div>
        </motion.div>
      </div>
    </AnimatePresence>
  );
};

