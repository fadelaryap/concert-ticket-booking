
import React, { useState, useEffect } from 'react';
import LoginForm from '../components/forms/LoginForm';
import RegisterForm from '../components/forms/RegisterForm';
import { useAuth } from '../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
import Alert from '../components/common/Alert';
import { register } from '../api/auth';
import LoadingSpinner from '../components/common/LoadingSpinner';

function AuthPage() {
  const [isLoginMode, setIsLoginMode] = useState(true);
  const { login: authLogin, isLoggedIn, isLoadingAuth } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    console.log("AuthPage useEffect:", { isLoggedIn, isLoadingAuth });

    if (!isLoadingAuth && isLoggedIn) {
        console.log("AuthPage: Logged in, redirecting to /dashboard.");
        navigate('/dashboard', { replace: true }); 
    }
  }, [isLoggedIn, navigate, isLoadingAuth]);

  const [loginError, setLoginError] = useState<string | null>(null);
  const [registerError, setRegisterError] = useState<string | null>(null);
  const [registerSuccess, setRegisterSuccess] = useState<string | null>(null);

  const handleLoginSubmit = async (values: any) => {
    setLoginError(null);
    try {
      await authLogin(values.username, values.password);
    } catch (err: any) {
      setLoginError(err.message || 'Login failed. Please check your credentials.');
    }
  };

  const handleRegisterSubmit = async (values: any) => {
    setRegisterError(null);
    setRegisterSuccess(null);
    try {
      const response = await register(values);
      if (!response || (response as any).error) {
        throw new Error((response as any).error || 'Registration failed');
      }
      setRegisterSuccess('Registration successful! Please log in.');
      setIsLoginMode(true);
    } catch (err: any) {
      setRegisterError(err.message || 'Registration failed.');
    }
  };

  if (isLoadingAuth) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner />
      </div>
    );
  }

  if (isLoggedIn) {
    return null;
  }

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-100">
      <div className="p-8 bg-white rounded-lg shadow-md w-full max-w-md"> {}
        <h1 className="text-3xl font-bold mb-6 text-center text-gray-800">{isLoginMode ? 'Login' : 'Register'}</h1>
        {loginError && <Alert type="error" description={loginError} />}
        {registerError && <Alert type="error" description={registerError} />}
        {registerSuccess && <Alert type="success" description={registerSuccess} />}

        {isLoginMode ? (
          <>
            <LoginForm onSubmit={handleLoginSubmit} isLoading={false} />
            <p className="text-center mt-4 text-sm text-gray-600">
              Don't have an account?{' '}
              <span className="text-blue-600 hover:text-blue-800 cursor-pointer" onClick={() => setIsLoginMode(false)}>
                Register here
              </span>
            </p>
          </>
        ) : (
          <>
            <RegisterForm onSubmit={handleRegisterSubmit} isLoading={false} />
            <p className="text-center mt-4 text-sm text-gray-600">
              Already have an account?{' '}
              <span className="text-blue-600 hover:text-blue-800 cursor-pointer" onClick={() => setIsLoginMode(true)}>
                Login here
              </span>
            </p>
          </>
        )}
      </div>
    </div>
  );
}

export default AuthPage;