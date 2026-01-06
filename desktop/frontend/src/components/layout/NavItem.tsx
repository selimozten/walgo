import React from 'react';
import { LucideIcon } from 'lucide-react';
import { cn } from '../../utils/helpers';

interface NavItemProps {
    id: string;
    icon: LucideIcon;
    label: string;
    activeTab: string;
    setActiveTab: (tab: string) => void;
}

export const NavItem: React.FC<NavItemProps> = ({ id, icon: Icon, label, activeTab, setActiveTab }) => (
    <button
        onClick={() => setActiveTab(id)}
        className={cn(
            "group relative flex items-center justify-center w-12 h-12 mb-4 transition-all duration-300",
            activeTab === id ? "text-accent" : "text-zinc-600 hover:text-zinc-400"
        )}
        title={label}
    >
        <div className={cn(
            "absolute inset-0 bg-accent/5 rounded-md scale-0 transition-transform duration-300",
            activeTab === id && "scale-100"
        )} />
        <div className={cn(
            "absolute left-0 w-1 h-8 bg-accent rounded-r-full transition-all duration-300",
            activeTab === id ? "opacity-100 translate-x-0" : "opacity-0 -translate-x-2"
        )} />
        <Icon size={20} strokeWidth={1.5} className="relative z-10" />
    </button>
);

