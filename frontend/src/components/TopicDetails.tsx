// src/components/TopicDetails.tsx
import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useParams } from 'react-router-dom';
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
        setError(e.message);
        setLoading(false);
      }
    };

    fetchTopicDetails();
  }, [topicName]);

  if (loading) {
    return <div className="p-4">Loading topic details...</div>;
  }

  if (error) {
    return <div className="p-4 text-red-500">Error: {error}</div>;
  }

  if (!topic) {
    return <div className="p-4">Topic not found.</div>;
  }

  return (
    <div className="p-4">
      <h2 className="text-xl font-semibold mb-4 text-pastel-purple">Topic Details: {topic.name}</h2>
      <p>Number of Partitions: {topic.numPartitions}</p>
      <p>Replication Factor: {topic.replicationFactor}</p>

      {topic.config && Object.keys(topic.config).length > 0 && (
        <>
          <h3 className="text-lg font-semibold mt-4 mb-2">Configuration:</h3>
          <ul className="list-disc pl-5">
            {Object.entries(topic.config).map(([key, value]) => (
              <li key={key}>
                {key}: {value}
              </li>
            ))}
          </ul>
        </>
      )}

      {topic.partitions && topic.partitions.length > 0 && (
        <>
          <h3 className="text-lg font-semibold mt-4 mb-2">Partitions:</h3>
          <ul className="list-disc pl-5">
            {topic.partitions.map((partition) => (
              <li key={partition.id}>
                Partition ID: {partition.id}, Leader: {partition.leader}, Replicas: {partition.replicas.join(', ')}, ISR: {partition.isr.join(', ')}
              </li>
            ))}
          </ul>
        </>
      )}
    </div>
  );
};

export default TopicDetails;