import React, { useEffect, useState } from 'react';
import { Minus, Square, X, Maximize2, Minimize2 } from 'lucide-react';
import { WindowMinimise, WindowToggleMaximise, Quit, WindowFullscreen } from '../../../wailsjs/runtime/runtime';

export const WindowControls: React.FC = () => {
    const [isMac, setIsMac] = useState(false);
    const [isFullscreen, setIsFullscreen] = useState(false);

    useEffect(() => {
        // Detect macOS using userAgent (platform is deprecated)
        const userAgent = navigator.userAgent.toLowerCase();
        setIsMac(userAgent.indexOf('mac') >= 0);
    }, []);

    const handleFullscreenToggle = () => {
        WindowFullscreen();
        setIsFullscreen(!isFullscreen);
    };

    // macOS: Use native traffic lights, don't show custom controls
    if (isMac) {
        return null;
    }

    // Windows/Linux: Show custom controls on the right side
    return (
        <div className="fixed top-0 right-0 z-[60] flex items-center h-10 px-2 gap-1">
        <button
            onClick={WindowMinimise}
                className="w-8 h-8 flex items-center justify-center text-zinc-400 hover:text-white hover:bg-white/10 rounded transition-colors"
                title="Minimize"
        >
            <Minus size={14} />
        </button>
        <button
                onClick={handleFullscreenToggle}
                className="w-8 h-8 flex items-center justify-center text-zinc-400 hover:text-white hover:bg-white/10 rounded transition-colors"
                title={isFullscreen ? "Exit Fullscreen" : "Enter Fullscreen"}
        >
                {isFullscreen ? <Minimize2 size={14} /> : <Maximize2 size={14} />}
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

