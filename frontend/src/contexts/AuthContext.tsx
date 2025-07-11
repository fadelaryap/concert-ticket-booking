
import React, { createContext, useState, useContext, useEffect, useRef, useCallback } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { getProfile, login, GetProfileResponse, logout as apiLogout } from '../api/auth'; 
import { useNavigate, useLocation } from 'react-router-dom';
import axios from 'axios';

type User = GetProfileResponse;

interface AuthContextType {
  user: User | null;
  isLoadingAuth: boolean;
  isLoggedIn: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  refetchProfile: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: React.ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const location = useLocation();

  const hasAttemptedLogoutOnInvalidSession = useRef(false);

  const {
    data: user,
    isLoading: isProfileLoading,
    refetch,
    isError: isProfileError,
    error: profileError,
    isFetched,
    isFetching,
  } = useQuery<User, Error>({
    queryKey: ['userProfile'],
    queryFn: getProfile,
    staleTime: 1000 * 60 * 5,
    retry: false,
    enabled: true, 
  });

  const performLogout = useCallback(async (): Promise<void> => { 
    if (!hasAttemptedLogoutOnInvalidSession.current) {
      console.log("AuthContext: Performing logout actions.");

      try {
        await apiLogout(); 
        console.log("AuthContext: Backend logout successful.");
      } catch (err) {
        console.error("AuthContext: Failed to call backend logout API:", err);

      }

      queryClient.setQueryData(['userProfile'], null);
      queryClient.removeQueries({ queryKey: ['userProfile'] });
      queryClient.invalidateQueries();

      hasAttemptedLogoutOnInvalidSession.current = true; 

      if (location.pathname !== '/auth') {
        console.log("AuthContext: Navigating to /auth.");
        navigate('/auth', { replace: true });
      } else {
        console.log("AuthContext: Already on /auth path, preventing redundant navigation.");
      }
    } else {
      console.log("AuthContext: Logout already attempted for this invalid session, skipping.");
    }
  }, [queryClient, navigate, location.pathname]);


  useEffect(() => {
    if (isFetched && isProfileError) {
      console.warn("AuthContext: Profile fetch failed (e.g., 401 Unauthorized or network issue).");
      performLogout();
    }
  }, [isFetched, isProfileError, profileError, performLogout]);

  const performLogin = async (username: string, password: string): Promise<void> => {
    try {
      await login({ username, password });
      hasAttemptedLogoutOnInvalidSession.current = false; 
      
      const { data: fetchedUser } = await refetch(); 
      if (fetchedUser) {
        navigate('/dashboard');
      } else {
        throw new Error('Login successful, but failed to retrieve user profile.');
      }
    } catch (err) {
      if (axios.isAxiosError(err)) {
        console.error("Login failed:", err.response?.data?.error || err.message);
        throw new Error(err.response?.data?.error || 'Login failed');
      } else if (err instanceof Error) {
        console.error("Login failed:", err.message);
        throw new Error(err.message);
      } else {
        console.error("An unexpected error occurred during login.");
        throw new Error("An unexpected error occurred during login.");
      }
    }
  };

  const refetchProfile = async (): Promise<void> => {
    await refetch();
  };

  const isLoadingAuth = isFetching && !isFetched; 
  const isLoggedIn = !!user && !isProfileError;

  const value: AuthContextType = {
    user: user || null,
    isLoadingAuth: isLoadingAuth,
    isLoggedIn: isLoggedIn,
    login: performLogin,
    logout: performLogout,
    refetchProfile,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};