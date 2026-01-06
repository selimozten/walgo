import React, { useState, useEffect } from "react";
import { X, Rocket, Loader2 } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { cn } from "../../utils/helpers";
import { Project } from "../../types";

export interface LaunchConfig {
  projectName: string;
  category: string;
  description: string;
  imageUrl?: string;
}

interface LaunchModalProps {
  isOpen: boolean;
  onClose: () => void;
  onLaunch: (config: LaunchConfig) => Promise<void>;
  project?: Project;
  isProcessing?: boolean;
}

export const LaunchModal: React.FC<LaunchModalProps> = ({
  isOpen,
  onClose,
  onLaunch,
  project,
  isProcessing = false,
}) => {
  // Original values (loaded from project)
  const [originalValues, setOriginalValues] = useState({
    projectName: "",
    category: "website",
    description: "A website deployed with Walgo to Walrus Sites",
    imageUrl: "",
  });

  // Current values
  const [projectName, setProjectName] = useState("");
  const [category, setCategory] = useState("website");
  const [description, setDescription] = useState(
    "A website deployed with Walgo to Walrus Sites"
  );
  const [imageUrl, setImageUrl] = useState("");

  // Track which fields have been modified
  const [modifiedFields, setModifiedFields] = useState({
    projectName: false,
    category: false,
    description: false,
    imageUrl: false,
  });

  // Auto-load project data when modal opens
  useEffect(() => {
    if (isOpen && project) {
      const initial = {
        projectName: project.name || "",
        category: project.category || "website",
        description: project.description || "A website deployed with Walgo to Walrus Sites",
        imageUrl: project.imageUrl || "",
      };

      setOriginalValues(initial);
      setProjectName(initial.projectName);
      setCategory(initial.category);
      setDescription(initial.description);
      setImageUrl(initial.imageUrl);

      // Reset modified fields
      setModifiedFields({
        projectName: false,
        category: false,
        description: false,
        imageUrl: false,
      });
    }
  }, [isOpen, project]);

  // Helper to check if field is modified
  const isFieldModified = (fieldName: keyof typeof modifiedFields, currentValue: string): boolean => {
    return currentValue !== originalValues[fieldName];
  };

  // Update handlers that track modifications
  const handleProjectNameChange = (value: string) => {
    setProjectName(value);
    setModifiedFields(prev => ({
      ...prev,
      projectName: isFieldModified('projectName', value)
    }));
  };

  const handleCategoryChange = (value: string) => {
    setCategory(value);
    setModifiedFields(prev => ({
      ...prev,
      category: isFieldModified('category', value)
    }));
  };

  const handleDescriptionChange = (value: string) => {
    setDescription(value);
    setModifiedFields(prev => ({
      ...prev,
      description: isFieldModified('description', value)
    }));
  };

  const handleImageUrlChange = (value: string) => {
    setImageUrl(value);
    setModifiedFields(prev => ({
      ...prev,
      imageUrl: isFieldModified('imageUrl', value)
    }));
  };

  if (!isOpen) return null;

  const handleLaunch = async () => {
    await onLaunch({
      projectName,
      category,
      description,
      imageUrl,
    });
  };

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50"
        onClick={onClose}
      >
        <motion.div
          initial={{ scale: 0.9, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          exit={{ scale: 0.9, opacity: 0 }}
          onClick={(e) => e.stopPropagation()}
          className="bg-zinc-900 border border-white/10 rounded-sm p-6 max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto"
        >
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-accent/20 rounded-sm flex items-center justify-center">
                <Rocket size={20} className="text-accent" />
              </div>
              <div>
                <h3 className="text-xl font-mono text-white uppercase tracking-wide">
                  Launch to Walrus
                </h3>
                <p className="text-xs text-zinc-500 font-mono mt-1">
                  Deploy your site to Walrus Sites
                </p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="text-zinc-500 hover:text-white transition-colors"
              disabled={isProcessing}
            >
              <X size={20} />
            </button>
          </div>

          <div className="space-y-4">
            {/* Project Name */}
            <div>
              <label className="block text-xs font-mono text-zinc-400 uppercase mb-2 tracking-widest">
                Project Name {modifiedFields.projectName && <span className="text-accent/70 text-[10px]">(modified)</span>}
              </label>
              <input
                type="text"
                value={projectName}
                onChange={(e) => handleProjectNameChange(e.target.value)}
                autoComplete="off"
                autoCapitalize="off"
                autoCorrect="off"
                spellCheck="false"
                className={cn(
                  "w-full px-4 py-2 bg-black/40 border rounded-sm text-sm text-zinc-300 font-mono focus:outline-none transition-all",
                  modifiedFields.projectName
                    ? "border-accent/50 bg-accent/5 focus:border-accent/70"
                    : "border-white/10 focus:border-accent/50"
                )}
                placeholder="my-walgo-site"
              />
            </div>

            {/* Category */}
            <div>
              <label className="block text-xs font-mono text-zinc-400 uppercase mb-2 tracking-widest">
                Category {modifiedFields.category && <span className="text-accent/70 text-[10px]">(modified)</span>}
              </label>
              <input
                type="text"
                value={category}
                onChange={(e) => handleCategoryChange(e.target.value)}
                autoComplete="off"
                autoCapitalize="off"
                autoCorrect="off"
                spellCheck="false"
                className={cn(
                  "w-full px-4 py-2 bg-black/40 border rounded-sm text-sm text-zinc-300 font-mono focus:outline-none transition-all",
                  modifiedFields.category
                    ? "border-accent/50 bg-accent/5 focus:border-accent/70"
                    : "border-white/10 focus:border-accent/50"
                )}
                placeholder="website"
              />
            </div>

            {/* Description */}
            <div>
              <label className="block text-xs font-mono text-zinc-400 uppercase mb-2 tracking-widest">
                Description {modifiedFields.description && <span className="text-accent/70 text-[10px]">(modified)</span>}
              </label>
              <textarea
                value={description}
                onChange={(e) => handleDescriptionChange(e.target.value)}
                autoComplete="off"
                autoCapitalize="off"
                autoCorrect="off"
                spellCheck="false"
                className={cn(
                  "w-full px-4 py-2 bg-black/40 border rounded-sm text-sm text-zinc-300 font-mono focus:outline-none transition-all resize-none",
                  modifiedFields.description
                    ? "border-accent/50 bg-accent/5 focus:border-accent/70"
                    : "border-white/10 focus:border-accent/50"
                )}
                placeholder="A website deployed with Walgo to Walrus Sites"
                rows={3}
              />
            </div>

            {/* Image URL (Optional) */}
            <div>
              <label className="block text-xs font-mono text-zinc-400 uppercase mb-2 tracking-widest">
                Image URL (Optional) {modifiedFields.imageUrl && <span className="text-accent/70 text-[10px]">(modified)</span>}
              </label>
              <input
                type="text"
                value={imageUrl}
                onChange={(e) => handleImageUrlChange(e.target.value)}
                autoComplete="off"
                autoCapitalize="off"
                autoCorrect="off"
                spellCheck="false"
                className={cn(
                  "w-full px-4 py-2 bg-black/40 border rounded-sm text-sm text-zinc-300 font-mono focus:outline-none transition-all",
                  modifiedFields.imageUrl
                    ? "border-accent/50 bg-accent/5 focus:border-accent/70"
                    : "border-white/10 focus:border-accent/50"
                )}
                placeholder="https://..."
              />
            </div>

            {/* Actions */}
            <div className="flex gap-3 pt-2">
              <button
                onClick={onClose}
                disabled={isProcessing}
                className="flex-1 px-4 py-3 bg-white/5 hover:bg-white/10 border border-white/10 rounded-sm text-sm text-zinc-300 hover:text-white font-mono transition-all uppercase tracking-wide disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                onClick={handleLaunch}
                disabled={isProcessing || !projectName}
                className="flex-1 px-4 py-3 bg-accent/10 hover:bg-accent/20 border border-accent/30 rounded-sm text-sm text-accent font-mono transition-all uppercase tracking-wide disabled:opacity-50 flex items-center justify-center gap-2"
              >
                {isProcessing ? (
                  <>
                    <Loader2 size={14} className="animate-spin" />
                    Launching...
                  </>
                ) : (
                  <>
                    <Rocket size={14} />
                    Launch
                  </>
                )}
              </button>
            </div>
          </div>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  );
};
