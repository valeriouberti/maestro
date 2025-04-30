import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Link, useNavigate } from 'react-router-dom';
import { API_BASE_URL } from '../apiConfig';
import { TopicInfo } from '../types';

interface ConfigItem {
  key: string;
  value: string;
}

const TopicList: React.FC = () => {
  const [topics, setTopics] = useState<TopicInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  
  // Form state
  const [topicName, setTopicName] = useState('');
  const [numPartitions, setNumPartitions] = useState(1);
  const [replicationFactor, setReplicationFactor] = useState(1);
  const [configItems, setConfigItems] = useState<ConfigItem[]>([
    { key: 'cleanup.policy', value: 'delete' },
    { key: 'retention.ms', value: '86400000' }
  ]);
  const [formError, setFormError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  
  const navigate = useNavigate();

  const fetchTopics = async () => {
    try {
      setLoading(true);
      const response = await axios.get(`${API_BASE_URL}/topics`);
      
      // Check if response.data has a topics property
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

  useEffect(() => {
    fetchTopics();
  }, []);

  const handleConfigChange = (index: number, field: 'key' | 'value', value: string) => {
    const newConfigItems = [...configItems];
    newConfigItems[index][field] = value;
    setConfigItems(newConfigItems);
  };

  const addConfigItem = () => {
    setConfigItems([...configItems, { key: '', value: '' }]);
  };

  const removeConfigItem = (index: number) => {
    const newConfigItems = [...configItems];
    newConfigItems.splice(index, 1);
    setConfigItems(newConfigItems);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);
    setSuccessMessage(null);

    // Convert config items array to object
    const configObject: Record<string, string> = {};
    configItems.forEach(item => {
      if (item.key && item.value) {
        configObject[item.key] = item.value;
      }
    });

    try {
      const response = await axios.post(`${API_BASE_URL}/topics`, {
        name: topicName,
        numPartitions: numPartitions,
        replicationFactor: replicationFactor,
        config: configObject
      });

      if (response.status === 201) {
        setSuccessMessage(`Topic "${topicName}" created successfully!`);
        // Reset form
        setTopicName('');
        setNumPartitions(1);
        setReplicationFactor(1);
        setConfigItems([
          { key: 'cleanup.policy', value: 'delete' },
          { key: 'retention.ms', value: '86400000' }
        ]);
        
        // Refresh topics list
        fetchTopics();
        
        // Hide form after successful creation
        setTimeout(() => {
          setShowCreateForm(false);
          setSuccessMessage(null);
        }, 1500);
      } else {
        setFormError(`Failed to create topic. Status code: ${response.status}`);
      }
    } catch (e: any) {
      setFormError(`Error creating topic: ${e.response?.data?.error || e.message}`);
    }
  };

  const toggleCreateForm = () => {
    setShowCreateForm(!showCreateForm);
    setFormError(null);
    setSuccessMessage(null);
  };

  const refreshTopics = () => {
    setLoading(true);
    fetchTopics();
  };

  if (loading && topics.length === 0) {
    return <div className="p-4 text-gray-600">Loading topics...</div>;
  }

  return (
    <div className="p-4">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-semibold text-gray-800">Topics</h2>
        <div className="flex space-x-2">
          <button
            onClick={refreshTopics}
            className="px-4 py-2 text-accent-blue bg-white border border-accent-blue rounded-md hover:bg-blue-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 flex items-center"
          >
            <svg 
              xmlns="http://www.w3.org/2000/svg" 
              className="h-5 w-5 mr-1" 
              viewBox="0 0 20 20" 
              fill="currentColor"
            >
              <path 
                fillRule="evenodd" 
                d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" 
                clipRule="evenodd" 
              />
            </svg>
            Refresh
          </button>
          <button
            onClick={toggleCreateForm}
            className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 flex items-center"
          >
            {showCreateForm ? (
              <span>Cancel</span>
            ) : (
              <>
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-1" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clipRule="evenodd" />
                </svg>
                <span>Create Topic</span>
              </>
            )}
          </button>
        </div>
      </div>
      
      {successMessage && (
        <div className="mb-4 p-2 bg-green-50 border border-green-200 rounded text-green-600">
          {successMessage}
        </div>
      )}

      {showCreateForm && (
        <div className="mb-6 bg-white rounded-lg shadow-sm p-6">
          <h3 className="text-lg font-medium text-gray-800 mb-4">Create New Topic</h3>
          {formError && <div className="text-red-500 mb-4 p-2 bg-red-50 border border-red-200 rounded">{formError}</div>}
          
          <form onSubmit={handleSubmit}>
            <div className="mb-4">
              <label htmlFor="topicName" className="block text-sm font-medium text-gray-700">
                Topic Name
              </label>
              <input
                type="text"
                id="topicName"
                className="mt-1 p-2 w-full border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue"
                value={topicName}
                onChange={(e) => setTopicName(e.target.value)}
                required
              />
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
              <div>
                <label htmlFor="numPartitions" className="block text-sm font-medium text-gray-700">
                  Number of Partitions
                </label>
                <input
                  type="number"
                  id="numPartitions"
                  className="mt-1 p-2 w-full border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue"
                  value={numPartitions}
                  onChange={(e) => setNumPartitions(parseInt(e.target.value) || 1)}
                  min="1"
                />
              </div>
              
              <div>
                <label htmlFor="replicationFactor" className="block text-sm font-medium text-gray-700">
                  Replication Factor
                </label>
                <input
                  type="number"
                  id="replicationFactor"
                  className="mt-1 p-2 w-full border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue"
                  value={replicationFactor}
                  onChange={(e) => setReplicationFactor(parseInt(e.target.value) || 1)}
                  min="1"
                />
              </div>
            </div>
            
            <div className="mb-6">
              <div className="flex justify-between items-center mb-2">
                <label className="block text-sm font-medium text-gray-700">
                  Topic Configuration
                </label>
                <button 
                  type="button" 
                  onClick={addConfigItem}
                  className="text-sm text-accent-blue hover:text-blue-700"
                >
                  + Add Configuration
                </button>
              </div>
              
              {configItems.map((item, index) => (
                <div key={index} className="flex items-center space-x-2 mb-2">
                  <input
                    type="text"
                    placeholder="Key"
                    className="p-2 w-1/2 border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue"
                    value={item.key}
                    onChange={(e) => handleConfigChange(index, 'key', e.target.value)}
                  />
                  <input
                    type="text"
                    placeholder="Value"
                    className="p-2 w-1/2 border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue"
                    value={item.value}
                    onChange={(e) => handleConfigChange(index, 'value', e.target.value)}
                  />
                  <button 
                    type="button" 
                    onClick={() => removeConfigItem(index)}
                    className="p-2 text-red-500 hover:text-red-700"
                  >
                    ×
                  </button>
                </div>
              ))}
            </div>
            
            <div className="flex justify-end">
              <button
                type="submit"
                className="bg-blue-500 text-white py-2 px-4 rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 font-medium shadow-sm"
              >
                Create Topic
              </button>
            </div>
          </form>
        </div>
      )}

      {error && <div className="p-4 text-red-600">Error: {error}</div>}
      
      {!loading && (!topics || topics.length === 0) ? (
        <div className="bg-white rounded-lg shadow p-6 text-center">
          <p className="text-gray-600 mb-4">No topics found.</p>
          <button
            onClick={() => setShowCreateForm(true)}
            className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-500 hover:bg-blue-600"
          >
            Create Your First Topic
          </button>
        </div>
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