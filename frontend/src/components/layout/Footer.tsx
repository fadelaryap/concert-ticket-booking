
import React from 'react';

const Footer: React.FC = () => {
  return (
    <footer className="bg-gray-800 text-white p-4 text-center mt-auto w-full"> {}
      <p>&copy; {new Date().getFullYear()} Tiketmu. All rights reserved.</p>
    </footer>
  );
};

export default Footer;