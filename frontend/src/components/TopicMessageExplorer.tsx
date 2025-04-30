import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useParams, Link } from 'react-router-dom';
import { API_BASE_URL } from '../apiConfig';
import { TopicInfo, TopicMessage, PartitionInfo } from '../types';

const TopicMessageExplorer: React.FC = () => {
  const { topicName } = useParams<{ topicName: string }>();
  const [topic, setTopic] = useState<TopicInfo | null>(null);
  const [messages, setMessages] = useState<TopicMessage[]>([]);
  const [loading, setLoading] = useState(true);
  const [messageLoading, setMessageLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [messageError, setMessageError] = useState<string | null>(null);

  // Query parameters
  const [selectedPartition, setSelectedPartition] = useState<number>(0);
  const [offset, setOffset] = useState<string>("earliest");
  const [limit, setLimit] = useState<number>(100);
  const [customOffset, setCustomOffset] = useState<number>(0);
  const [expandedMessages, setExpandedMessages] = useState<Set<number>>(new Set());
  const [messageFormat, setMessageFormat] = useState<'text' | 'json'>('text');
  const [searchTerm, setSearchTerm] = useState<string>('');
  const [filteredMessages, setFilteredMessages] = useState<TopicMessage[]>([]);

  useEffect(() => {
    // Fetch topic details first to get partition information
    const fetchTopicDetails = async () => {
      if (!topicName) return;

      try {
        setLoading(true);
        const response = await axios.get(`${API_BASE_URL}/topics/${topicName}`);
        if (response.data && response.data.topic) {
          setTopic(response.data.topic);
        } else {
          setError("Received invalid topic data format from server");
        }
      } catch (e: any) {
        console.error("Error fetching topic details:", e);
        setError(e.response?.data?.message || e.message);
      } finally {
        setLoading(false);
      }
    };

    fetchTopicDetails();
  }, [topicName]);

  // Fetch messages when parameters change or when explicitly requested
  const fetchMessages = async () => {
    if (!topicName) return;

    try {
      setMessageLoading(true);
      setMessageError(null);
      
      // Determine which offset to use
      let offsetValue = offset;
      if (offset === 'custom') {
        offsetValue = customOffset.toString();
      }

      const response = await axios.get(`${API_BASE_URL}/topics/${topicName}/messages`, {
        params: {
          partition: selectedPartition,
          offset: offset === 'latest' ? 'latest' : offsetValue,
          limit: limit
        }
      });

      if (response.data && Array.isArray(response.data.messages)) {
        setMessages(response.data.messages);
      } else {
        setMessageError("Received invalid messages data format from server");
        setMessages([]);
      }
    } catch (e: any) {
      console.error("Error fetching messages:", e);
      setMessageError(e.response?.data?.message || e.message);
      setMessages([]);
    } finally {
      setMessageLoading(false);
    }
  };

  // Apply search filter whenever messages or search term changes
  useEffect(() => {
    if (searchTerm.trim() === '') {
      setFilteredMessages(messages);
      return;
    }

    const lowerSearchTerm = searchTerm.toLowerCase();
    const filtered = messages.filter(msg => {
      const keyMatch = msg.key && msg.key.toLowerCase().includes(lowerSearchTerm);
      const valueMatch = msg.value && msg.value.toLowerCase().includes(lowerSearchTerm);
      
      // Also search in headers if they exist
      let headerMatch = false;
      if (msg.headers) {
        headerMatch = Object.entries(msg.headers).some(([key, value]) => {
          return key.toLowerCase().includes(lowerSearchTerm) || 
                 (value && value.toLowerCase().includes(lowerSearchTerm));
        });
      }
      
      return keyMatch || valueMatch || headerMatch;
    });
    
    setFilteredMessages(filtered);
  }, [messages, searchTerm]);

  const toggleMessageExpand = (index: number) => {
    const newExpanded = new Set(expandedMessages);
    if (newExpanded.has(index)) {
      newExpanded.delete(index);
    } else {
      newExpanded.add(index);
    }
    setExpandedMessages(newExpanded);
  };

  const formatMessageValue = (value: string, format: 'text' | 'json'): string => {
    if (!value) return '';
    
    if (format === 'json') {
      try {
        const parsed = JSON.parse(value);
        return JSON.stringify(parsed, null, 2);
      } catch (e) {
        // Not valid JSON, return as is
        return value;
      }
    }
    
    return value;
  };

  const isJsonString = (str: string): boolean => {
    try {
      JSON.parse(str);
      return true;
    } catch (e) {
      return false;
    }
  };

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
        <h2 className="text-2xl font-semibold text-gray-800">Messages: {topic.name}</h2>
        <div className="flex items-center space-x-2">
          <Link 
            to={`/topics/${topicName}`} 
            className="px-4 py-2 text-sm text-accent-blue hover:text-blue-700"
          >
            Topic Details
          </Link>
          <Link 
            to="/topics" 
            className="px-4 py-2 text-sm text-accent-blue hover:text-blue-700"
          >
            ← Back to Topics
          </Link>
        </div>
      </div>

      {/* Query Controls */}
      <div className="bg-white rounded-lg shadow mb-6 p-6">
        <h3 className="text-lg font-medium text-gray-800 mb-4">Message Explorer</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
          <div>
            <label htmlFor="partition" className="block text-sm font-medium text-gray-700 mb-1">
              Partition
            </label>
            <select
              id="partition"
              className="w-full p-2 border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue"
              value={selectedPartition}
              onChange={(e) => setSelectedPartition(parseInt(e.target.value))}
            >
              {topic.partitions && topic.partitions.map((p: PartitionInfo) => (
                <option key={p.id} value={p.id}>
                  Partition {p.id}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label htmlFor="offset" className="block text-sm font-medium text-gray-700 mb-1">
              Offset
            </label>
            <select
              id="offset"
              className="w-full p-2 border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue"
              value={offset}
              onChange={(e) => setOffset(e.target.value)}
            >
              <option value="earliest">Earliest</option>
              <option value="latest">Latest</option>
              <option value="custom">Custom</option>
            </select>
            {offset === 'custom' && (
              <div className="mt-2">
                <input
                  type="number"
                  min="0"
                  className="w-full p-2 border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue"
                  value={customOffset}
                  onChange={(e) => setCustomOffset(parseInt(e.target.value))}
                  placeholder="Enter offset value"
                />
              </div>
            )}
          </div>

          <div>
            <label htmlFor="limit" className="block text-sm font-medium text-gray-700 mb-1">
              Limit
            </label>
            <select
              id="limit"
              className="w-full p-2 border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue"
              value={limit}
              onChange={(e) => setLimit(parseInt(e.target.value))}
            >
              <option value={10}>10 messages</option>
              <option value={50}>50 messages</option>
              <option value={100}>100 messages</option>
              <option value={500}>500 messages</option>
              <option value={1000}>1000 messages</option>
            </select>
          </div>

          <div>
            <label htmlFor="format" className="block text-sm font-medium text-gray-700 mb-1">
              Message Format
            </label>
            <select
              id="format"
              className="w-full p-2 border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue"
              value={messageFormat}
              onChange={(e) => setMessageFormat(e.target.value as 'text' | 'json')}
            >
              <option value="text">Plain Text</option>
              <option value="json">JSON (pretty)</option>
            </select>
          </div>
        </div>

        <div className="mb-4">
          <label htmlFor="search" className="block text-sm font-medium text-gray-700 mb-1">
            Search Messages
          </label>
          <input
            type="text"
            id="search"
            className="w-full p-2 border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Search in message keys and values..."
          />
        </div>

        <div className="flex justify-end">
          <button
            onClick={fetchMessages}
            className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 flex items-center"
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
            Fetch Messages
          </button>
        </div>
      </div>

      {/* Messages Display */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        {messageLoading ? (
          <div className="p-8 text-center">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-accent-blue"></div>
            <p className="mt-2 text-gray-600">Loading messages...</p>
          </div>
        ) : messageError ? (
          <div className="p-6 text-red-500">
            <p className="font-medium">Error fetching messages:</p>
            <p>{messageError}</p>
          </div>
        ) : filteredMessages.length === 0 ? (
          <div className="p-6 text-center">
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
                d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" 
              />
            </svg>
            <h3 className="mt-2 text-sm font-medium text-gray-900">No messages found</h3>
            <p className="mt-1 text-sm text-gray-500">
              {messages.length > 0 ? 
                "No messages match your search criteria." : 
                "Try adjusting your query parameters and fetch again."}
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Offset
                  </th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Timestamp
                  </th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Key
                  </th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-1/2">
                    Value
                  </th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {filteredMessages.map((message, index) => {
                  const isExpanded = expandedMessages.has(index);
                  const hasJsonValue = isJsonString(message.value);
                  
                  return (
                    <tr key={`${message.partition}-${message.offset}`} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {message.offset}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {new Date(message.timestamp).toLocaleString()}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {message.key?.length > 30 && !isExpanded ? 
                          `${message.key.substring(0, 30)}...` : 
                          message.key || <span className="text-gray-400 italic">null</span>}
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500">
                        {isExpanded ? (
                          <pre className="overflow-x-auto max-w-full whitespace-pre-wrap break-words bg-gray-50 p-3 rounded">
                            {formatMessageValue(message.value, messageFormat)}
                          </pre>
                        ) : (
                          <div className="truncate max-w-md">
                            {message.value?.length > 50 ? 
                              `${message.value.substring(0, 50)}...` : 
                              message.value || <span className="text-gray-400 italic">null</span>}
                            {hasJsonValue && <span className="ml-1 text-blue-500">[json]</span>}
                          </div>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        <button
                          onClick={() => toggleMessageExpand(index)}
                          className="text-accent-blue hover:text-blue-700"
                        >
                          {isExpanded ? "Collapse" : "Expand"}
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
        {filteredMessages.length > 0 && (
          <div className="px-6 py-3 bg-gray-50 text-sm text-gray-500">
            Showing {filteredMessages.length} of {messages.length} messages
            {searchTerm && ` (filtered by "${searchTerm}")`}
          </div>
        )}
      </div>
    </div>
  );
};

export default TopicMessageExplorer;