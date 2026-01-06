import { useState, useCallback } from 'react';
import { CheckToolVersions, UpdateTools } from '../../wailsjs/go/main/App';

export interface ToolVersionInfo {
    tool: string;
    currentVersion: string;
    latestVersion: string;
    updateRequired: boolean;
    installed: boolean;
}

export interface VersionCheckResult {
    success: boolean;
    tools: ToolVersionInfo[];
    message: string;
    error?: string;
}

export const useVersionCheck = () => {
    const [versions, setVersions] = useState<ToolVersionInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const [updating, setUpdating] = useState<string[]>([]);
    const [error, setError] = useState<string | null>(null);

    const checkVersions = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const result = await CheckToolVersions();
            if (result.success) {
                setVersions(result.tools || []);
            } else {
                setError(result.error || 'Failed to check versions');
            }
        } catch (err) {
            console.error('Failed to check versions:', err);
            setError('Failed to check versions');
        } finally {
            setLoading(false);
        }
    }, []);

    const updateTool = useCallback(async (tool: string, network: string = 'testnet') => {
        setUpdating(prev => [...prev, tool]);
        setError(null);
        try {
            const result = await UpdateTools({
                tools: [tool],
                network: network
            });
            
            if (result.success) {
                // Refresh versions after update
                await checkVersions();
                return { success: true, message: result.message };
            } else {
                const errorMsg = result.failedTools?.[tool] || 'Update failed';
                setError(errorMsg);
                return { success: false, message: errorMsg };
            }
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, message: errorMsg };
        } finally {
            setUpdating(prev => prev.filter(t => t !== tool));
        }
    }, [checkVersions]);

    const updateMultipleTools = useCallback(async (tools: string[], network: string = 'testnet') => {
        setUpdating(tools);
        setError(null);
        try {
            const result = await UpdateTools({
                tools: tools,
                network: network
            });
            
            // Refresh versions after update
            await checkVersions();
            
            return result;
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, message: errorMsg, updatedTools: [], failedTools: {} };
        } finally {
            setUpdating([]);
        }
    }, [checkVersions]);

    const getToolVersion = useCallback((toolName: string): ToolVersionInfo | undefined => {
        return versions.find(v => v.tool === toolName);
    }, [versions]);

    const hasUpdates = useCallback((excludeHugo: boolean = false): boolean => {
        return versions.some(v => {
            if (excludeHugo && v.tool === 'hugo') return false;
            return v.updateRequired;
        });
    }, [versions]);

    const getSuiToolsWithUpdates = useCallback((): string[] => {
        return versions
            .filter(v => ['sui', 'walrus', 'site-builder'].includes(v.tool) && v.updateRequired)
            .map(v => v.tool);
    }, [versions]);

    return {
        versions,
        loading,
        updating,
        error,
        checkVersions,
        updateTool,
        updateMultipleTools,
        getToolVersion,
        hasUpdates,
        getSuiToolsWithUpdates
    };
};

