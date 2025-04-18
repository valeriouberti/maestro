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
            className={`flex items-center py-2 px-4 rounded transition-colors ${
              isActive('/clusters') 
                ? 'bg-sidebar-hover text-accent-blue font-medium' 
                : 'hover:bg-sidebar-hover text-gray-700'
            }`}
          >
            <svg 
              xmlns="http://www.w3.org/2000/svg" 
              className="h-5 w-5 mr-2" 
              viewBox="0 0 20 20" 
              fill="currentColor"
              aria-hidden="true"
            >
              <path 
                d="M13 7H7v6h6V7z" 
              />
              <path 
                fillRule="evenodd" 
                d="M7 2a1 1 0 012 0v1h2V2a1 1 0 112 0v1h2a2 2 0 012 2v2h1a1 1 0 110 2h-1v2h1a1 1 0 110 2h-1v2a2 2 0 01-2 2h-2v1a1 1 0 11-2 0v-1H9v1a1 1 0 11-2 0v-1H5a2 2 0 01-2-2v-2H2a1 1 0 110-2h1V9H2a1 1 0 010-2h1V5a2 2 0 012-2h2V2zM5 5h10v10H5V5z" 
                clipRule="evenodd" 
              />
            </svg>
            Clusters
          </Link>
        </li>
        <li className="mb-2">
          <Link 
            to="/topics" 
            className={`flex items-center py-2 px-4 rounded transition-colors ${
              isActive('/topics') 
                ? 'bg-sidebar-hover text-accent-blue font-medium' 
                : 'hover:bg-sidebar-hover text-gray-700'
            }`}
          >
            <svg 
              xmlns="http://www.w3.org/2000/svg" 
              className="h-5 w-5 mr-2" 
              viewBox="0 0 20 20" 
              fill="currentColor"
              aria-hidden="true"
            >
              <path 
                fillRule="evenodd" 
                d="M2 5a2 2 0 012-2h12a2 2 0 012 2v10a2 2 0 01-2 2H4a2 2 0 01-2-2V5zm3.293 1.293a1 1 0 011.414 0l3 3a1 1 0 010 1.414l-3 3a1 1 0 01-1.414-1.414L7.586 10 5.293 7.707a1 1 0 010-1.414zM11 12a1 1 0 100 2h3a1 1 0 100-2h-3z" 
                clipRule="evenodd" 
              />
            </svg>
            Topics
          </Link>
        </li>
        <li className="mb-2">
          <Link 
            to="/consumer-groups" 
            className={`flex items-center py-2 px-4 rounded transition-colors ${
              isActive('/consumer-groups') 
                ? 'bg-sidebar-hover text-accent-blue font-medium' 
                : 'hover:bg-sidebar-hover text-gray-700'
            }`}
          >
            <svg 
              xmlns="http://www.w3.org/2000/svg" 
              className="h-5 w-5 mr-2" 
              viewBox="0 0 20 20" 
              fill="currentColor"
              aria-hidden="true"
            >
              <path 
                d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v1h8v-1zM6 8a2 2 0 11-4 0 2 2 0 014 0zM16 18v-1a5.972 5.972 0 00-.75-2.906A3.005 3.005 0 0119 15v1h-3zM4.75 12.094A5.973 5.973 0 004 15v1H1v-1a3 3 0 013.75-2.906z" 
              />
            </svg>
            Consumer Groups
          </Link>
        </li>
      </ul>
    </div>
  );
};

export default Sidebar;