import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { API_BASE_URL } from '../apiConfig';
import { BrokerInfo } from '../types';

const ClusterList: React.FC = () => {
  const [clusters, setClusters] = useState<BrokerInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchClusters = async () => {
      try {
        const response = await axios.get(`${API_BASE_URL}/clusters`);
        // The API response likely contains brokers in a nested property
        const brokersData = response.data.brokers;
        
        if (Array.isArray(brokersData)) {
          setClusters(brokersData);
        } else {
          // Handle case where brokers data isn't an array
          console.error('Expected brokers data to be an array, got:', brokersData);
          setError('Invalid data format received from server');
        }
        setLoading(false);
      } catch (e: any) {
        console.error('Error fetching clusters:', e);
        setError(e.message);
        setLoading(false);
      }
    };

    fetchClusters();
  }, []);

  if (loading) {
    return <div className="p-4 text-gray-600">Loading clusters...</div>;
  }

  if (error) {
    return <div className="p-4 text-red-600">Error: {error}</div>;
  }

  return (
    <div className="p-4">
      <h2 className="text-2xl font-semibold mb-6 text-gray-800">Clusters</h2>
      {clusters.length === 0 ? (
        <p className="text-gray-600">No clusters found.</p>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <ul className="divide-y divide-gray-200">
            {clusters.map((cluster) => (
              <li key={cluster.id} className="p-4 hover:bg-gray-50">
                <div className="flex items-center">
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium text-accent-blue">
                      Broker ID: {cluster.id}
                    </p>
                    <p className="text-sm text-gray-600">
                      Host: {cluster.host}:{cluster.port}
                    </p>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
};

export default ClusterList;