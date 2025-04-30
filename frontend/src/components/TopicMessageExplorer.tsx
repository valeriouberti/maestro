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

  // New Publishing state
  const [showPublishForm, setShowPublishForm] = useState<boolean>(false);
  const [publishKey, setPublishKey] = useState<string>('');
  const [publishValue, setPublishValue] = useState<string>('');
  const [publishPartition, setPublishPartition] = useState<number>(-1);
  const [publishHeaders, setPublishHeaders] = useState<{key: string, value: string}[]>([]);
  const [isPublishing, setIsPublishing] = useState<boolean>(false);
  const [publishSuccess, setPublishSuccess] = useState<string | null>(null);
  const [publishError, setPublishError] = useState<string | null>(null);
  const [valueIsJson, setValueIsJson] = useState<boolean>(false);

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

      // Add timeout handling with a longer timeout for 'latest' offset
      const timeoutMs = offset === 'latest' ? 60000 : 30000; // 60s for latest, 30s otherwise
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

      const response = await axios.get(`${API_BASE_URL}/topics/${topicName}/messages`, {
        signal: controller.signal,
        params: {
          partition: selectedPartition,
          offset: offset === 'latest' ? 'latest' : offsetValue,
          limit: limit
        },
        // Increase the axios timeout
        timeout: timeoutMs
      });
      
      clearTimeout(timeoutId);

      if (response.data && Array.isArray(response.data.messages)) {
        setMessages(response.data.messages);
      } else {
        setMessageError("Received invalid messages data format from server");
        setMessages([]);
      }
    } catch (e: any) {
      console.error("Error fetching messages:", e);
      
      // Handle different error types
      if (e.name === 'AbortError' || e.code === 'ECONNABORTED' || e.code === 'ETIMEDOUT') {
        setMessageError(
          "Request timed out. Retrieving messages with 'latest' offset might take longer than expected. " +
          "Try reducing the number of messages or using a specific offset instead."
        );
      } else if (e.response?.status === 504) {
        setMessageError(
          "The server took too long to process your request. " + 
          "When using 'latest' offset with large topics, try reducing the message limit or using a specific offset."
        );
      } else if (e.message === 'Network Error' || e.message.includes('socket hangup')) {
        setMessageError("Network error occurred. Please check your connection to the server.");
      } else {
        setMessageError(e.response?.data?.message || e.message);
      }
      
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

  // Add Header to publish form
  const addHeader = () => {
    setPublishHeaders([...publishHeaders, { key: '', value: '' }]);
  };

  // Remove Header from publish form
  const removeHeader = (index: number) => {
    const newHeaders = [...publishHeaders];
    newHeaders.splice(index, 1);
    setPublishHeaders(newHeaders);
  };

  // Update Header in publish form
  const updateHeader = (index: number, field: 'key' | 'value', value: string) => {
    const newHeaders = [...publishHeaders];
    newHeaders[index][field] = value;
    setPublishHeaders(newHeaders);
  };

  // Format JSON in publish form
  const formatJson = () => {
    if (!valueIsJson || !publishValue) return;

    try {
      const parsed = JSON.parse(publishValue);
      setPublishValue(JSON.stringify(parsed, null, 2));
    } catch (e) {
      // If it's not valid JSON, leave it as is
    }
  };

  // Publish Message
  const handlePublishMessage = async () => {
    if (!topicName) return;
    
    setPublishError(null);
    setPublishSuccess(null);
    setIsPublishing(true);

    const messageValue = publishValue;
    
    // If value is not required, but we want to ensure it's valid JSON if marked as such
    if (valueIsJson && publishValue) {
      try {
        // Parse and stringify to validate and normalize the JSON
        JSON.parse(publishValue);
      } catch (e) {
        setPublishError("Value is not valid JSON");
        setIsPublishing(false);
        return;
      }
    }

    // Construct headers object
    const headers: Record<string, string> = {};
    publishHeaders.forEach(header => {
      if (header.key && header.value) {
        headers[header.key] = header.value;
      }
    });

    // Prepare request payload
    const payload = {
      key: publishKey,
      value: messageValue,
      headers: Object.keys(headers).length > 0 ? headers : undefined,
      partition: publishPartition >= 0 ? publishPartition : undefined
    };

    try {
      const response = await axios.post(`${API_BASE_URL}/topics/${topicName}/messages`, payload);
      
      setPublishSuccess("Message published successfully!");
      
      // Clear the form
      setTimeout(() => {
        setPublishSuccess(null);
      }, 3000);
      
      // Option to fetch messages after publishing
      if (offset === 'latest') {
        fetchMessages();
      }
    } catch (e: any) {
      console.error("Error publishing message:", e);
      setPublishError(e.response?.data?.message || e.message);
    } finally {
      setIsPublishing(false);
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

      {/* Tab Buttons */}
      <div className="mb-6 border-b border-gray-200">
        <div className="flex">
          <button
            onClick={() => setShowPublishForm(false)}
            className={`py-2 px-4 focus:outline-none ${!showPublishForm ? 'text-blue-500 border-b-2 border-blue-500 font-medium' : 'text-gray-500 hover:text-gray-700'}`}
          >
            Explore Messages
          </button>
          <button
            onClick={() => setShowPublishForm(true)}
            className={`py-2 px-4 focus:outline-none ${showPublishForm ? 'text-blue-500 border-b-2 border-blue-500 font-medium' : 'text-gray-500 hover:text-gray-700'}`}
          >
            Publish Message
          </button>
        </div>
      </div>

      {showPublishForm ? (
        // Publish Message Form
        <div className="bg-white rounded-lg shadow mb-6 p-6">
          <h3 className="text-lg font-medium text-gray-800 mb-4">Publish New Message</h3>
          
          {publishSuccess && (
            <div className="mb-4 p-3 bg-green-50 border border-green-200 rounded text-green-600">
              {publishSuccess}
            </div>
          )}
          
          {publishError && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-600">
              {publishError}
            </div>
          )}
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Key
              </label>
              <input
                type="text"
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue"
                value={publishKey}
                onChange={(e) => setPublishKey(e.target.value)}
                placeholder="Message key (optional)"
              />
              <p className="mt-1 text-xs text-gray-500">
                Keys are used for partition assignment and message ordering
              </p>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Partition
              </label>
              <select
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue"
                value={publishPartition}
                onChange={(e) => setPublishPartition(parseInt(e.target.value))}
              >
                <option value={-1}>Auto-select partition</option>
                {topic.partitions && topic.partitions.map((p: PartitionInfo) => (
                  <option key={p.id} value={p.id}>
                    Partition {p.id}
                  </option>
                ))}
              </select>
              <p className="mt-1 text-xs text-gray-500">
                Leave on auto-select to let Kafka determine the partition
              </p>
            </div>
          </div>
          
          <div className="mb-6">
            <div className="flex justify-between items-center mb-1">
              <label className="block text-sm font-medium text-gray-700">
                Value
              </label>
              <div className="flex items-center">
                <input
                  type="checkbox"
                  id="json-value"
                  className="mr-2"
                  checked={valueIsJson}
                  onChange={(e) => setValueIsJson(e.target.checked)}
                />
                <label htmlFor="json-value" className="text-sm text-gray-600">
                  JSON Value
                </label>
                {valueIsJson && (
                  <button
                    onClick={formatJson}
                    className="ml-2 text-sm text-accent-blue hover:text-blue-700"
                    type="button"
                  >
                    Format JSON
                  </button>
                )}
              </div>
            </div>
            <textarea
              className="w-full p-2 border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue h-40 font-mono"
              value={publishValue}
              onChange={(e) => setPublishValue(e.target.value)}
              placeholder={valueIsJson ? '{\n  "example": "json value"\n}' : "Message value"}
              required
            />
          </div>
          
          <div className="mb-6">
            <div className="flex justify-between items-center mb-2">
              <label className="block text-sm font-medium text-gray-700">
                Headers
              </label>
              <button 
                type="button" 
                onClick={addHeader}
                className="text-sm text-accent-blue hover:text-blue-700"
              >
                + Add Header
              </button>
            </div>
            
            {publishHeaders.map((header, index) => (
              <div key={index} className="flex items-center space-x-2 mb-2">
                <input
                  type="text"
                  placeholder="Header key"
                  className="p-2 w-1/2 border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue"
                  value={header.key}
                  onChange={(e) => updateHeader(index, 'key', e.target.value)}
                />
                <input
                  type="text"
                  placeholder="Header value"
                  className="p-2 w-1/2 border border-gray-300 rounded-md focus:ring-accent-blue focus:border-accent-blue"
                  value={header.value}
                  onChange={(e) => updateHeader(index, 'value', e.target.value)}
                />
                <button 
                  type="button" 
                  onClick={() => removeHeader(index)}
                  className="p-2 text-red-500 hover:text-red-700"
                >
                  ×
                </button>
              </div>
            ))}
            
            {publishHeaders.length === 0 && (
              <p className="text-sm text-gray-500 italic">No headers added</p>
            )}
          </div>
          
          <div className="flex justify-end">
            <button
              type="button"
              onClick={handlePublishMessage}
              className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 flex items-center disabled:opacity-50"
              disabled={isPublishing || !publishValue}
            >
              {isPublishing ? (
                <>
                  <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Publishing...
                </>
              ) : (
                <>
                  <svg 
                    xmlns="http://www.w3.org/2000/svg" 
                    className="h-5 w-5 mr-1" 
                    viewBox="0 0 20 20" 
                    fill="currentColor"
                  >
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-8.707l-3-3a1 1 0 00-1.414 0l-3 3a1 1 0 001.414 1.414L9 9.414V13a1 1 0 102 0V9.414l1.293 1.293a1 1 0 001.414-1.414z" clipRule="evenodd" />
                  </svg>
                  Publish Message
                </>
              )}
            </button>
          </div>
        </div>
      ) : (
        // Message Explorer
        <>
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
              <div className="p-6">
                <div className="bg-red-50 border border-red-200 rounded p-4 mb-4">
                  <h3 className="text-red-700 font-medium mb-2">Error fetching messages:</h3>
                  <p className="text-red-600">{messageError}</p>
                  
                  {messageError.includes('timed out') && (
                    <div className="mt-3 text-sm text-gray-700">
                      <p className="font-medium">Troubleshooting tips:</p>
                      <ul className="list-disc pl-5 mt-1 space-y-1">
                        <li>Try reducing the number of messages (current: {limit})</li>
                        <li>Use a specific offset instead of "latest"</li>
                        <li>Try a different partition</li>
                        <li>The topic might have very large messages</li>
                      </ul>
                    </div>
                  )}
                </div>
                
                <div className="flex space-x-4">
                  <button
                    onClick={fetchMessages}
                    className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                  >
                    Try Again
                  </button>
                  
                  {offset === 'latest' && (
                    <button
                      onClick={() => {
                        setOffset('earliest');
                        setTimeout(fetchMessages, 100);
                      }}
                      className="px-4 py-2 border border-blue-500 text-blue-500 rounded-md hover:bg-blue-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                    >
                      Try with Earliest Offset
                    </button>
                  )}
                </div>
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
        </>
      )}
    </div>
  );
};

export default TopicMessageExplorer;