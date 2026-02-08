import React, { useEffect, useState } from 'react';
import { Minus, Square, X } from 'lucide-react';
import { WindowMinimise, WindowToggleMaximise, Quit, Environment } from '../../../wailsjs/runtime/runtime';

type Platform = 'windows' | 'linux' | 'darwin' | 'unknown';

export const WindowControls: React.FC = () => {
    const [platform, setPlatform] = useState<Platform>('unknown');
    const [isMaximized, setIsMaximized] = useState(false);

    useEffect(() => {
        // Use Wails Environment API for reliable platform detection
        Environment().then((env) => {
            setPlatform(env.platform as Platform);
        }).catch(() => {
            // Fallback: try to detect from userAgent
            const ua = navigator.userAgent.toLowerCase();
            if (ua.includes('mac os') || ua.includes('macos')) {
                setPlatform('darwin');
            } else if (ua.includes('linux')) {
                setPlatform('linux');
            } else if (ua.includes('windows')) {
                setPlatform('windows');
            }
        });
    }, []);

    const handleMaximizeToggle = () => {
        WindowToggleMaximise();
        setIsMaximized(!isMaximized);
    };

    // macOS: Use native traffic lights, don't show custom controls
    if (platform === 'darwin') {
        return null;
    }

    // Don't render until platform is detected
    if (platform === 'unknown') {
        return null;
    }

    // Windows/Linux: Show custom controls on the right side
    return (
        <div className="fixed top-0 right-0 z-[60] flex items-center h-10 px-2 gap-1 wails-nodrag">
            <button
                onClick={WindowMinimise}
                className="w-8 h-8 flex items-center justify-center text-zinc-400 hover:text-white hover:bg-white/10 rounded transition-colors"
                title="Minimize"
            >
                <Minus size={14} />
            </button>
            <button
                onClick={handleMaximizeToggle}
                className="w-8 h-8 flex items-center justify-center text-zinc-400 hover:text-white hover:bg-white/10 rounded transition-colors"
                title={isMaximized ? "Restore" : "Maximize"}
            >
                <Square size={12} />
            </button>
            <button
                onClick={Quit}
                className="w-8 h-8 flex items-center justify-center text-zinc-400 hover:text-white hover:bg-red-500/20 hover:text-red-400 rounded transition-colors"
                title="Close"
            >
                <X size={14} />
            </button>
        </div>
    );
};

