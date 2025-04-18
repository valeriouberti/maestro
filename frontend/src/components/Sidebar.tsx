// src/components/Sidebar.tsx
import React from 'react';
import { Link, useLocation } from 'react-router-dom';

const Sidebar: React.FC = () => {
  const location = useLocation();
  
  const isActive = (path: string) => {
    return location.pathname === path || 
           (path !== '/clusters' && location.pathname.startsWith(path));
  };

  return (
    <div className="w-64 bg-sidebar-blue h-screen fixed top-0 left-0 p-4 shadow-sm">
      <h2 className="text-2xl font-bold mb-6 text-accent-blue">Maestro</h2>
      <ul>
        <li className="mb-2">
          <Link 
            to="/clusters" 
            className={`block py-2 px-4 rounded transition-colors ${
              isActive('/clusters') 
                ? 'bg-sidebar-hover text-accent-blue font-medium' 
                : 'hover:bg-sidebar-hover text-gray-700'
            }`}
          >
            Clusters
          </Link>
        </li>
        <li className="mb-2">
          <Link 
            to="/topics" 
            className={`block py-2 px-4 rounded transition-colors ${
              isActive('/topics') 
                ? 'bg-sidebar-hover text-accent-blue font-medium' 
                : 'hover:bg-sidebar-hover text-gray-700'
            }`}
          >
            Topics
          </Link>
        </li>
        <li className="mb-2">
          <Link 
            to="/consumer-groups" 
            className={`block py-2 px-4 rounded transition-colors ${
              isActive('/consumer-groups') 
                ? 'bg-sidebar-hover text-accent-blue font-medium' 
                : 'hover:bg-sidebar-hover text-gray-700'
            }`}
          >
            Consumer Groups
          </Link>
        </li>
        <li className="mb-2">
          <Link 
            to="/topics/create" 
            className={`block py-2 px-4 rounded transition-colors ${
              isActive('/topics/create') 
                ? 'bg-sidebar-hover text-accent-blue font-medium' 
                : 'hover:bg-sidebar-hover text-gray-700'
            }`}
          >
            Create Topic
          </Link>
        </li>
      </ul>
    </div>
  );
};

export default Sidebar;