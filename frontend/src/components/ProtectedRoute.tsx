
import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import LoadingSpinner from './common/LoadingSpinner';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children }) => {
  const { isLoggedIn, isLoadingAuth } = useAuth();
  const location = useLocation();

  console.log("ProtectedRoute:", { 
    isLoggedIn, 
    isLoadingAuth, 
    pathname: location.pathname 
  });


  if (isLoadingAuth) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner />
      </div>
    );
  }


  if (!isLoggedIn) {
    console.log("ProtectedRoute: Not logged in. Redirecting to /auth.");
    return <Navigate to="/auth" state={{ from: location }} replace />;
  }

  return <>{children}</>;
};

export default ProtectedRoute;