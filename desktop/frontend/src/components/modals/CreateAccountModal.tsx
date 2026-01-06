import React, { useState } from 'react';
import { X, FileText, Copy, Check, AlertCircle } from 'lucide-react';
import { LoadingOverlay } from '../ui';
import { motion } from 'framer-motion';
import { buttonVariants, iconButtonVariants } from '../../utils/constants';

interface CreateAccountModalProps {
    isOpen: boolean;
    onClose: () => void;
    onCreate: (keyScheme: string, alias?: string) => Promise<{ success: boolean; mnemonic?: string; address?: string; error?: string }>;
    isProcessing?: boolean;
}

export const CreateAccountModal: React.FC<CreateAccountModalProps> = ({
    isOpen,
    onClose,
    onCreate,
    isProcessing
}) => {
    const [step, setStep] = useState<'select' | 'result'>('select');
    const [keyScheme, setKeyScheme] = useState('ed25519');
    const [alias, setAlias] = useState('');
    const [result, setResult] = useState<{ mnemonic?: string; address?: string } | null>(null);
    const [copiedItem, setCopiedItem] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);

    // Reset state when modal closes
    if (!isOpen && (step !== 'select' || alias !== '' || keyScheme !== 'ed25519' || result !== null)) {
        setStep('select');
        setKeyScheme('ed25519');
        setAlias('');
        setResult(null);
        setCopiedItem(null);
        setLoading(false);
    }

    if (!isOpen) return null;

    const handleCreate = async () => {
        setLoading(true);
        try {
            const response = await onCreate(keyScheme, alias);
            if (response.success && response.mnemonic && response.address) {
                setResult({
                    mnemonic: response.mnemonic,
                    address: response.address
                });
                setStep('result');
            }
        } finally {
            setLoading(false);
        }
    };

    const handleCopy = async (text: string, item: string) => {
        try {
            await navigator.clipboard.writeText(text);
            setCopiedItem(item);
            setTimeout(() => setCopiedItem(null), 2000);
        } catch (err) {
            console.error('Failed to copy:', err);
        }
    };

    const handleClose = () => {
        setStep('select');
        setKeyScheme('ed25519');
        setAlias('');
        setResult(null);
        setCopiedItem(null);
        setLoading(false);
        onClose();
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
            <div className="bg-zinc-900 border border-white/10 rounded-sm w-full max-w-2xl p-6 max-h-[90vh] overflow-y-auto relative">
                {/* Loading Overlay */}
                {(isProcessing || loading) && <LoadingOverlay message="Creating account..." />}

                {/* Header */}
                <div className="flex items-center justify-between mb-6">
                    <h2 className="text-lg font-semibold text-white">
                        {step === 'select' ? 'Create New Account' : 'Account Created Successfully'}
                    </h2>
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

                {step === 'select' ? (
                    <>
                        {/* Key Scheme Selection */}
                        <div className="space-y-4 mb-6">
                            <div>
                                <label className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest mb-2 block">
                                    Key Scheme
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

                            {/* Optional Alias */}
                            <div>
                                <label className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest mb-2 block">
                                    Alias (Optional)
                                </label>
                                <input
                                    type="text"
                                    value={alias}
                                    onChange={(e) => setAlias(e.target.value)}
                                    placeholder="e.g., My Main Account"
                                    autoComplete="off"
                                    autoCapitalize="off"
                                    autoCorrect="off"
                                    spellCheck="false"
                                    className="w-full bg-black/20 border border-white/10 rounded-sm px-3 py-2.5 text-sm font-mono text-white placeholder-zinc-600 focus:outline-none focus:border-accent/50 transition-all"
                                    disabled={isProcessing}
                                />
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
                                onClick={handleCreate}
                                disabled={isProcessing || loading}
                                className="flex-1 px-4 py-2.5 bg-accent/10 hover:bg-accent/20 border border-accent/30 text-sm font-mono text-accent rounded-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                                variants={buttonVariants}
                                whileHover="hover"
                                whileTap="tap"
                            >
                                {(isProcessing || loading) ? 'Creating...' : 'Create Account'}
                            </motion.button>
                        </div>
                    </>
                ) : (
                    <>
                        {/* Warning Message */}
                        <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-sm p-4 mb-6 flex gap-3">
                            <AlertCircle size={20} className="text-yellow-400 flex-shrink-0 mt-0.5" />
                            <div>
                                <p className="text-sm text-yellow-200 font-semibold mb-1">
                                    Save Your Recovery Information
                                </p>
                                <p className="text-xs text-yellow-300/80 font-mono">
                                    Store your mnemonic phrase and private key in a secure location. 
                                    You'll need them to recover your account. Never share them with anyone.
                                </p>
                            </div>
                        </div>

                        {/* Account Details */}
                        <div className="space-y-4 mb-6">
                            {/* Address */}
                            <div>
                                <label className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest mb-2 block">
                                    Account Address
                                </label>
                                <div className="bg-black/20 border border-white/10 rounded-sm p-3 flex items-center justify-between gap-3">
                                    <span className="text-sm font-mono text-white break-all">
                                        {result?.address}
                                    </span>
                                    <motion.button
                                        onClick={() => handleCopy(result?.address || '', 'address')}
                                        className="flex-shrink-0 p-2 hover:bg-white/5 rounded-sm transition-colors"
                                        title="Copy address"
                                        variants={iconButtonVariants}
                                        whileHover="hover"
                                        whileTap="tap"
                                    >
                                        {copiedItem === 'address' ? (
                                            <Check size={16} className="text-green-400" />
                                        ) : (
                                            <Copy size={16} className="text-zinc-400" />
                                        )}
                                    </motion.button>
                                </div>
                            </div>

                            {/* Mnemonic Phrase */}
                            <div>
                                <label className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest mb-2 flex items-center gap-2">
                                    <FileText size={12} />
                                    Mnemonic Phrase (12 Words)
                                </label>
                                <div className="bg-black/20 border border-white/10 rounded-sm p-3">
                                    <div className="flex items-start justify-between gap-3 mb-2">
                                        <p className="text-sm font-mono text-white break-all leading-relaxed">
                                            {result?.mnemonic}
                                        </p>
                                        <motion.button
                                            onClick={() => handleCopy(result?.mnemonic || '', 'mnemonic')}
                                            className="flex-shrink-0 p-2 hover:bg-white/5 rounded-sm transition-colors"
                                            title="Copy mnemonic"
                                            variants={iconButtonVariants}
                                            whileHover="hover"
                                            whileTap="tap"
                                        >
                                            {copiedItem === 'mnemonic' ? (
                                                <Check size={16} className="text-green-400" />
                                            ) : (
                                                <Copy size={16} className="text-zinc-400" />
                                            )}
                                        </motion.button>
                                    </div>
                                    <p className="text-xs text-zinc-500 font-mono">
                                        Use this phrase to recover your account in any wallet
                                    </p>
                                </div>
                            </div>


                            {/* Key Scheme Info */}
                            <div className="bg-accent/5 border border-accent/20 rounded-sm p-3">
                                <p className="text-xs text-zinc-400 font-mono">
                                    <span className="text-accent font-semibold">Key Scheme:</span> {keyScheme}
                                </p>
                            </div>
                        </div>

                        {/* Action */}
                        <motion.button
                            onClick={handleClose}
                            className="w-full px-4 py-2.5 bg-accent/10 hover:bg-accent/20 border border-accent/30 text-sm font-mono text-accent rounded-sm transition-all"
                            variants={buttonVariants}
                            whileHover="hover"
                            whileTap="tap"
                        >
                            I've Saved My Recovery Information
                        </motion.button>
                    </>
                )}
            </div>
        </div>
    );
};

