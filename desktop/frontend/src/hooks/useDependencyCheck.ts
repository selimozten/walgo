import { useMemo } from 'react';
import { SystemHealth } from '../types';

interface DependencyCheckResult {
  canLaunch: boolean;
  canServe: boolean;
  canCreateSite: boolean;
  missingDeps: string[];
  needsUpdate: boolean;
  updateMessage?: string;
}

interface UseDependencyCheckProps {
  systemHealth?: SystemHealth;
  hasUpdates?: boolean;
  updatingTools?: string[];
}

export const useDependencyCheck = ({
  systemHealth,
  hasUpdates = false,
  updatingTools = []
}: UseDependencyCheckProps): DependencyCheckResult => {
  
  return useMemo(() => {
    const missingDeps: string[] = [];
    
    // Check each dependency (suiup tools only)
    if (!systemHealth?.suiInstalled) {
      missingDeps.push('Sui CLI');
    }
    if (!systemHealth?.walrusInstalled) {
      missingDeps.push('Walrus CLI');
    }
    if (!systemHealth?.siteBuilder) {
      missingDeps.push('Site Builder');
    }
    // Hugo removed - optional, users install via package manager
    
    // Check if network is online
    const isOnline = systemHealth?.netOnline ?? false;
    
    // All dependencies installed
    const allDepsInstalled = missingDeps.length === 0;
    
    // Check if any tools are being updated
    const isUpdating = updatingTools.length > 0;
    
    // Can launch: all suiup deps + online + no updates needed or updating
    const canLaunch = allDepsInstalled && isOnline && !hasUpdates && !isUpdating;
    
    // Can serve: needs Hugo (for local preview)
    const canServe = systemHealth?.hugoInstalled ?? false;
    
    // Can create site: needs Hugo
    const canCreateSite = systemHealth?.hugoInstalled ?? false;
    
    // Build update message
    let updateMessage: string | undefined;
    if (hasUpdates && !isUpdating) {
      updateMessage = 'Please update tools to the latest version before launching';
    } else if (isUpdating) {
      updateMessage = `Updating ${updatingTools.join(', ')}...`;
    }
    
    return {
      canLaunch,
      canServe,
      canCreateSite,
      missingDeps,
      needsUpdate: hasUpdates || isUpdating,
      updateMessage
    };
  }, [systemHealth, hasUpdates, updatingTools]);
};

