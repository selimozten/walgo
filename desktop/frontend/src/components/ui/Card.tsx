import React from 'react';
import { motion } from 'framer-motion';
import { cn } from '../../utils/helpers';

interface CardProps {
    children: React.ReactNode;
    className?: string;
    onClick?: () => void;
}

export const Card: React.FC<CardProps> = ({ children, className, onClick }) => (
    <motion.div
        whileHover={onClick ? { scale: 1.01 } : {}}
        onClick={onClick}
        className={cn(
            "glass-panel-tech rounded-sm p-6 relative overflow-hidden group",
            onClick && "cursor-pointer",
            className
        )}
    >
        <div className="scanline opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
        <div className="absolute top-0 left-0 w-2 h-2 border-t border-l border-white/20" />
        <div className="absolute top-0 right-0 w-2 h-2 border-t border-r border-white/20" />
        <div className="absolute bottom-0 left-0 w-2 h-2 border-b border-l border-white/20" />
        <div className="absolute bottom-0 right-0 w-2 h-2 border-b border-r border-white/20" />
        <div className="relative z-10">{children}</div>
    </motion.div>
);

