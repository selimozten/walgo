import React, { createContext, useContext, useState, useEffect } from "react";
import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime";

interface AIProgressState {
  isActive: boolean;
  siteName: string;
  phase: string;
  message: string;
  currentFile: string;
  current: number;
  total: number;
  progress: number;
}

interface AIProgressContextType {
  progressState: AIProgressState;
  isModalOpen: boolean;
  isMinimized: boolean;
  startProgress: (siteName: string) => void;
  completeProgress: () => void;
  openModal: () => void;
  closeModal: () => void;
  toggleMinimize: () => void;
}

const AIProgressContext = createContext<AIProgressContextType | undefined>(
  undefined
);

export const useAIProgress = () => {
  const context = useContext(AIProgressContext);
  if (!context) {
    throw new Error("useAIProgress must be used within AIProgressProvider");
  }
  return context;
};

export const AIProgressProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [progressState, setProgressState] = useState<AIProgressState>({
    isActive: false,
    siteName: "",
    phase: "",
    message: "",
    currentFile: "",
    current: 0,
    total: 0,
    progress: 0,
  });

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isMinimized, setIsMinimized] = useState(false);

  // Listen for progress events globally
  useEffect(() => {
    const handleProgress = (data: any) => {
      setProgressState((prev) => {
        const updates: Partial<AIProgressState> = { isActive: true };

        if (data.phase === "planning") {
          updates.phase = "Planning site structure...";
        } else if (data.phase === "generating") {
          updates.phase = "Generating content...";
        }

        if (data.message) {
          updates.message = data.message;
        }

        if (data.pagePath) {
          updates.currentFile = data.pagePath;
        }

        if (data.current !== undefined && data.total !== undefined) {
          updates.current = data.current;
          updates.total = data.total;
          updates.progress = data.total > 0 ? data.current / data.total : 0;
        }

        return { ...prev, ...updates };
      });
    };

    EventsOn("ai:progress", handleProgress);

    return () => {
      EventsOff("ai:progress");
    };
  }, []);

  const startProgress = (siteName: string) => {
    setProgressState({
      isActive: true,
      siteName,
      phase: "Initializing...",
      message: "",
      currentFile: "",
      current: 0,
      total: 0,
      progress: 0,
    });
    setIsModalOpen(true);
    setIsMinimized(false);
  };

  const completeProgress = () => {
    setProgressState((prev) => ({
      ...prev,
      isActive: false,
      phase: "Completed!",
    }));
    // Auto-close after 2 seconds
    setTimeout(() => {
      setIsModalOpen(false);
      setIsMinimized(false);
    }, 2000);
  };

  const openModal = () => {
    setIsModalOpen(true);
    setIsMinimized(false);
  };

  const closeModal = () => {
    if (progressState.isActive) {
      // If still in progress, minimize instead of close
      setIsMinimized(true);
      setIsModalOpen(false);
    } else {
      setIsModalOpen(false);
    }
  };

  const toggleMinimize = () => {
    if (isModalOpen) {
      setIsMinimized(true);
      setIsModalOpen(false);
    } else {
      setIsMinimized(false);
      setIsModalOpen(true);
    }
  };

  return (
    <AIProgressContext.Provider
      value={{
        progressState,
        isModalOpen,
        isMinimized,
        startProgress,
        completeProgress,
        openModal,
        closeModal,
        toggleMinimize,
      }}
    >
      {children}
    </AIProgressContext.Provider>
  );
};
