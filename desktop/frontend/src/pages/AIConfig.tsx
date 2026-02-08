import React, { useState, useEffect, useRef } from 'react';
import { Wand2, Save, Trash2, Eye, EyeOff, Check, AlertCircle, ExternalLink } from 'lucide-react';
import { Card } from '../components/ui/Card';
import { LoadingOverlay } from '../components/ui';
import { AI_PROVIDERS, buttonVariants, iconButtonVariants } from '../utils/constants';
import { useAIConfig } from '../hooks/useAIConfig';
import { cn } from '../utils/helpers';
import { motion } from 'framer-motion';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';

interface AIConfigProps {
    onConfigChange?: () => void;
}

export const AIConfig: React.FC<AIConfigProps> = ({ onConfigChange }) => {
    const {
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
        loadProviderCredentials,
        updateConfig,
        cleanConfig,
        cleanProvider
    } = useAIConfig();

    const [showApiKey, setShowApiKey] = useState(false);
    const [hasChanges, setHasChanges] = useState(false);
    const [status, setStatus] = useState<{ type: 'success' | 'error', message: string } | null>(null);
    const [isProcessing, setIsProcessing] = useState(false);
    const [previousProvider, setPreviousProvider] = useState('');
    const [isInitialLoad, setIsInitialLoad] = useState(true);

    // Track initial values to detect changes - use ref to avoid infinite loops
    const initialValuesRef = useRef({ provider: '', apiKey: '', baseUrl: '', model: '' });

    // Track if we just loaded credentials
    const [justLoadedCredentials, setJustLoadedCredentials] = useState(false);

    // Track which fields have changed
    const [changedFields, setChangedFields] = useState({
        apiKey: false,
        baseUrl: false,
        model: false
    });

    // Set initial values on first load
    useEffect(() => {
        if (isInitialLoad && provider && !loading) {
            initialValuesRef.current = { provider, apiKey, baseUrl, model };
            setPreviousProvider(provider);
            setIsInitialLoad(false);
        }
    }, [isInitialLoad, provider, apiKey, baseUrl, model, loading]);

    // Load provider credentials when provider changes
    useEffect(() => {
        const handleProviderChange = async () => {
            if (provider && provider !== previousProvider) {
                setIsProcessing(true);
                setStatus({ type: 'success', message: `Loading ${provider} configuration...` });
                
                try {
                    // Load credentials for the provider
                    await loadProviderCredentials(provider);
                    setStatus(null);
                    setJustLoadedCredentials(true);
                } catch (err) {
                    console.error('Failed to load provider credentials:', err);
                    setStatus({ type: 'error', message: 'Failed to load provider credentials' });
                } finally {
                    setIsProcessing(false);
                    setPreviousProvider(provider);
                }
            }
        };

        handleProviderChange();
    }, [provider, previousProvider, loadProviderCredentials]);

    // Update initial values ref ONLY when credentials are just loaded
    useEffect(() => {
        if (justLoadedCredentials && !isProcessing) {
            initialValuesRef.current = { provider, apiKey, baseUrl, model };
            setChangedFields({ apiKey: false, baseUrl: false, model: false });
            setJustLoadedCredentials(false);
        }
    }, [justLoadedCredentials, isProcessing, provider, apiKey, baseUrl, model]);

    // Detect changes
    useEffect(() => {
        const apiKeyChanged = apiKey !== initialValuesRef.current.apiKey;
        const baseUrlChanged = baseUrl !== initialValuesRef.current.baseUrl;
        const modelChanged = model !== initialValuesRef.current.model;
        
        const changed = apiKeyChanged || baseUrlChanged || modelChanged;
        
        setChangedFields({
            apiKey: apiKeyChanged,
            baseUrl: baseUrlChanged,
            model: modelChanged
        });
        
        setHasChanges(changed);
    }, [apiKey, baseUrl, model]);

    const handleUpdate = async () => {
        if (!provider) {
            setStatus({ type: 'error', message: 'No provider selected' });
            return;
        }

        if (!apiKey) {
            setStatus({ type: 'error', message: 'API key required' });
            return;
        }

        setIsProcessing(true);
        setStatus({ type: 'success', message: 'Saving configuration...' });
        
        try {
            // First, clear all other providers if there's a configured provider different from current
            if (configured && config?.currentProvider && config.currentProvider !== provider) {
                setStatus({ type: 'success', message: `Clearing ${config.currentProvider} configuration...` });
                const clearResult = await cleanProvider(config.currentProvider);
                if (!clearResult.success) {
                    setStatus({ type: 'error', message: `Failed to clear ${config.currentProvider}` });
                    return;
                }
            }

            // Now save the new configuration
            setStatus({ type: 'success', message: 'Saving configuration...' });
            const result = await updateConfig(provider, { apiKey, baseURL: baseUrl, model });
            
            if (result.success) {
                setStatus({ type: 'success', message: 'Configuration saved successfully! Only one provider can be active at a time.' });
                initialValuesRef.current = { provider, apiKey, baseUrl, model };
                setHasChanges(false);
                setChangedFields({ apiKey: false, baseUrl: false, model: false });
                // Notify parent component that config changed
                if (onConfigChange) {
                    onConfigChange();
                }
            } else {
                setStatus({ type: 'error', message: result.error || 'Save failed' });
            }
        } finally {
            setIsProcessing(false);
        }
    };


    const handleClean = async () => {
        setIsProcessing(true);
        setStatus({ type: 'success', message: 'Cleaning...' });
        
        try {
            const result = await cleanConfig();
            
            if (result.success) {
                setStatus({ type: 'success', message: 'All configurations cleared' });
                // Notify parent component that config changed
                if (onConfigChange) {
                    onConfigChange();
                }
            } else {
                setStatus({ type: 'error', message: result.error || 'Clean failed' });
            }
        } finally {
            setIsProcessing(false);
        }
    };

    const handleCleanProvider = async () => {
        if (!provider) return;
        
        setIsProcessing(true);
        setStatus({ type: 'success', message: 'Cleaning...' });
        
        try {
            const result = await cleanProvider(provider);
            
            if (result.success) {
                setStatus({ type: 'success', message: 'Provider configuration cleared' });
                // Notify parent component that config changed
                if (onConfigChange) {
                    onConfigChange();
                }
            } else {
                setStatus({ type: 'error', message: result.error || 'Clean failed' });
            }
        } finally {
            setIsProcessing(false);
        }
    };

    // Provider-specific help text
    const getProviderHelp = (prov: string) => {
        const helpTexts: Record<string, { description: string; link: string; linkText: string }> = {
            openai: {
                description: 'Get your API key from OpenAI platform. Supports GPT-4, GPT-4o, GPT-3.5-turbo, and other models.',
                link: 'https://platform.openai.com/api-keys',
                linkText: 'Get OpenAI API Key'
            },
            openrouter: {
                description: 'OpenRouter provides access to multiple AI models through a single API. Get your API key from OpenRouter platform.',
                link: 'https://openrouter.ai/keys',
                linkText: 'Get OpenRouter API Key'
            }
        };
        return helpTexts[prov] || null;
    };

    return (
        <>
            {isProcessing && <LoadingOverlay message="Processing..." />}
            
            <div className="space-y-6">
                {/* Header */}
                <div className="flex items-center justify-between mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-white mb-1">AI Configuration</h1>
                    <p className="text-zinc-400 text-sm">Configure your AI providers for content generation</p>
                </div>
                <div className={cn(
                    "px-3 py-1 rounded-sm text-sm font-mono",
                    configured ? "bg-green-500/10 text-green-400" : "bg-zinc-800 text-zinc-400"
                )}>
                    {configured ? <Check size={14} className="inline mr-2" /> : null}
                    {configured ? 'CONFIGURED' : 'NOT CONFIGURED'}
                </div>
            </div>

            {/* Status Banner */}
            {status && (
                <div className={cn(
                    "px-4 py-3 rounded-sm border backdrop-blur-sm",
                    status.type === 'success' 
                        ? "border-green-500/30 bg-green-500/10 text-green-300"
                        : "border-red-500/30 bg-red-500/10 text-red-300"
                )}>
                    {status.type === 'success' ? (
                        <Check size={16} className="inline mr-2" />
                    ) : (
                        <AlertCircle size={16} className="inline mr-2" />
                    )}
                    {status.message}
                </div>
            )}

            {/* Provider Selection */}
            <Card>
                <div className="flex items-center gap-2 mb-4">
                    <Wand2 size={20} className="text-accent" />
                    <h2 className="text-lg font-semibold text-white">AI Provider</h2>
                </div>

                <div className="space-y-4">
                    <div>
                        <label className="block text-xs font-mono text-zinc-400 uppercase tracking-wider mb-2">
                            Select Provider
                        </label>
                        <select
                            value={provider}
                            onChange={(e) => setProvider(e.target.value)}
                            className="w-full px-4 py-2 bg-zinc-900 border border-zinc-700 rounded-sm text-sm text-white focus:border-accent focus:outline-none transition-colors"
                        >
                            <option value="">Choose a provider...</option>
                            {AI_PROVIDERS.map((prov) => (
                                <option key={prov.value} value={prov.value}>
                                    {prov.label}
                                </option>
                            ))}
                        </select>
                    </div>

                    {provider && (
                        <div className="space-y-4 pt-4 border-t border-zinc-800">
                            {/* Warning when switching providers */}
                            {configured && previousProvider && provider !== previousProvider && (
                                <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-sm p-3 flex items-start gap-2">
                                    <AlertCircle size={16} className="text-yellow-400 flex-shrink-0 mt-0.5" />
                                    <div className="flex-1">
                                        <p className="text-xs text-yellow-300 font-mono font-semibold mb-1">
                                            Provider Switch Detected
                                        </p>
                                        <p className="text-xs text-yellow-300/80 font-mono">
                                            Only one AI provider can be active at a time. The previous provider ({previousProvider}) has been automatically cleared.
                                        </p>
                                    </div>
                                </div>
                            )}

                            {/* Provider Help Text */}
                            {getProviderHelp(provider) && (
                                <div className="bg-blue-500/10 border border-blue-500/30 rounded-sm p-3 flex items-start gap-2">
                                    <AlertCircle size={16} className="text-blue-400 flex-shrink-0 mt-0.5" />
                                    <div className="flex-1">
                                        <p className="text-xs text-blue-300 font-mono">
                                            {getProviderHelp(provider)!.description}
                                        </p>
                                        {getProviderHelp(provider)!.link && (
                                            <motion.button
                                                onClick={() => BrowserOpenURL(getProviderHelp(provider)!.link)}
                                                className="inline-flex items-center gap-1 text-xs text-blue-400 hover:text-blue-300 font-mono mt-1 cursor-pointer"
                                                variants={buttonVariants}
                                                whileHover="hover"
                                                whileTap="tap"
                                            >
                                                {getProviderHelp(provider)!.linkText}
                                                <ExternalLink size={12} />
                                            </motion.button>
                                        )}
                                    </div>
                                </div>
                            )}

                            <div>
                                <label className="block text-xs font-mono text-zinc-400 uppercase tracking-wider mb-2">
                                    API Key
                                </label>
                                <div className="relative">
                                    <input
                                        type={showApiKey ? "text" : "password"}
                                        value={apiKey}
                                        onChange={(e) => setApiKey(e.target.value)}
                                        placeholder="Enter your API key"
                                        autoComplete="off"
                                        autoCapitalize="off"
                                        autoCorrect="off"
                                        spellCheck="false"
                                        className={cn(
                                            "w-full px-4 py-2 bg-zinc-900 border rounded-sm text-sm text-white placeholder-zinc-500 focus:outline-none transition-colors pr-10",
                                            changedFields.apiKey 
                                                ? "border-green-500 focus:border-green-400" 
                                                : "border-zinc-700 focus:border-accent"
                                        )}
                                    />
                                    <motion.button
                                        type="button"
                                        onClick={() => setShowApiKey(!showApiKey)}
                                        className="absolute right-2 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300"
                                        variants={iconButtonVariants}
                                        whileHover="hover"
                                        whileTap="tap"
                                    >
                                        {showApiKey ? <EyeOff size={16} /> : <Eye size={16} />}
                                    </motion.button>
                                </div>
                            </div>

                            <div>
                                <label className="block text-xs font-mono text-zinc-400 uppercase tracking-wider mb-2">
                                    Base URL <span className="text-zinc-500">(optional)</span>
                                </label>
                                <input
                                    type="text"
                                    value={baseUrl}
                                    onChange={(e) => setBaseUrl(e.target.value)}
                                    placeholder={provider === 'openrouter' ? 'https://openrouter.ai/api/v1' : 'https://api.openai.com/v1'}
                                    autoComplete="off"
                                    autoCapitalize="off"
                                    autoCorrect="off"
                                    spellCheck="false"
                                    className={cn(
                                        "w-full px-4 py-2 bg-zinc-900 border rounded-sm text-sm text-white placeholder-zinc-500 focus:outline-none transition-colors",
                                        changedFields.baseUrl 
                                            ? "border-green-500 focus:border-green-400" 
                                            : "border-zinc-700 focus:border-accent"
                                    )}
                                />
                            </div>

                            <div>
                                <label className="block text-xs font-mono text-zinc-400 uppercase tracking-wider mb-2">
                                    Model <span className="text-zinc-500">(optional)</span>
                                </label>
                                <input
                                    type="text"
                                    value={model}
                                    onChange={(e) => setModel(e.target.value)}
                                    placeholder={provider === 'openrouter' ? 'openai/gpt-4, anthropic/claude-3, etc.' : 'gpt-4o, gpt-4-turbo, gpt-3.5-turbo'}
                                    autoComplete="off"
                                    autoCapitalize="off"
                                    autoCorrect="off"
                                    spellCheck="false"
                                    className={cn(
                                        "w-full px-4 py-2 bg-zinc-900 border rounded-sm text-sm text-white placeholder-zinc-500 focus:outline-none transition-colors",
                                        changedFields.model 
                                            ? "border-green-500 focus:border-green-400" 
                                            : "border-zinc-700 focus:border-accent"
                                    )}
                                />
                            </div>

                            <div className="flex gap-2 pt-4">
                                <motion.button
                                    onClick={handleUpdate}
                                    disabled={loading || !hasChanges || !apiKey}
                                    className="flex-1 px-4 py-2 bg-accent hover:bg-accent/90 text-black rounded-sm text-sm font-semibold transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                                    variants={buttonVariants}
                                    whileHover="hover"
                                    whileTap="tap"
                                >
                                    <Save size={16} />
                                    Save Configuration
                                </motion.button>
                                <motion.button
                                    onClick={handleCleanProvider}
                                    disabled={loading}
                                    className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 rounded-sm text-sm font-semibold transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                                    variants={buttonVariants}
                                    whileHover="hover"
                                    whileTap="tap"
                                >
                                    <Trash2 size={16} />
                                    Clear Provider
                                </motion.button>
                            </div>
                        </div>
                    )}
                </div>
            </Card>

            {/* Clear All Configuration */}
            {configured && (
                <Card>
                    <div className="flex items-center justify-between">
                        <div>
                            <h3 className="text-sm font-semibold text-white mb-1">Clear All Configuration</h3>
                            <p className="text-xs text-zinc-400">Remove all AI provider configurations</p>
                        </div>
                        <motion.button
                            onClick={handleClean}
                            disabled={loading}
                            className="px-4 py-2 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/30 rounded-sm text-sm font-semibold transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                            variants={buttonVariants}
                            whileHover="hover"
                            whileTap="tap"
                        >
                            <Trash2 size={16} />
                            Clear All
                        </motion.button>
                    </div>
                </Card>
            )}
            </div>
        </>
    );
};

