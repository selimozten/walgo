import React from 'react';
import { motion } from 'framer-motion';
import { Check, AlertCircle, X, Loader2 } from 'lucide-react';
import { Status } from '../../types';
import { cn } from '../../utils/helpers';

interface StatusBannerProps {
    status: Status | null;
    onClose?: () => void;
}

export const StatusBanner: React.FC<StatusBannerProps> = ({ status, onClose }) => {
    if (!status) return null;

    const getStatusIcon = () => {
        switch (status.type) {
            case 'success':
                return <Check size={16} className="text-green-400" />;
            case 'error':
                return <AlertCircle size={16} className="text-red-400" />;
            case 'info':
                return <Loader2 size={16} className="text-blue-400 animate-spin" />;
        }
    };

    const getStatusColor = () => {
        switch (status.type) {
            case 'success':
                return 'border-green-500/30 bg-green-500/10 text-green-300';
            case 'error':
                return 'border-red-500/30 bg-red-500/10 text-red-300';
            case 'info':
                return 'border-blue-500/30 bg-blue-500/10 text-blue-300';
        }
    };

    return (
        <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.2 }}
            className={cn(
                "fixed top-10 left-1/2 -translate-x-1/2 z-50 px-4 py-3 rounded-sm border backdrop-blur-sm flex items-center gap-3 shadow-lg",
                getStatusColor()
            )}
        >
            {getStatusIcon()}
            <span className="text-sm font-medium font-mono">{status.message}</span>
            {onClose && (
                <button
                    onClick={onClose}
                    className="ml-2 hover:opacity-70 transition-opacity"
                >
                    <X size={14} />
                </button>
            )}
        </motion.div>
    );
};

