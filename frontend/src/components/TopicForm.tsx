// src/components/TopicForm.tsx
import React, { useState } from 'react';
import axios from 'axios';
import { API_BASE_URL } from '../apiConfig';
import { useNavigate } from 'react-router-dom';

interface ConfigItem {
  key: string;
  value: string;
}

const TopicForm: React.FC = () => {
    const [topicName, setTopicName] = useState('');
    const [numPartitions, setNumPartitions] = useState(1);
    const [replicationFactor, setReplicationFactor] = useState(1);
    const [configItems, setConfigItems] = useState<ConfigItem[]>([
      { key: 'cleanup.policy', value: 'delete' },
      { key: 'retention.ms', value: '86400000' }
    ]);
    const [error, setError] = useState<string | null>(null);
    const [successMessage, setSuccessMessage] = useState<string | null>(null);
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
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
                // Redirect to the topic list page after creation
                setTimeout(() => {
                    navigate('/topics');
                }, 1500); // Redirect after 1.5 seconds
            } else {
                setError(`Failed to create topic. Status code: ${response.status}`);
            }
        } catch (e: any) {
            setError(`Error creating topic: ${e.response?.data?.error || e.message}`);
        }
    };

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

    return (
        <div className="p-4">
            <h2 className="text-2xl font-semibold mb-6 text-gray-800">Create New Topic</h2>
            {error && <div className="text-red-500 mb-4 p-2 bg-red-50 border border-red-200 rounded">{error}</div>}
            {successMessage && <div className="text-green-500 mb-4 p-2 bg-green-50 border border-green-200 rounded">{successMessage}</div>}
            
            <form onSubmit={handleSubmit} className="bg-white rounded-lg shadow-sm p-6">
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
                        type="button"
                        className="mr-2 bg-gray-100 text-gray-700 py-2 px-4 rounded-md hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-500"
                        onClick={() => navigate('/topics')}
                    >
                        Cancel
                    </button>
                    <button
                        type="submit"
                        className="bg-blue-500 text-white py-2 px-4 rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 font-medium shadow-sm"
                    >
                        Create Topic
                    </button>
                </div>
            </form>
        </div>
    );
};

export default TopicForm;