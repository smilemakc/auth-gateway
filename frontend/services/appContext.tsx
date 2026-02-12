import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from './apiClient';
import { queryKeys } from './queryClient';
import type { Application, ListApplicationsResponse } from '../types';

const STORAGE_KEY = 'auth_gateway_current_app_id';

interface ApplicationContextType {
  currentApplicationId: string | null;
  currentApplication: Application | null;
  applications: Application[];
  isLoading: boolean;
  setCurrentApplicationId: (id: string | null) => void;
}

const ApplicationContext = createContext<ApplicationContextType | undefined>(undefined);

export const ApplicationProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const queryClient = useQueryClient();
  const [currentApplicationId, setCurrentAppId] = useState<string | null>(() => {
    return localStorage.getItem(STORAGE_KEY);
  });

  const API_BASE = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api`;

  const { data, isLoading } = useQuery<ListApplicationsResponse>({
    queryKey: queryKeys.applications.list(1, 100),
    queryFn: async () => {
      const token = localStorage.getItem('auth_gateway_access_token');
      const response = await fetch(`${API_BASE}/admin/applications?page=1&per_page=100`, {
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
      });
      if (!response.ok) {
        throw new Error('Failed to fetch applications');
      }
      return response.json();
    },
  });

  const applications = data?.applications ?? [];

  const currentApplication = currentApplicationId
    ? applications.find((app) => app.id === currentApplicationId) ?? null
    : null;

  const setCurrentApplicationId = useCallback(
    (id: string | null) => {
      setCurrentAppId(id);

      if (id) {
        localStorage.setItem(STORAGE_KEY, id);
        apiClient.setHeader('X-Application-ID', id);
      } else {
        localStorage.removeItem(STORAGE_KEY);
        apiClient.removeHeader('X-Application-ID');
      }

      // Invalidate queries that are scoped to application
      queryClient.invalidateQueries();
    },
    [queryClient],
  );

  // Set header on mount if there's a stored app ID
  useEffect(() => {
    if (currentApplicationId) {
      apiClient.setHeader('X-Application-ID', currentApplicationId);
    }
  }, []);

  return (
    <ApplicationContext.Provider
      value={{
        currentApplicationId,
        currentApplication,
        applications,
        isLoading,
        setCurrentApplicationId,
      }}
    >
      {children}
    </ApplicationContext.Provider>
  );
};

export const useApplication = () => {
  const context = useContext(ApplicationContext);
  if (!context) {
    throw new Error('useApplication must be used within ApplicationProvider');
  }
  return context;
};
