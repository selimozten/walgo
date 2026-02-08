import { SiteType } from '../types';

// Animation variants for Framer Motion
export const containerVariants = {
    hidden: { opacity: 0 },
    show: {
        opacity: 1,
        transition: {
            staggerChildren: 0.1
        }
    }
};

export const itemVariants = {
    hidden: { opacity: 0, y: 20 },
    show: { opacity: 1, y: 0 }
};

// Button animation variants
export const buttonVariants = {
    hover: { 
        scale: 1.02,
        transition: { duration: 0.2 }
    },
    tap: { 
        scale: 0.98,
        transition: { duration: 0.1 }
    }
};

// Icon button animation variants (for smaller buttons)
export const iconButtonVariants = {
    hover: { 
        scale: 1.1,
        transition: { duration: 0.2 }
    },
    tap: { 
        scale: 0.9,
        transition: { duration: 0.1 }
    }
};

// Site types for AI creation
export const SITE_TYPES: SiteType[] = [
    {
        value: 'blog',
        label: 'Blog',
        description: 'Personal or professional blog with posts'
    },
    {
        value: 'documentation',
        label: 'Documentation',
        description: 'Technical or product documentation'
    },
    {
        value: 'landing-page',
        label: 'Landing Page',
        description: 'Marketing or product landing page'
    },
    {
        value: 'wiki',
        label: 'Wiki',
        description: 'Knowledge base or wiki site'
    }
];

// AI providers
export const AI_PROVIDERS = [
    { value: 'openai', label: 'OpenAI' },
    { value: 'openrouter', label: 'OpenRouter' }
];

// Networks
export const NETWORKS = [
    { value: 'mainnet', label: 'Mainnet' },
    { value: 'testnet', label: 'Testnet' }
];

// Key schemes for import
export const KEY_SCHEMES = [
    { value: 'ed25519', label: 'ed25519' },
    { value: 'secp256k1', label: 'secp256k1' }
];

// Import methods
export const IMPORT_METHODS = [
    { value: 'mnemonic', label: 'Mnemonic Phrase' },
    { value: 'privateKey', label: 'Private Key' },
    { value: 'seed', label: 'Seed Phrase' }
];

// Pagination
export const PROJECTS_PER_PAGE = 6;

// System health check interval (ms)
export const HEALTH_CHECK_INTERVAL = 30000;

// Site URLs for installation guides
export const INSTALL_GUIDE_URLS = {
    sui: 'https://docs.sui.io/guides/developer/getting-started/sui-install',
    walrus: 'https://docs.walrus.site/walrus-sites/tutorial-install.html',
    hugo: 'https://gohugo.io/installation/'
};

