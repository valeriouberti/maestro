// src/components/ConsumerGroupList.tsx
import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Link } from 'react-router-dom';
import { API_BASE_URL } from '../apiConfig';
import { ConsumerGroupInfo } from '../types';

const ConsumerGroupList: React.FC = () => {
    const [consumerGroups, setConsumerGroups] = useState<ConsumerGroupInfo[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchConsumerGroups = async () => {
            try {
                const response = await axios.get(`${API_BASE_URL}/consumer-groups`);
                
                // Check if response.data has a groups property
                if (response.data && Array.isArray(response.data.groups)) {
                    setConsumerGroups(response.data.groups);
                } else if (Array.isArray(response.data)) {
                    // Handle case where response might be an array directly
                    setConsumerGroups(response.data);
                } else {
                    console.error("Unexpected response format:", response.data);
                    setConsumerGroups([]);
                }
                
                setLoading(false);
            } catch (e: any) {
                console.error("Error fetching consumer groups:", e);
                setError(e.message);
                setLoading(false);
                setConsumerGroups([]); // Ensure we have an empty array on error
            }
        };

        fetchConsumerGroups();
    }, []);

    if (loading) {
        return <div className="p-4 text-gray-600">Loading consumer groups...</div>;
    }

    // Always show the empty state when there are no groups, regardless of error
    const isEmpty = consumerGroups.length === 0;

    return (
        <div className="p-4">
            <h2 className="text-2xl font-semibold mb-6 text-gray-800">Consumer Groups</h2>
            
            {isEmpty ? (
                <div className="bg-white rounded-lg shadow overflow-hidden p-6 text-center">
                    <svg 
                        className="mx-auto h-12 w-12 text-gray-400" 
                        fill="none" 
                        viewBox="0 0 24 24" 
                        stroke="currentColor" 
                        aria-hidden="true"
                    >
                        <path 
                            strokeLinecap="round" 
                            strokeLinejoin="round" 
                            strokeWidth="2" 
                            d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 0 002-2M9 5a2 2 0 012-2h2a2 0 012 2" 
                        />
                    </svg>
                    <h3 className="mt-2 text-sm font-medium text-gray-900">
                        No consumer groups
                    </h3>
                    <p className="mt-1 text-sm text-gray-500">
                        There are no consumer groups available in the Kafka cluster.
                    </p>
                    <div className="mt-6">
                        <Link 
                            to="/topics" 
                            className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium focus:bg-blue-600 focus:text-white focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                        >
                            View Topics
                        </Link>
                    </div>
                </div>
            ) : (
                <div className="bg-white rounded-lg shadow overflow-hidden">
                    <ul className="divide-y divide-gray-200">
                        {consumerGroups.map((group) => (
                            <li key={group.groupId} className="p-4 hover:bg-gray-50">
                                <Link 
                                    to={`/consumer-groups/${group.groupId}`} 
                                    className="text-accent-blue hover:underline font-medium"
                                >
                                    {group.groupId}
                                </Link>
                                <p className="text-gray-600 mt-1">
                                    State: {group.state || 'Unknown'}
                                </p>
                            </li>
                        ))}
                    </ul>
                </div>
            )}
        </div>
    );
};

export default ConsumerGroupList;