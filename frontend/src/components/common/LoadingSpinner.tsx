
import React from 'react';

interface LoadingSpinnerProps {
  size?: 'small' | 'medium' | 'large';
  className?: string;
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({ size = 'medium', className = '' }) => {
  let spinnerSizeClasses = '';
  switch (size) {
    case 'small':
      spinnerSizeClasses = 'h-4 w-4 border-2';
      break;
    case 'medium':
      spinnerSizeClasses = 'h-8 w-8 border-4';
      break;
    case 'large':
      spinnerSizeClasses = 'h-12 w-12 border-4';
      break;
  }

  return (
    <div className={`inline-block rounded-full border-solid border-gray-300 border-t-blue-500 animate-spin ${spinnerSizeClasses} ${className}`}></div>
  );
};

export default LoadingSpinner;