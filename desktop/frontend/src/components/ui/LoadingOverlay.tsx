import React, { useEffect } from 'react';
import { motion } from 'framer-motion';
import WalgoLogo from '../../assets/walgo-Wlogo-no_background.svg';

interface LoadingOverlayProps {
    message?: string;
    fullScreen?: boolean;
    autoRefresh?: boolean;
    refreshInterval?: number; // in milliseconds
    onRefresh?: () => void;
}

export const LoadingOverlay: React.FC<LoadingOverlayProps> = ({ 
    message = 'Processing...', 
    fullScreen = false,
    autoRefresh = false,
    refreshInterval = 3000,
    onRefresh
}) => {
    // Auto-refresh functionality
    useEffect(() => {
        if (autoRefresh && onRefresh) {
            const interval = setInterval(() => {
                onRefresh();
            }, refreshInterval);

            return () => clearInterval(interval);
        }
    }, [autoRefresh, onRefresh, refreshInterval]);

    const containerClasses = fullScreen
        ? 'fixed inset-0 bg-black/80 backdrop-blur-sm z-[9999] flex items-center justify-center'
        : 'absolute inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center rounded-sm';

    return (
        <div className={containerClasses}>
            <div className="flex flex-col items-center gap-6">
                {/* Heartbeat Logo Animation */}
                <motion.div
                    animate={{
                        scale: [1, 1.12, 1],
                        opacity: [0.85, 1, 0.85],
                    }}
                    transition={{
                        duration: 2,
                        repeat: Infinity,
                        ease: [0.4, 0, 0.6, 1], // Custom cubic-bezier for smooth flow
                    }}
                    className="relative"
                >
                    <img 
                        src={WalgoLogo} 
                        alt="Walgo" 
                        className="w-20 h-20 drop-shadow-[0_0_20px_rgba(139,92,246,0.5)]"
                    />
                    
                    {/* Pulse rings */}
                    <motion.div
                        animate={{
                            scale: [0, 2.5],
                            opacity: [0, 0.5, 0],
                        }}
                        transition={{
                            duration: 2,
                            repeat: Infinity,
                            ease: "easeOut",
                            times: [0, 0.2, 1],
                        }}
                        className="absolute inset-0 rounded-full border-2 border-accent"
                    />
                    <motion.div
                        animate={{
                            scale: [0, 2.5],
                            opacity: [0, 0.5, 0],
                        }}
                        transition={{
                            duration: 2,
                            repeat: Infinity,
                            ease: "easeOut",
                            delay: 1,
                            times: [0, 0.2, 1],
                        }}
                        className="absolute inset-0 rounded-full border-2 border-accent"
                    />
                </motion.div>

                {/* Loading Message */}
                <div className="text-center">
                    <motion.p
                        animate={{
                            opacity: [0.6, 1, 0.6],
                        }}
                        transition={{
                            duration: 2,
                            repeat: Infinity,
                            ease: [0.4, 0, 0.6, 1],
                        }}
                        className="text-white font-mono text-sm mb-2"
                    >
                        {message}
                    </motion.p>
                    
                    {/* Loading dots */}
                    <div className="flex gap-1 justify-center">
                        {[0, 1, 2].map((i) => (
                            <motion.div
                                key={i}
                                animate={{
                                    y: [0, -10, 0],
                                    opacity: [0.4, 1, 0.4],
                                }}
                                transition={{
                                    duration: 1.2,
                                    repeat: Infinity,
                                    ease: [0.4, 0, 0.6, 1],
                                    delay: i * 0.2,
                                }}
                                className="w-2 h-2 bg-accent rounded-full"
                            />
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
};

