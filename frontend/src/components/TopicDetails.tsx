// src/components/TopicDetails.tsx
import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useParams, Link } from 'react-router-dom';
import { API_BASE_URL } from '../apiConfig';
import { TopicInfo } from '../types';

const TopicDetails: React.FC = () => {
  const { topicName } = useParams<{ topicName: string }>();
  const [topic, setTopic] = useState<TopicInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchTopicDetails = async () => {
      if (!topicName) return;

      try {
        const response = await axios.get<TopicInfo>(`${API_BASE_URL}/topics/${topicName}`);
        setTopic(response.data);
        setLoading(false);
      } catch (e: any) {
        console.error("Error fetching topic details:", e);
        setError(e.response?.data?.error || e.message);
        setLoading(false);
      }
    };

    fetchTopicDetails();
  }, [topicName]);

  if (loading) {
    return <div className="p-4 text-gray-600">Loading topic details...</div>;
  }

  if (error) {
    return <div className="p-4 text-red-600 bg-white rounded-lg shadow p-6">Error: {error}</div>;
  }

  if (!topic) {
    return (
      <div className="p-4">
        <div className="bg-white rounded-lg shadow p-6 text-center">
          <p className="text-gray-600 mb-4">Topic not found.</p>
          <Link 
            to="/topics" 
            className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-accent-blue hover:bg-blue-600"
          >
            Back to Topics
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="p-4">
      <div className="mb-4 flex justify-between items-center">
        <h2 className="text-2xl font-semibold text-gray-800">Topic: {topic.name}</h2>
        <Link 
          to="/topics" 
          className="px-4 py-2 text-sm text-accent-blue hover:text-blue-700"
        >
          ← Back to Topics
        </Link>
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="p-6 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-800 mb-4">Overview</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="bg-gray-50 p-4 rounded">
              <p className="text-sm text-gray-500">Number of Partitions</p>
              <p className="text-lg font-medium text-gray-800">{topic.numPartitions}</p>
            </div>
            <div className="bg-gray-50 p-4 rounded">
              <p className="text-sm text-gray-500">Replication Factor</p>
              <p className="text-lg font-medium text-gray-800">{topic.replicationFactor}</p>
            </div>
          </div>
        </div>

        {topic.partitions && topic.partitions.length > 0 && (
          <div className="p-6 border-b border-gray-200">
            <h3 className="text-lg font-medium text-gray-800 mb-4">Partitions</h3>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Partition ID
                    </th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Leader
                    </th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Replicas
                    </th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      In-Sync Replicas
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {topic.partitions.map((partition) => (
                    <tr key={partition.id}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {partition.id}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {partition.leader}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {partition.replicas.join(', ')}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {partition.isr.join(', ')}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {topic.config && Object.keys(topic.config).length > 0 && (
          <div className="p-6">
            <h3 className="text-lg font-medium text-gray-800 mb-4">Configuration</h3>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Key
                    </th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Value
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {Object.entries(topic.config).map(([key, value]) => (
                    <tr key={key}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {key}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {value}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default TopicDetails;