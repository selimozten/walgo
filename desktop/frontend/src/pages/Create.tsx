import React, { useState } from "react";
import { motion } from "framer-motion";
import { Wand2, Rocket, Folder, Plus, AlertCircle } from "lucide-react";
import { Card } from "../components/ui";
import { itemVariants } from "../utils/constants";
import { cn } from "../utils/helpers";
import { useAIProgress } from "../contexts/AIProgressContext";
import { SystemHealth } from "../types";

interface CreateProps {
  onNavigate: (tab: string) => void;
  aiConfigured?: boolean;
  systemHealth?: SystemHealth;
  onStatusChange?: (status: { type: 'success' | 'error' | 'info'; message: string }) => void;
  onRefreshHealth?: () => Promise<void>;
}

export const Create: React.FC<CreateProps> = ({ onNavigate, aiConfigured = false, systemHealth, onStatusChange, onRefreshHealth }) => {
  const { progressState } = useAIProgress();
  
  const hugoInstalled = systemHealth?.hugoInstalled ?? false;

  const handleAIPipelineClick = () => {
    if (!hugoInstalled) {
      if (onStatusChange) {
        onStatusChange({
          type: 'error',
          message: 'Hugo is required to create sites. Please install Hugo from System Health page.'
        });
      }
      return;
    }
    if (progressState.isActive) {
      if (onStatusChange) {
        onStatusChange({
          type: 'info',
          message: 'AI in progress: AI is currently creating a site. Please wait until it finishes.'
        });
      }
    } else if (aiConfigured) {
      onNavigate("ai-create-site");
    } else {
      if (onStatusChange) {
        onStatusChange({
          type: 'error',
          message: 'AI not configured: Please configure AI in Settings to use AI features.'
        });
      }
    }
  };

  const handleQuickStartClick = () => {
    if (!hugoInstalled) {
      if (onStatusChange) {
        onStatusChange({
          type: 'error',
          message: 'Hugo is required to create sites. Please install Hugo from System Health page.'
        });
      }
      return;
    }
    onNavigate("quickstart");
  };

  const handleObsidianClick = () => {
    if (!hugoInstalled) {
      if (onStatusChange) {
        onStatusChange({
          type: 'error',
          message: 'Hugo is required to import Obsidian vaults. Please install Hugo from System Health page.'
        });
      }
      return;
    }
    onNavigate("import");
  };

  const handleInitClick = () => {
    if (!hugoInstalled) {
      if (onStatusChange) {
        onStatusChange({
          type: 'error',
          message: 'Hugo is required to initialize sites. Please install Hugo from System Health page.'
        });
      }
      return;
    }
    onNavigate("create-site");
  };

  return (
    <motion.div variants={itemVariants} className="max-w-4xl mx-auto pt-10">
      <div className="grid grid-cols-2 gap-6">
        <Card
          onClick={handleAIPipelineClick}
          className={cn(
            "h-full flex flex-col justify-center items-center gap-4 p-8 relative",
            (!aiConfigured || progressState.isActive || !hugoInstalled)
              ? "bg-zinc-900/20 border-zinc-800 opacity-60 cursor-not-allowed"
              : "bg-gradient-to-br from-accent/20 to-purple-500/10 border-accent/30 cursor-pointer"
          )}
        >
          {!hugoInstalled && (
            <div className="absolute top-2 right-2">
              <AlertCircle size={20} className="text-red-400" />
            </div>
          )}
          <Wand2 size={48} className={cn(
            aiConfigured && !progressState.isActive && hugoInstalled ? "text-accent" : "text-zinc-600"
          )} />
          <div className="text-center">
            <h3 className="text-2xl font-display text-white mb-2">
              AI Pipeline
            </h3>
            <p className="text-sm text-zinc-500 font-mono">
              {!hugoInstalled ? "Hugo Required" : progressState.isActive ? "AI is busy creating a site..." : "Full AI Wizard"}
            </p>
          </div>
        </Card>

        <Card
          onClick={handleQuickStartClick}
          className={cn(
            "bg-zinc-900/20 h-full flex flex-col justify-center items-center gap-4 p-8 relative",
            !hugoInstalled && "opacity-60 cursor-not-allowed"
          )}
        >
          {!hugoInstalled && (
            <div className="absolute top-2 right-2">
              <AlertCircle size={20} className="text-red-400" />
            </div>
          )}
          <Rocket size={48} className={cn(hugoInstalled ? "text-accent" : "text-zinc-600")} />
          <div className="text-center">
            <h3 className="text-2xl font-display text-white mb-2">
              QuickStart
            </h3>
            <p className="text-sm text-zinc-500 font-mono">
              {!hugoInstalled ? "Hugo Required" : "Create & Build"}
            </p>
          </div>
        </Card>

        <Card
          onClick={handleObsidianClick}
          className={cn(
            "bg-zinc-900/20 h-full flex flex-col justify-center items-center gap-4 p-8 relative",
            !hugoInstalled && "opacity-60 cursor-not-allowed"
          )}
        >
          {!hugoInstalled && (
            <div className="absolute top-2 right-2">
              <AlertCircle size={20} className="text-red-400" />
            </div>
          )}
          <Folder size={48} className={cn(hugoInstalled ? "text-purple-400" : "text-zinc-600")} />
          <div className="text-center">
            <h3 className="text-2xl font-display text-white mb-2">
              Obsidian Import
            </h3>
            <p className="text-sm text-zinc-500 font-mono">
              {!hugoInstalled ? "Hugo Required" : "Vault to Hugo"}
            </p>
          </div>
        </Card>

        <Card
          onClick={handleInitClick}
          className={cn(
            "bg-zinc-900/20 h-full flex flex-col justify-center items-center gap-4 p-8 relative",
            !hugoInstalled && "opacity-60 cursor-not-allowed"
          )}
        >
          {!hugoInstalled && (
            <div className="absolute top-2 right-2">
              <AlertCircle size={20} className="text-red-400" />
            </div>
          )}
          <Plus size={48} className={cn(hugoInstalled ? "text-zinc-400" : "text-zinc-600")} />
          <div className="text-center">
            <h3 className="text-2xl font-display text-white mb-2">
              Init Walgo
            </h3>
            <p className="text-sm text-zinc-500 font-mono">
              {!hugoInstalled ? "Hugo Required" : "Initialize Site"}
            </p>
          </div>
        </Card>
      </div>
    </motion.div>
  );
};
