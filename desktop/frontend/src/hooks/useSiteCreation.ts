import { useState, useEffect } from "react";
import {
  QuickStart,
  InitSite,
  SelectDirectory,
  GetDefaultSitesDir,
  ProjectNameExists,
} from "../../wailsjs/go/main/App";
import { Project } from "../types";

interface UseSiteCreationReturn {
  siteName: string;
  setSiteName: (name: string) => void;
  parentDir: string;
  setParentDir: (dir: string) => void;
  siteNameExists: boolean;
  siteNameCheckLoading: boolean;
  isProcessing: boolean;
  handleQuickStart: (siteType?: string) => Promise<{
    success: boolean;
    sitePath?: string;
    error?: string;
  }>;
  handleInit: () => Promise<{ success: boolean; sitePath?: string; error?: string }>;
  handleSelectParentDir: () => Promise<void>;
  resetForm: () => void;
}

export const useSiteCreation = (
  onProjectCreated?: (sitePath: string) => void
): UseSiteCreationReturn => {
  const [siteName, setSiteName] = useState("");
  const [parentDir, setParentDir] = useState("");
  const [siteNameExists, setSiteNameExists] = useState(false);
  const [siteNameCheckLoading, setSiteNameCheckLoading] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);

  // Load default sites directory on mount
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

  // Check if project name exists (debounced)
  useEffect(() => {
    if (!siteName || siteName.length < 2) {
      setSiteNameExists(false);
      setSiteNameCheckLoading(false);
      return;
    }

    setSiteNameCheckLoading(true);
    const timeoutId = setTimeout(async () => {
      try {
        const result = await ProjectNameExists(siteName);
        setSiteNameExists(result);
      } catch (err) {
        console.error("Failed to check if project name exists:", err);
        setSiteNameExists(false);
      } finally {
        setSiteNameCheckLoading(false);
      }
    }, 500);

    return () => clearTimeout(timeoutId);
  }, [siteName]);

  const handleQuickStart = async (siteType?: string) => {
    if (!siteName) {
      return { success: false, error: "Site name required" };
    }

    setIsProcessing(true);
    try {
      const result = await QuickStart({
        parentDir: parentDir || "", // Empty string will use default walgo-sites directory
        siteName: siteName,
        siteType: siteType || "biolink",
        skipBuild: false,
      });

      if (result.success) {
        const sitePath = result.sitePath;
        resetForm();
        if (onProjectCreated && sitePath) {
          onProjectCreated(sitePath);
        }
        return { success: true, sitePath };
      } else {
        return { success: false, error: result.error || 'Unknown error' };
      }
    } catch (err) {
      return { success: false, error: err?.toString() || "Unknown error" };
    } finally {
      setIsProcessing(false);
    }
  };

  const handleInit = async () => {
    if (!siteName) {
      return { success: false, error: "Site name required" };
    }

    setIsProcessing(true);
    try {
      const result = await InitSite(parentDir || "", siteName);

      if (result.success) {
        const sitePath = result.sitePath;
        resetForm();
        if (onProjectCreated && sitePath) {
          onProjectCreated(sitePath);
        }
        return { success: true, sitePath };
      } else {
        return { success: false, error: result.error || 'Unknown error' };
      }
    } catch (err) {
      return { success: false, error: err?.toString() || "Unknown error" };
    } finally {
      setIsProcessing(false);
    }
  };

  const handleSelectParentDir = async () => {
    try {
      const dir = await SelectDirectory("Select Parent Directory");
      if (dir) {
        setParentDir(dir);
      }
    } catch (err) {
      console.error("Error selecting directory:", err);
    }
  };

  const resetForm = () => {
    setSiteName("");
    setParentDir("");
  };

  return {
    siteName,
    setSiteName,
    parentDir,
    setParentDir,
    siteNameExists,
    siteNameCheckLoading,
    isProcessing,
    handleQuickStart,
    handleInit,
    handleSelectParentDir,
    resetForm,
  };
};
