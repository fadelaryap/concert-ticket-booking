
import React from 'react';
import { Link } from 'react-router-dom';
import { Button } from '../components/ui/button'; 

const NotFoundPage: React.FC = () => {
  return (
    <div className="flex flex-col items-center justify-center min-h-[calc(100vh-80px)] bg-gray-50 p-4 text-center">
      <h1 className="text-6xl font-bold text-gray-800 mb-4">404</h1>
      <p className="text-xl text-gray-600 mb-8">Page Not Found</p>
      <p className="text-gray-500 mb-8">
        Sorry, the page you are looking for does not exist.
      </p>
      <Link to="/dashboard">
        <Button className="bg-blue-600 hover:bg-blue-700 text-white">
          Go to Dashboard
        </Button>
      </Link>
    </div>
  );
};

export default NotFoundPage;