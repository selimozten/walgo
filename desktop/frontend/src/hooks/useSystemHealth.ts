import { useState, useEffect } from 'react';
import { GetSystemHealth, CheckSetupDeps, GetVersion } from '../../wailsjs/go/main/App';
import { SystemHealth } from '../types';
import { HEALTH_CHECK_INTERVAL } from '../utils/constants';

export const useSystemHealth = () => {
    const [health, setHealth] = useState<SystemHealth>({
        netOnline: false,
        suiInstalled: false,
        suiConfigured: false,
        walrusInstalled: false,
        siteBuilder: false,
        hugoInstalled: false
    });
    const [version, setVersion] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const loadHealth = async () => {
        setLoading(true);
        setError(null);
        try {
            const result = await GetSystemHealth();
            setHealth(result);
        } catch (err) {
            console.error('Failed to load system health:', err);
            setError('Failed to check system health');
        } finally {
            setLoading(false);
        }
    };

    const loadVersion = async () => {
        try {
            const result = await GetVersion();
            if (result && result.version) {
                setVersion(result.version);
            }
        } catch (err) {
            console.error('Failed to load version:', err);
        }
    };

    const checkDeps = async () => {
        setLoading(true);
        setError(null);
        try {
            // Check what's missing (no auto-install anymore)
            const checkResult = await CheckSetupDeps();
            
            // Reload health to get latest status
            await loadHealth();
            
            return checkResult;
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, error: errorMsg };
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadHealth();
        loadVersion();
        const interval = setInterval(loadHealth, HEALTH_CHECK_INTERVAL);
        return () => clearInterval(interval);
    }, []);

    return {
        health,
        version,
        loading,
        error,
        reload: loadHealth,
        checkDeps
    };
};

