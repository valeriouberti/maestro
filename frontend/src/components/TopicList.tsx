import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Link } from 'react-router-dom';
import { API_BASE_URL } from '../apiConfig';
import { TopicInfo } from '../types';

const TopicList: React.FC = () => {
  const [topics, setTopics] = useState<TopicInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchTopics = async () => {
      try {
        const response = await axios.get(`${API_BASE_URL}/topics`);
        
        // Check if response.data has a topics property (likely response format is { topics: [...] })
        if (response.data && Array.isArray(response.data.topics)) {
          setTopics(response.data.topics);
        } else if (Array.isArray(response.data)) {
          // Handle case where response might be an array directly
          setTopics(response.data);
        } else {
          console.error("Unexpected response format:", response.data);
          setError("Received invalid data format from server");
        }
        
        setLoading(false);
      } catch (e: any) {
        console.error("Error fetching topics:", e);
        setError(e.message);
        setLoading(false);
      }
    };

    fetchTopics();
  }, []);

  if (loading) {
    return <div className="p-4 text-gray-600">Loading topics...</div>;
  }

  if (error) {
    return <div className="p-4 text-red-600">Error: {error}</div>;
  }

  return (
    <div className="p-4">
      <h2 className="text-2xl font-semibold mb-6 text-gray-800">Topics</h2>
      {!topics || topics.length === 0 ? (
        <p className="text-gray-600">No topics found.</p>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <ul className="divide-y divide-gray-200">
            {topics.map((topic) => (
              <li key={topic.name} className="p-4 hover:bg-gray-50">
                <Link to={`/topics/${topic.name}`} className="text-accent-blue hover:underline font-medium">
                  {topic.name}
                </Link>
                <p className="text-gray-600 mt-1">
                  Partitions: {topic.numPartitions}, Replication Factor: {topic.replicationFactor}
                </p>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
};

export default TopicList;