import { useState, useEffect } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import {
  LayoutGrid,
  Plus,
  Edit3,
  Database,
  Sparkles,
  Activity,
  MessageSquare,
  LucideIcon,
} from 'lucide-react';
import { BrowserOpenURL } from '../wailsjs/runtime/runtime';
import {
  Dashboard,
  AIConfig,
  Projects,
  Create,
  QuickStart,
  InitSite,
  Import,
  NewContent,
  AIGenerate,
  AICreateSite,
  Edit,
  SystemHealth,
} from './pages';
import { ImportAccountModal, CreateAccountModal } from './components/modals';
import { WindowControls, StatusBanner, LoadingOverlay } from './components/ui';
import { AIProgressModal, AIProgressIndicator } from './components';
import { AIProgressProvider } from './contexts';
import { containerVariants } from './utils/constants';
import { useProjects, useAIConfig, useWallet, useSystemHealth } from './hooks';
import {
  SwitchNetwork,
  ListProjects,
}
from '../wailsjs/go/main/App';
import WalgoLogo from './assets/walgo-Wlogo-no_background.svg';
import { cn } from './utils/helpers';

// NavItem component
const NavItem = ({
  id,
  icon: Icon,
  label,
  activeTab,
  setActiveTab,
  badge,
}: {
  id: string;
  icon: LucideIcon;
  label: string;
  activeTab: string;
  setActiveTab: (tab: string) => void;
  badge?: number;
}) => (
  <button
    onClick={() => setActiveTab(id)}
    className={cn(
      'group relative flex items-center gap-3 w-full px-4 py-3 rounded-sm transition-all',
      activeTab === id
        ? 'bg-accent/10 text-accent'
        : 'text-zinc-500 hover:text-zinc-300 hover:bg-white/5'
    )}
  >
    <Icon size={18} strokeWidth={1.5} />
    <span className="text-sm font-mono uppercase tracking-wider">{label}</span>
    {badge !== undefined && badge > 0 && (
      <span className="ml-auto px-2 py-0.5 bg-accent/20 text-accent text-[10px] font-mono rounded-full">
        {badge}
      </span>
    )}
  </button>
);

