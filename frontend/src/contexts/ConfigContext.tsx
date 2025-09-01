import React, { createContext, useContext, useState, ReactNode } from 'react';

export interface Configuration {
  // UI Style
  theme: 'modern' | 'colorful' | 'dark' | 'light';
  colorScheme: 'blue' | 'green' | 'purple' | 'custom';
  layout: 'compact' | 'spacious' | 'sidebar' | 'top';
  
  // Features
  realTimeTesting: boolean;
  moderationDashboard: boolean;
  contentForms: boolean;
  statistics: boolean;
  userManagement: boolean;
  
  // Authentication
  demoMode: boolean;
  loginForm: boolean;
  userRegistration: boolean;
  
  // Content Testing
  postTesting: boolean;
  commentTesting: boolean;
  profileTesting: boolean;
  customContent: boolean;
  
  // Deployment
  port: number;
  apiEndpoint: string;
  proxySettings: boolean;
  
  // Custom Colors
  customColors: {
    primary: string;
    secondary: string;
    accent: string;
  };
}

const defaultConfig: Configuration = {
  theme: 'modern',
  colorScheme: 'blue',
  layout: 'top',
  realTimeTesting: true,
  moderationDashboard: true,
  contentForms: true,
  statistics: true,
  userManagement: false,
  demoMode: true,
  loginForm: false,
  userRegistration: false,
  postTesting: true,
  commentTesting: true,
  profileTesting: true,
  customContent: false,
  port: 3000,
  apiEndpoint: 'http://localhost:8080',
  proxySettings: true,
  customColors: {
    primary: '#1976d2',
    secondary: '#dc004e',
    accent: '#ff9800',
  },
};

interface ConfigContextType {
  config: Configuration;
  updateConfig: (updates: Partial<Configuration>) => void;
  resetConfig: () => void;
  saveConfig: () => void;
  loadConfig: () => void;
}

const ConfigContext = createContext<ConfigContextType | undefined>(undefined);

export const useConfig = () => {
  const context = useContext(ConfigContext);
  if (!context) {
    throw new Error('useConfig must be used within a ConfigProvider');
  }
  return context;
};

interface ConfigProviderProps {
  children: ReactNode;
}

export const ConfigProvider: React.FC<ConfigProviderProps> = ({ children }) => {
  const [config, setConfig] = useState<Configuration>(() => {
    const saved = localStorage.getItem('moderation-config');
    return saved ? { ...defaultConfig, ...JSON.parse(saved) } : defaultConfig;
  });

  const updateConfig = (updates: Partial<Configuration>) => {
    setConfig(prev => {
      const newConfig = { ...prev, ...updates };
      localStorage.setItem('moderation-config', JSON.stringify(newConfig));
      return newConfig;
    });
  };

  const resetConfig = () => {
    setConfig(defaultConfig);
    localStorage.removeItem('moderation-config');
  };

  const saveConfig = () => {
    localStorage.setItem('moderation-config', JSON.stringify(config));
  };

  const loadConfig = () => {
    const saved = localStorage.getItem('moderation-config');
    if (saved) {
      setConfig({ ...defaultConfig, ...JSON.parse(saved) });
    }
  };

  return (
    <ConfigContext.Provider value={{
      config,
      updateConfig,
      resetConfig,
      saveConfig,
      loadConfig,
    }}>
      {children}
    </ConfigContext.Provider>
  );
};

