import React, { createContext, useContext, useState, useEffect, useRef, useCallback } from "react";
import { GetAIProgress } from "../../wailsjs/go/main/App";

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

interface AICompleteEvent {
  success?: boolean;
  sitePath?: string;
  totalPages?: number;
  filesCreated?: number;
  error?: string;
}

interface AIProgressContextType {
  progressState: AIProgressState;
  isModalOpen: boolean;
  isMinimized: boolean;
  startProgress: (siteName: string) => void;
  openModal: () => void;
  closeModal: () => void;
  toggleMinimize: () => void;
  completionResult: AICompleteEvent | null;
  clearCompletionResult: () => void;
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
  const [completionResult, setCompletionResult] = useState<AICompleteEvent | null>(null);
  const [polling, setPolling] = useState(false);
  const autoCloseTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Poll GetAIProgress() when polling is true
  useEffect(() => {
    if (!polling) return;

    const intervalId = setInterval(async () => {
      try {
        const data = await GetAIProgress();

        if (data.complete) {
          // Pipeline finished — stop polling and handle completion
          setPolling(false);

          setCompletionResult({
            success: data.success,
            sitePath: data.sitePath,
            totalPages: data.totalPages,
            filesCreated: data.filesCreated,
            error: data.error,
          });

          setProgressState((prev) => ({
            ...prev,
            isActive: false,
            phase: "Completed!",
          }));

          // Auto-close after 2 seconds
          if (autoCloseTimerRef.current) {
            clearTimeout(autoCloseTimerRef.current);
          }
          autoCloseTimerRef.current = setTimeout(() => {
            setIsModalOpen(false);
            setIsMinimized(false);
            autoCloseTimerRef.current = null;
          }, 2000);

          return;
        }

        if (data.isActive) {
          setProgressState((prev) => {
            const updates: Partial<AIProgressState> = { isActive: true };

            if (data.phase === "planning") {
              updates.phase = "Planning site structure...";
            } else if (data.phase === "generating") {
              updates.phase = "Generating content...";
            } else if (data.phase === "completed") {
              updates.phase = "Finishing up...";
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
        }
      } catch {
        // Ignore poll errors — will retry on next interval
      }
    }, 500);

    return () => clearInterval(intervalId);
  }, [polling]);

  const startProgress = useCallback((siteName: string) => {
    setCompletionResult(null);
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
    setPolling(true);
  }, []);

  const clearCompletionResult = useCallback(() => {
    setCompletionResult(null);
  }, []);

  // Clean up timers on unmount
  useEffect(() => {
    return () => {
      if (autoCloseTimerRef.current) {
        clearTimeout(autoCloseTimerRef.current);
      }
    };
  }, []);

  const openModal = useCallback(() => {
    setIsModalOpen(true);
    setIsMinimized(false);
  }, []);

  const closeModal = useCallback(() => {
    if (progressState.isActive) {
      setIsMinimized(true);
      setIsModalOpen(false);
    } else {
      setIsModalOpen(false);
    }
  }, [progressState.isActive]);

  const toggleMinimize = useCallback(() => {
    if (isModalOpen) {
      setIsMinimized(true);
      setIsModalOpen(false);
    } else {
      setIsMinimized(false);
      setIsModalOpen(true);
    }
  }, [isModalOpen]);

  return (
    <AIProgressContext.Provider
      value={{
        progressState,
        isModalOpen,
        isMinimized,
        startProgress,
        openModal,
        closeModal,
        toggleMinimize,
        completionResult,
        clearCompletionResult,
      }}
    >
      {children}
    </AIProgressContext.Provider>
  );
};