function App() {
  // Platform detection
  const [isMac, setIsMac] = useState(false);

  // Navigation
  const [activeTab, setActiveTab] = useState('dashboard');

  // Status
  const [status, setStatus] = useState<{
    type: 'success' | 'error' | 'info';
    message: string;
  } | null>(null);

  // Processing state
  const [isProcessing, setIsProcessing] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [showCreateModal, setShowCreateModal] = useState(false);

  // Hooks
  const { projects, reloadProjects } = useProjects();
  const { configured: aiConfigured, loadConfig: reloadAIConfig } = useAIConfig();
  const { walletInfo, addressList, reload: reloadWallet, switchAddress: switchAddressHook, createAddress, importAddress } = useWallet();
  const { health, version, checkDeps, reload: reloadHealth } = useSystemHealth();

  // Detect platform
  useEffect(() => {
    const userAgent = navigator.userAgent.toLowerCase();
    setIsMac(userAgent.indexOf('mac') >= 0);
  }, []);

  // Load initial data
  useEffect(() => {
    reloadProjects();
    reloadWallet();
  }, []);

  // Status handler
  const handleStatusChange = (newStatus: {
    type: 'success' | 'error' | 'info';
    message: string;
  }) => {
    setStatus(newStatus);
    setTimeout(() => setStatus(null), 5000); // Increased from 2000ms to 5000ms
  };

  // Navigation handler
  const handleNavigate = (tab: string) => {
    setActiveTab(tab);
  };

  // Project creation success - navigate to edit page with the new project
  const handleProjectCreated = async (createdSitePath: string) => {
    // Reload projects in background
    reloadProjects();

    // Fetch fresh project list to find the newly created project
    try {
      const freshProjects = await ListProjects();
      const newProject = freshProjects?.find(
        (p) => p.sitePath === createdSitePath
      );

      if (newProject) {
        // Save project to localStorage for Edit page to load
        localStorage.setItem('selectedProject', JSON.stringify(newProject));
        // Navigate to edit page
        setActiveTab('edit');
      } else {
        // Fallback to projects page if project not found
        console.warn('Project not found in list, falling back to projects page');
        setActiveTab('projects');
      }
    } catch (err) {
      console.error('Failed to fetch projects:', err);
      // Fallback to projects page on error
      setActiveTab('projects');
    }

    handleStatusChange({
      type: 'success',
      message: 'Project created',
    });
  };


  // Network handler
  const handleSwitchNetwork = async (network: string) => {
    setIsProcessing(true);
    try {
      const result = await SwitchNetwork(network);
      if (result.success) {
        reloadWallet();
        handleStatusChange({
          type: 'success',
          message: 'Network switched',
        });
      } else {
        handleStatusChange({
          type: 'error',
          message: 'Network switch failed',
        });
      }
    } finally {
      setIsProcessing(false);
    }
  };

  // Wallet handlers
  const handleSwitchAccount = async (address: string) => {
    setIsProcessing(true);
    try {
      const result = await switchAddressHook(address);
      if (result.success) {
        reloadWallet();
        handleStatusChange({
          type: 'success',
          message: 'Account switched',
        });
      } else {
        handleStatusChange({
          type: 'error',
          message: `Switch failed: ${result.error}`,
        });
      }
    } finally {
      setIsProcessing(false);
    }
  };

  const handleCreateAccount = () => {
    setShowCreateModal(true);
  };

  const handleCreateAccountSubmit = async (keyScheme: string, alias?: string) => {
    const result = await createAddress(keyScheme, alias);
    if (result.success) {
      // Auto switch to the newly created account
      if (result.address) {
        const switchResult = await switchAddressHook(result.address);
        if (switchResult.success) {
          reloadWallet();
        }
      }
      handleStatusChange({
        type: 'success',
        message: 'Account created and switched',
      });
      // Return the result with mnemonic (recoveryPhrase) for display
      return {
        success: true,
        address: result.address,
        mnemonic: result.result?.recoveryPhrase,
      };
    } else {
      handleStatusChange({
        type: 'error',
        message: `Create failed: ${result.error}`,
      });
      return { success: false, error: result.error };
    }
  };

  const handleImportAccount = () => {
    setShowImportModal(true);
  };

  const handleImportAccountSubmit = async (method: string, input: string, keyScheme?: string) => {
    const result = await importAddress(method, input, keyScheme);
    if (result.success) {
      // Auto switch to the newly imported account
      if (result.address) {
        const switchResult = await switchAddressHook(result.address);
        if (switchResult.success) {
          reloadWallet();
        }
      }
      handleStatusChange({
        type: 'success',
        message: 'Account imported and switched',
      });
      setShowImportModal(false);
    } else {
      handleStatusChange({
        type: 'error',
        message: `Import failed: ${result.error}`,
      });
    }
  };

  // Render current page
  const renderPage = () => {
    switch (activeTab) {
      case 'dashboard':
        return (
          <Dashboard
            version={version}
            walletInfo={walletInfo}
            addressList={addressList}
            aiConfigured={aiConfigured}
            systemHealth={health}
            onNavigate={handleNavigate}
            onRefreshHealth={reloadHealth}
            onSwitchNetwork={handleSwitchNetwork}
            onSwitchAccount={handleSwitchAccount}
            onCreateAccount={handleCreateAccount}
            onImportAccount={handleImportAccount}
            onStatusChange={handleStatusChange}
          />
        );

      case 'create':
        return <Create onNavigate={handleNavigate} aiConfigured={aiConfigured} systemHealth={health} onRefreshHealth={reloadHealth} onStatusChange={handleStatusChange} />;

      case 'quickstart':
        return (
          <QuickStart
            onSuccess={handleProjectCreated}
            onStatusChange={handleStatusChange}
          />
        );

      case 'create-site':
        return (
          <InitSite
            onSuccess={handleProjectCreated}
            onStatusChange={handleStatusChange}
          />
        );

      case 'edit':
        return (
          <Edit
            aiConfigured={aiConfigured}
            systemHealth={health}
            onStatusChange={handleStatusChange}
            onProjectUpdate={reloadProjects}
            onRefreshHealth={reloadHealth}
          />
        );

      case 'projects':
        return (
          <Projects
            projects={projects}
            loading={false}
            onStatusChange={handleStatusChange}
            onRefresh={reloadProjects}
            onNavigateToEdit={() => setActiveTab('edit')}
          />
        );

      case 'ai':
        return <AIConfig onConfigChange={reloadAIConfig} />;

      case 'new-content':
        return (
          <NewContent
            onSuccess={() => handleNavigate('edit')}
            onStatusChange={handleStatusChange}
          />
        );

      case 'ai-generate':
        return (
          <AIGenerate
            onSuccess={() => handleNavigate('edit')}
            onStatusChange={handleStatusChange}
            onNavigateToAI={() => handleNavigate('ai')}
          />
        );

      case 'import':
        return (
          <Import
            onSuccess={handleProjectCreated}
            onStatusChange={handleStatusChange}
          />
        );

      case 'ai-create-site':
        return (
          <AICreateSite
            onSuccess={handleProjectCreated}
            onStatusChange={handleStatusChange}
            onNavigateToAI={() => handleNavigate('ai')}
          />
        );

      case 'system-health':
        return (
          <SystemHealth
            systemHealth={health}
            onCheckDeps={async () => {
              const result = await checkDeps();
              if (result.success) {
                handleStatusChange({
                  type: 'success',
                  message: 'Dependencies installed',
                });
              } else {
                const errorMsg = 'message' in result ? result.message : ('error' in result ? result.error : 'Unknown error');
                handleStatusChange({
                  type: 'error',
                  message: `Install failed: ${errorMsg}`,
                });
              }
            }}
            onRefresh={async () => {
              setIsProcessing(true);
              await reloadHealth();
              setIsProcessing(false);
              handleStatusChange({
                type: 'info',
                message: 'Refreshed',
              });
            }}
          />
        );

      default:
        return (
          <Dashboard
            version={version}
            walletInfo={walletInfo}
            addressList={addressList}
            aiConfigured={aiConfigured}
            systemHealth={health}
            onNavigate={handleNavigate}
            onRefreshHealth={reloadHealth}
            onSwitchNetwork={handleSwitchNetwork}
            onSwitchAccount={handleSwitchAccount}
            onCreateAccount={handleCreateAccount}
            onImportAccount={handleImportAccount}
            onStatusChange={handleStatusChange}
          />
        );
    }
  };

  return (
    <AIProgressProvider>
    <div className="h-screen flex font-sans overflow-hidden relative bg-black">
      {/* Background noise texture */}
      <div className="bg-noise" />

      {/* Global Loading Overlay */}
      {isProcessing && <LoadingOverlay message="Processing..." fullScreen />}

      {/* Window drag region - full width for dragging */}
      <div className="fixed top-0 left-0 right-0 h-10 z-50 wails-drag" />

      {/* Window Controls */}
      <WindowControls />

      {/* Navigation Sidebar */}
      <nav className="w-64 border-r border-white/10 bg-black/40 backdrop-blur-sm flex flex-col relative z-10">
        {/* macOS: Add padding for native traffic lights, Windows/Linux: normal padding */}
        <div className={cn(
          "p-6 border-b border-white/5 wails-drag",
          isMac && "pt-12"
        )}>
          <button
            onClick={() => setActiveTab('dashboard')}
            className="group w-full flex flex-col items-center gap-3 transition-all duration-300 hover:scale-[1.02] active:scale-[0.98] cursor-pointer wails-nodrag"
            aria-label="Go to Dashboard"
          >
            {/* Logo with gradient background */}
            <div className="relative">
              <div className="absolute inset-0 bg-gradient-to-br from-[#4DA2FF]/20 to-[#2563eb]/20 rounded-2xl blur-xl group-hover:blur-2xl transition-all duration-300" />
              <div className="relative bg-gradient-to-br from-[#4DA2FF]/10 to-[#2563eb]/10 p-4 rounded-2xl border border-white/10 group-hover:border-accent/40 transition-all duration-300">
                <img 
                  src={WalgoLogo} 
                  alt="Walgo" 
                  className="w-16 h-16 transition-all duration-300 group-hover:drop-shadow-[0_0_15px_rgba(77,162,255,0.6)]" 
                />
              </div>
            </div>
            
            {/* Brand name */}
            <div className="flex flex-col items-center gap-1">
              <h1 className="text-2xl font-display font-bold text-white tracking-wider group-hover:text-accent transition-colors duration-300">
                WALGO
              </h1>
              <span className="px-2.5 py-0.5 text-[9px] font-bold bg-gradient-to-r from-yellow-500/20 to-orange-500/20 text-yellow-400 border border-yellow-500/30 rounded-full uppercase tracking-widest shadow-lg shadow-yellow-500/10 animate-pulse">
                Beta
              </span>
              <div className="h-px w-16 bg-gradient-to-r from-transparent via-accent/50 to-transparent group-hover:via-accent transition-all duration-300 mt-1" />
            </div>
          </button>
        </div>

        <div className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
          <NavItem
            id="dashboard"
            icon={LayoutGrid}
            label="Dashboard"
            activeTab={activeTab}
            setActiveTab={setActiveTab}
          />
          <NavItem
            id="create"
            icon={Plus}
            label="Create"
            activeTab={activeTab}
            setActiveTab={setActiveTab}
          />
          <NavItem
            id="edit"
            icon={Edit3}
            label="Edit"
            activeTab={activeTab}
            setActiveTab={setActiveTab}
          />
          <NavItem
            id="projects"
            icon={Database}
            label="Projects"
            activeTab={activeTab}
            setActiveTab={setActiveTab}
            badge={projects.length}
          />
          <NavItem
            id="ai"
            icon={Sparkles}
            label="AI Config"
            activeTab={activeTab}
            setActiveTab={setActiveTab}
          />
          <NavItem
            id="system-health"
            icon={Activity}
            label="System Health"
            activeTab={activeTab}
            setActiveTab={setActiveTab}
          />
        </div>

        {/* Version & Feedback - Bottom of Sidebar */}
        <div className="px-4 py-3 border-t border-white/10 bg-black/20 space-y-3">
          {/* Feedback Button */}
          <button
            onClick={() => BrowserOpenURL('https://tally.so/r/447YBO')}
            className="w-full flex items-center justify-center gap-2 px-3 py-2 bg-gradient-to-r from-accent/20 to-blue-500/20 hover:from-accent/30 hover:to-blue-500/30 border border-accent/30 hover:border-accent/50 rounded-lg transition-all duration-200 group wails-nodrag"
          >
            <MessageSquare size={14} className="text-accent group-hover:scale-110 transition-transform" />
            <span className="text-xs font-medium text-accent">Feedback</span>
          </button>

          {/* Version */}
          <div className="text-center">
            <div className="text-[10px] font-mono text-zinc-600 uppercase tracking-wider mb-1">
              Version
            </div>
            <div className="text-sm font-mono text-accent font-semibold">
              {version}
            </div>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="flex-1 flex flex-col relative z-10 overflow-hidden">
        {/* Status Banner */}
        {status && (
          <StatusBanner status={status} onClose={() => setStatus(null)} />
        )}

        {/* Page Content */}
        <div className="flex-1 overflow-y-auto p-8">
          <AnimatePresence mode="wait">
            <motion.div
              key={activeTab}
              variants={containerVariants}
              initial="hidden"
              animate="show"
              exit="hidden"
            >
              {renderPage()}
            </motion.div>
          </AnimatePresence>
        </div>
      </main>

      {/* Modals - Wallet Operations Only */}
      <ImportAccountModal
        isOpen={showImportModal}
        onClose={() => setShowImportModal(false)}
        onImport={handleImportAccountSubmit}
        isProcessing={isProcessing}
      />

      <CreateAccountModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onCreate={handleCreateAccountSubmit}
        isProcessing={isProcessing}
      />

      {/* AI Progress Components */}
      <AIProgressModal />
      <AIProgressIndicator />
    </div>
    </AIProgressProvider>
  );
}

export default App;
