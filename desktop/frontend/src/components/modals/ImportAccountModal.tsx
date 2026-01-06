import React, { useState } from 'react';
import { X, Key, FileText, AlertCircle } from 'lucide-react';
import { LoadingOverlay } from '../ui';
import { motion } from 'framer-motion';
import { buttonVariants, iconButtonVariants } from '../../utils/constants';

interface ImportAccountModalProps {
    isOpen: boolean;
    onClose: () => void;
    onImport: (method: string, input: string, keyScheme?: string) => Promise<void>;
    isProcessing?: boolean;
}

export const ImportAccountModal: React.FC<ImportAccountModalProps> = ({
    isOpen,
    onClose,
    onImport,
    isProcessing
}) => {
    const [step, setStep] = useState<'scheme' | 'input'>('scheme');
    const [keyScheme, setKeyScheme] = useState('ed25519');
    const [method, setMethod] = useState<'mnemonic' | 'privateKey'>('mnemonic');
    const [input, setInput] = useState('');
    const [loading, setLoading] = useState(false);

    // Reset state when modal closes
    if (!isOpen && (step !== 'scheme' || input !== '' || keyScheme !== 'ed25519' || method !== 'mnemonic')) {
        setStep('scheme');
        setKeyScheme('ed25519');
        setMethod('mnemonic');
        setInput('');
        setLoading(false);
    }

    if (!isOpen) return null;

    const handleClose = () => {
        setStep('scheme');
        setKeyScheme('ed25519');
        setMethod('mnemonic');
        setInput('');
        setLoading(false);
        onClose();
    };

    const handleNext = () => {
        setStep('input');
    };

    const handleBack = () => {
        setStep('scheme');
    };

    const handleImport = async () => {
        setLoading(true);
        try {
            await onImport(method, input, keyScheme);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div 
            className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4"
            onClick={(e) => {
                // Only close if clicking the backdrop and not processing
                if (e.target === e.currentTarget && !isProcessing && !loading) {
                    handleClose();
                }
            }}
        >
            <div className="bg-zinc-900 border border-white/10 rounded-sm w-full max-w-md p-6 relative">
                {/* Loading Overlay */}
                {(isProcessing || loading) && <LoadingOverlay message="Importing account..." />}

                {/* Header */}
                <div className="flex items-center justify-between mb-6">
                    <div>
                        <h2 className="text-lg font-semibold text-white">Import Account</h2>
                        <p className="text-xs text-zinc-500 font-mono mt-1">
                            Step {step === 'scheme' ? '1' : '2'} of 2
                        </p>
                    </div>
                    <motion.button
                        onClick={handleClose}
                        className="text-zinc-500 hover:text-white transition-colors"
                        disabled={isProcessing || loading}
                        variants={iconButtonVariants}
                        whileHover="hover"
                        whileTap="tap"
                    >
                        <X size={20} />
                    </motion.button>
                </div>

                {step === 'scheme' ? (
                    <>
                        {/* Key Scheme Selection */}
                        <div className="space-y-4 mb-6">
                            <div>
                                <label className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest mb-3 block">
                                    Select Key Scheme
                                </label>
                                <div className="grid grid-cols-3 gap-3">
                                    {['ed25519', 'secp256k1', 'secp256r1'].map((scheme) => (
                                        <motion.button
                                            key={scheme}
                                            onClick={() => setKeyScheme(scheme)}
                                            className={`px-4 py-3 text-sm font-mono rounded-sm border transition-all ${
                                                keyScheme === scheme
                                                    ? 'bg-accent/10 border-accent text-accent'
                                                    : 'bg-black/20 border-white/10 hover:border-white/20 text-zinc-400'
                                            }`}
                                            disabled={isProcessing}
                                            variants={buttonVariants}
                                            whileHover="hover"
                                            whileTap="tap"
                                        >
                                            {scheme}
                                        </motion.button>
                                    ))}
                                </div>
                                <p className="text-xs text-zinc-500 font-mono mt-2">
                                    {keyScheme === 'ed25519' 
                                        ? 'Recommended for most use cases. Fast and secure. (m/44\'/784\'/0\'/0\'/0\')'
                                        : keyScheme === 'secp256k1'
                                        ? 'Compatible with Ethereum and Bitcoin wallets. (m/54\'/784\'/0\'/0/0)'
                                        : 'NIST P-256 curve, widely used in enterprise. (m/74\'/784\'/0\'/0/0)'}
                                </p>
                            </div>
                        </div>

                        {/* Info Box */}
                        <div className="bg-blue-500/10 border border-blue-500/30 rounded-sm p-4 mb-6 flex gap-3">
                            <AlertCircle size={20} className="text-blue-400 flex-shrink-0 mt-0.5" />
                            <div>
                                <p className="text-xs text-blue-300/80 font-mono">
                                    Choose the same key scheme that was used when the account was created. 
                                    If unsure, try ed25519 first.
                                </p>
                            </div>
                        </div>

                        {/* Actions */}
                        <div className="flex gap-3">
                            <motion.button
                                onClick={handleClose}
                                className="flex-1 px-4 py-2.5 bg-black/20 hover:bg-black/30 border border-white/10 text-sm font-mono text-zinc-300 rounded-sm transition-all"
                                disabled={isProcessing || loading}
                                variants={buttonVariants}
                                whileHover="hover"
                                whileTap="tap"
                            >
                                Cancel
                            </motion.button>
                            <motion.button
                                onClick={handleNext}
                                disabled={loading}
                                className="flex-1 px-4 py-2.5 bg-accent/10 hover:bg-accent/20 border border-accent/30 text-sm font-mono text-accent rounded-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                                variants={buttonVariants}
                                whileHover="hover"
                                whileTap="tap"
                            >
                                Next →
                            </motion.button>
                        </div>
                    </>
                ) : (
                    <>
                        {/* Import Method Selection */}
                        <div className="space-y-4 mb-6">
                            <div>
                                <label className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest mb-3 block">
                                    Import Method
                                </label>
                                <div className="grid grid-cols-2 gap-3">
                                    <motion.button
                                        onClick={() => setMethod('mnemonic')}
                                        className={`flex flex-col items-center gap-2 p-3 rounded-sm border transition-all ${
                                            method === 'mnemonic'
                                                ? 'bg-accent/10 border-accent text-accent'
                                                : 'bg-black/20 border-white/10 hover:border-white/20 text-zinc-400'
                                        }`}
                                        disabled={isProcessing}
                                        variants={buttonVariants}
                                        whileHover="hover"
                                        whileTap="tap"
                                    >
                                        <FileText size={20} />
                                        <span className="text-xs font-mono">Mnemonic</span>
                                    </motion.button>
                                    <motion.button
                                        onClick={() => setMethod('privateKey')}
                                        className={`flex flex-col items-center gap-2 p-3 rounded-sm border transition-all ${
                                            method === 'privateKey'
                                                ? 'bg-accent/10 border-accent text-accent'
                                                : 'bg-black/20 border-white/10 hover:border-white/20 text-zinc-400'
                                        }`}
                                        disabled={isProcessing}
                                        variants={buttonVariants}
                                        whileHover="hover"
                                        whileTap="tap"
                                    >
                                        <Key size={20} />
                                        <span className="text-xs font-mono">Private Key</span>
                                    </motion.button>
                                </div>
                            </div>

                            {/* Input Field */}
                            <div>
                                <label className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest mb-2 block">
                                    {method === 'mnemonic' ? 'Mnemonic Phrase (12-24 words)' : 'Private Key'}
                                </label>
                                <textarea
                                    value={input}
                                    onChange={(e) => setInput(e.target.value)}
                                    placeholder={
                                        method === 'mnemonic'
                                            ? 'Enter your 12 or 24 word recovery phrase...'
                                            : 'Enter your private key...'
                                    }
                                    rows={method === 'mnemonic' ? 3 : 2}
                                    autoComplete="off"
                                    autoCapitalize="off"
                                    autoCorrect="off"
                                    spellCheck="false"
                                    className="w-full bg-black/20 border border-white/10 rounded-sm px-3 py-2.5 text-sm font-mono text-white placeholder-zinc-600 focus:outline-none focus:border-accent/50 transition-all resize-none"
                                    disabled={isProcessing}
                                />
                            </div>

                            {/* Key Scheme Display */}
                            <div className="bg-accent/5 border border-accent/20 rounded-sm p-3">
                                <p className="text-xs text-zinc-400 font-mono">
                                    <span className="text-accent font-semibold">Key Scheme:</span> {keyScheme}
                                </p>
                            </div>
                        </div>

                        {/* Warning */}
                        <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-sm p-3 mb-6 flex gap-2">
                            <AlertCircle size={16} className="text-yellow-400 flex-shrink-0 mt-0.5" />
                            <p className="text-xs text-yellow-300/80 font-mono">
                                Never share your recovery phrase or private key with anyone.
                            </p>
                        </div>

                        {/* Actions */}
                        <div className="flex gap-3">
                            <motion.button
                                onClick={handleBack}
                                className="px-4 py-2.5 bg-black/20 hover:bg-black/30 border border-white/10 text-sm font-mono text-zinc-300 rounded-sm transition-all"
                                disabled={isProcessing || loading}
                                variants={buttonVariants}
                                whileHover="hover"
                                whileTap="tap"
                            >
                                ← Back
                            </motion.button>
                            <motion.button
                                onClick={handleImport}
                                disabled={!input.trim() || isProcessing || loading}
                                className="flex-1 px-4 py-2.5 bg-accent/10 hover:bg-accent/20 border border-accent/30 text-sm font-mono text-accent rounded-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                                variants={buttonVariants}
                                whileHover="hover"
                                whileTap="tap"
                            >
                                {(isProcessing || loading) ? 'Importing...' : 'Import Account'}
                            </motion.button>
                        </div>
                    </>
                )}
            </div>
        </div>
    );
};

