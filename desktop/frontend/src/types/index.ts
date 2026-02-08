// Global type definitions for Walgo application

export type StatusType = 'success' | 'error' | 'info';

export interface Status {
    type: StatusType;
    message: string;
}

export interface DeploymentHistory {
    timestamp: string;
    objectId: string;
    network: string;
    epochs?: number;
    size?: number;
    status: 'success' | 'failed';
    wallet?: string;
    gasFee?: string; // Actual gas cost (e.g., "0.001234 WAL + 0.061646 SUI")
}

export interface Project {
    id?: number;
    name: string;
    path?: string; // Alias for sitePath for backward compatibility
    sitePath?: string; // From API
    category?: string;
    description?: string;
    imageUrl?: string;
    suins?: string;
    status?: 'draft' | 'active' | 'archived' | string;
    network?: 'mainnet' | 'testnet' | string;
    objectId?: string;
    epochs?: number; // Last deployment epochs
    totalEpochs?: number; // Sum of all deployment epochs (for accurate expiry)
    expiresAt?: string; // Calculated expiry date (RFC3339)
    expiresIn?: string; // Human-readable expiry duration (e.g., "2 weeks, 3 days")
    gasFee?: string; // Actual gas cost from last deployment (e.g., "0.001234 WAL + 0.061646 SUI")
    deployments?: number;
    lastDeploy?: string;
    lastDeployAt?: string; // From API
    firstDeployAt?: string; // First deployment date (for accurate expiry)
    updatedAt?: string;
    createdAt?: string;
    lastModified?: string;
    url?: string;
    deployedAt?: string;
    size?: number;
    fileCount?: number;
    wallet?: string;
    deploymentHistory?: DeploymentHistory[];
}

export interface AIConfig {
    success: boolean;
    enabled: boolean;
    currentProvider?: string;
    currentModel?: string;
    providers?: Record<string, {
        apiKey?: string;
        baseURL?: string;
        model?: string;
    }>;
}

export interface WalletInfo {
    address: string;
    suiBalance: number;
    walBalance: number;
    network: string;
    active: boolean;
}

export interface SystemHealth {
    netOnline: boolean;
    suiInstalled: boolean;
    suiConfigured: boolean;
    walrusInstalled: boolean;
    siteBuilder: boolean;
    hugoInstalled: boolean;
    message?: string;
}

export interface FileTreeNode {
    name: string;
    path: string;
    isDir: boolean;
    size?: number;
    modified?: number;
    children?: FileTreeNode[];
}

export interface LaunchConfig {
    projectName: string;
    category: string;
    description: string;
    imageUrl?: string;
}

export interface ImportMethod {
    type: 'mnemonic' | 'privateKey' | 'seed';
    keyScheme: 'ed25519' | 'secp256k1';
}

export interface InstallStatus {
    installing: boolean;
    message?: string;
    progress?: number;
}

export interface SiteType {
    value: string;
    label: string;
    description: string;
}


