
import React from 'react';
import { Alert as ShadcnAlert, AlertDescription, AlertTitle } from '../ui/alert'; 

interface AlertProps {
  type: 'default' | 'success' | 'info' | 'warning' | 'error';
  title?: string;
  description: string;
  className?: string; 
}

const Alert: React.FC<AlertProps> = ({ type, title, description, className }) => {
  let alertClasses = "border ";
  let icon = null;

  switch (type) {
    case 'success':
      alertClasses += "border-green-400 bg-green-50 text-green-700";

      title = title || "Success!";
      break;
    case 'info':
      alertClasses += "border-blue-400 bg-blue-50 text-blue-700";

      title = title || "Info";
      break;
    case 'warning':
      alertClasses += "border-yellow-400 bg-yellow-50 text-yellow-700";

      title = title || "Warning!";
      break;
    case 'error':
      alertClasses += "border-red-400 bg-red-50 text-red-700";

      title = title || "Error!";
      break;
    default:
      alertClasses += "border-gray-400 bg-gray-50 text-gray-700";
      title = title || "Notification";
  }

  return (
    <ShadcnAlert className={`mb-4 flex items-start space-x-2 ${alertClasses} ${className}`}> {}
      {icon && <div className="mt-1">{icon}</div>}
      <div>
        {title && <AlertTitle className="text-sm font-bold">{title}</AlertTitle>}
        <AlertDescription className="text-sm">{description}</AlertDescription>
      </div>
    </ShadcnAlert>
  );
};

export default Alert;