import { useState, useEffect } from 'react';
import { 
    GetWalletInfo, 
    GetAddressList, 
    SwitchAddress, 
    CreateAddress as CreateWalletAddress,
    ImportAddress,
    SwitchNetwork 
} from '../../wailsjs/go/main/App';
import { WalletInfo } from '../types';

export const useWallet = () => {
    const [walletInfo, setWalletInfo] = useState<WalletInfo | null>(null);
    const [addressList, setAddressList] = useState<string[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const loadWalletInfo = async () => {
        setLoading(true);
        setError(null);
        try {
            const result = await GetWalletInfo();
            setWalletInfo(result);
        } catch (err) {
            console.error('Failed to load wallet info:', err);
            setError('Failed to load wallet information');
        } finally {
            setLoading(false);
        }
    };

    const loadAddressList = async () => {
        setLoading(true);
        setError(null);
        try {
            const result = await GetAddressList();
            if (result && result.addresses) {
                setAddressList(result.addresses);
            }
        } catch (err) {
            console.error('Failed to load address list:', err);
            setError('Failed to load address list');
        } finally {
            setLoading(false);
        }
    };

    const switchAddress = async (address: string) => {
        setLoading(true);
        setError(null);
        try {
            const result = await SwitchAddress(address);
            await loadWalletInfo();
            return { success: true, result };
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, error: errorMsg };
        } finally {
            setLoading(false);
        }
    };

    const createAddress = async (keyScheme?: string, alias?: string) => {
        setLoading(true);
        setError(null);
        try {
            const result = await CreateWalletAddress(keyScheme || 'ed25519', alias || '');
            await Promise.all([loadWalletInfo(), loadAddressList()]);
            return { success: true, result, address: result?.address };
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, error: errorMsg };
        } finally {
            setLoading(false);
        }
    };

    const importAddress = async (method: string, input: string, keyScheme?: string) => {
        setLoading(true);
        setError(null);
        try {
            const result = await ImportAddress({
                method,
                input,
                keyScheme: keyScheme || 'ed25519',
            });
            await Promise.all([loadWalletInfo(), loadAddressList()]);
            return { success: true, result, address: result?.address };
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, error: errorMsg };
        } finally {
            setLoading(false);
        }
    };

    const switchNetwork = async (network: string) => {
        setLoading(true);
        setError(null);
        try {
            const result = await SwitchNetwork(network);
            await loadWalletInfo();
            return { success: true, result };
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : 'Unknown error';
            setError(errorMsg);
            return { success: false, error: errorMsg };
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadWalletInfo();
        loadAddressList();
    }, []);

    return {
        walletInfo,
        addressList,
        loading,
        error,
        reload: () => {
            loadWalletInfo();
            loadAddressList();
        },
        switchAddress,
        createAddress,
        importAddress,
        switchNetwork
    };
};

