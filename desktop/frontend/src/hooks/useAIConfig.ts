import { useState, useEffect } from 'react';
import { 
    GetAIConfig, 
    UpdateAIConfig as UpdateAIConfigAPI, 
    CleanAIConfig, 
    CleanProviderConfig,
    GetProviderCredentials 
} from '../../wailsjs/go/main/App';
import { AIConfig } from '../types';

export const useAIConfig = () => {
    const [configured, setConfigured] = useState(false);
    const [config, setConfig] = useState<AIConfig | null>(null);
    const [provider, setProvider] = useState('');
    const [apiKey, setApiKey] = useState('');
    const [baseUrl, setBaseUrl] = useState('');
    const [model, setModel] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const loadConfig = async () => {
        setLoading(true);
        setError(null);
        try {
            const result = await GetAIConfig();
            if (result.success) {
                setConfigured(result.enabled);
                setConfig(result);
                if (result.currentProvider) {
                    setProvider(result.currentProvider);
                    setModel(result.currentModel || '');
                    await loadProviderCredentials(result.currentProvider);
                }
            }
        } catch (err) {
            console.error('Failed to load AI config:', err);
            setError('Failed to load AI configuration');
        } finally {
            setLoading(false);
        }
    };

    const loadProviderCredentials = async (prov: string) => {
        if (!prov) {
            setApiKey('');
            setBaseUrl('');
            setModel('');
            return;
        }

        try {
            const result = await GetProviderCredentials(prov);
            const key = result.success ? (result.apiKey || '') : '';
            const url = result.success ? (result.baseURL || '') : '';
            const mdl = result.success ? (result.model || '') : '';

            setApiKey(key);
            setBaseUrl(url);
            setModel(mdl);
        } catch (err) {
            console.error('Failed to load provider credentials:', err);
        }
    };

    const updateConfig = async (providerName: string, credentials: {
        apiKey?: string;
        baseURL?: string;
        model?: string;
    }) => {
        setLoading(true);
        setError(null);
        try {
            await UpdateAIConfigAPI({
                provider: providerName,
                apiKey: credentials.apiKey || '',
                baseURL: credentials.baseURL,
                model: credentials.model,
            });
            await loadConfig();
            return { success: true };
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, error: errorMsg };
        } finally {
            setLoading(false);
        }
    };

    const cleanConfig = async () => {
        setLoading(true);
        setError(null);
        try {
            await CleanAIConfig();
            await loadConfig();
            return { success: true };
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, error: errorMsg };
        } finally {
            setLoading(false);
        }
    };

    const cleanProvider = async (providerName: string) => {
        setLoading(true);
        setError(null);
        try {
            await CleanProviderConfig(providerName);
            await loadConfig();
            return { success: true };
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, error: errorMsg };
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadConfig();
    }, []);

    return {
        configured,
        config,
        provider,
        apiKey,
        baseUrl,
        model,
        setProvider,
        setApiKey,
        setBaseUrl,
        setModel,
        loading,
        error,
        loadConfig,
        loadProviderCredentials,
        updateConfig,
        cleanConfig,
        cleanProvider
    };
};

