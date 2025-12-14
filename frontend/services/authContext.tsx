import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';
import type { User } from '@auth-gateway/client-sdk';
import { apiClient, AUTH_FAILURE_EVENT } from './apiClient';

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Handle auth failure event from apiClient (e.g., when token refresh fails)
  const handleAuthFailure = useCallback(() => {
    console.log('[AuthContext] Auth failure event received - logging out user');
    setUser(null);
  }, []);

  // Listen to auth failure events
  useEffect(() => {
    window.addEventListener(AUTH_FAILURE_EVENT, handleAuthFailure);
    return () => {
      window.removeEventListener(AUTH_FAILURE_EVENT, handleAuthFailure);
    };
  }, [handleAuthFailure]);

  // Check authentication on mount
  useEffect(() => {
    const checkAuth = async () => {
      const token = localStorage.getItem('auth_gateway_access_token');
      if (token) {
        try {
          const profile = await apiClient.auth.getProfile();
          setUser(profile as User);
        } catch (error) {
          console.error('Failed to fetch profile:', error);
          localStorage.removeItem('auth_gateway_access_token');
          localStorage.removeItem('auth_gateway_refresh_token');
          localStorage.removeItem('auth_token');
        }
      }
      setIsLoading(false);
    };

    checkAuth();
  }, []);

  const login = async (email: string, password: string) => {
    const response = await apiClient.auth.signIn({ email, password });

    console.log('[Auth] Login response:', response);

    // Manually store tokens if SDK didn't do it
    // Check various possible response formats
    const accessToken = (response as any).accessToken || (response as any).tokens?.accessToken || (response as any).access_token;
    const refreshToken = (response as any).refreshToken || (response as any).tokens?.refreshToken || (response as any).refresh_token;

    if (accessToken) {
      console.log('[Auth] Manually storing access token');
      localStorage.setItem('auth_gateway_access_token', accessToken);
    }

    if (refreshToken) {
      console.log('[Auth] Manually storing refresh token');
      localStorage.setItem('auth_gateway_refresh_token', refreshToken);
    }

    // Set user state
    setUser(response.user as User);

    // Debug: Check if tokens are stored
    console.log('[Auth] Tokens after login:', {
      accessToken: !!localStorage.getItem('auth_gateway_access_token'),
      refreshToken: !!localStorage.getItem('auth_gateway_refresh_token'),
    });
  };

  const logout = async () => {
    try {
      await apiClient.auth.logout();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      // Clear all auth-related items
      localStorage.removeItem('auth_gateway_access_token');
      localStorage.removeItem('auth_gateway_refresh_token');
      localStorage.removeItem('auth_token');
      setUser(null);
    }
  };

  const refreshUser = async () => {
    const profile = await apiClient.auth.getProfile();
    setUser(profile as User);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        logout,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};

/**
 * Check if user has a specific role by name
 */
export const hasRole = (user: User | null, roleName: string): boolean => {
  if (!user || !user.roles) return false;
  return user.roles.some(role => role.name === roleName);
};

/**
 * Check if user has any of the specified roles
 */
export const hasAnyRole = (user: User | null, roleNames: string[]): boolean => {
  if (!user || !user.roles) return false;
  return user.roles.some(role => roleNames.includes(role.name));
};

/**
 * Check if user is admin
 */
export const isAdmin = (user: User | null): boolean => {
  return hasRole(user, 'admin');
};
