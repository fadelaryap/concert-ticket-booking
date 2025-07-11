import React, { lazy, Suspense } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import queryClient from './queryClient';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import Header from './components/layout/Header';
import LoadingSpinner from './components/common/LoadingSpinner';
import Footer from './components/layout/Footer';
import DummyQrPage from './pages/DummyQrPage'; 


const AuthPage = lazy(() => import('./pages/AuthPage'));
const DashboardPage = lazy(() => import('./pages/DashboardPage'));
const ConcertDetailPage = lazy(() => import('./pages/ConcertDetailPage'));
const MyBookingsPage = lazy(() => import('./pages/MyBookingsPage'));
const BookingDetailPage = lazy(() => import('./pages/BookingDetailPage'));
const UserProfilePage = lazy(() => import('./pages/UserProfilePage'));
const BookingFormPage = lazy(() => import('./pages/BookingFormPage'));
const PaymentPage = lazy(() => import('./pages/PaymentPage'));
const NotFoundPage = lazy(() => import('./pages/NotFoundPage'));


const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isLoggedIn, isLoadingAuth } = useAuth();
  console.log("ProtectedRoute Render:", { isLoggedIn, isLoadingAuth }); 
  

  if (isLoadingAuth) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner />
      </div>
    );
  }


  if (!isLoggedIn) {
    console.log("ProtectedRoute: Not logged in. Redirecting to /auth.");
    return <Navigate to="/auth" replace />; 
  }

  return children;
};

function AppRoutes() {
  return (
    <Routes>
      <Route path="/auth" element={<AuthPage />} />
      <Route path="/" element={<Navigate to="/dashboard" replace />} />
      {}
      <Route
        path="/dashboard"
        element={
          <ProtectedRoute>
            <DashboardPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/concerts/:concertId"
        element={
          <ProtectedRoute>
            <ConcertDetailPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/bookings/confirm/:concertId"
        element={
          <ProtectedRoute>
            <BookingFormPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/payment/:bookingId"
        element={
          <ProtectedRoute>
            <PaymentPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/my-bookings"
        element={
          <ProtectedRoute>
            <MyBookingsPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/bookings/:bookingId"
        element={
          <ProtectedRoute>
            <BookingDetailPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/profile"
        element={
          <ProtectedRoute>
            <UserProfilePage />
          </ProtectedRoute>
        }
      />
      {}
      <Route
        path="/qr-payment/:bookingId"
        element={
          <ProtectedRoute>
            <DummyQrPage />
          </ProtectedRoute>
        }
      />
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  );
}

function App() {
  const HEADER_HEIGHT_PX = 64; 
  const FOOTER_HEIGHT_PX = 56; 
  const mainMinHeight = `calc(100vh - ${HEADER_HEIGHT_PX}px - ${FOOTER_HEIGHT_PX}px)`;

  return (
    <QueryClientProvider client={queryClient}>
      <Router>
        <AuthProvider>
          <Header />
          {}
          <main className="flex-grow w-full px-4 py-8" style={{ minHeight: mainMinHeight }}>
            <Suspense fallback={<LoadingSpinner />}>
              <AppRoutes />
            </Suspense>
          </main>
          <Footer />
        </AuthProvider>
      </Router>
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
}

export default App;