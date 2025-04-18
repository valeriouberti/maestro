// src/components/ConsumerGroupDetails.tsx
import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useParams } from 'react-router-dom';
import { API_BASE_URL } from '../apiConfig';
import type { ConsumerGroupDetails } from '../types';

const ConsumerGroupDetails: React.FC = () => {
    const { groupId } = useParams<{ groupId: string }>();
    const [consumerGroup, setConsumerGroup] = useState<ConsumerGroupDetails | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchConsumerGroupDetails = async () => {
            if (!groupId) return;

            try {
                const response = await axios.get<ConsumerGroupDetails>(`${API_BASE_URL}/consumer-groups/${groupId}`);
                setConsumerGroup(response.data);
                setLoading(false);
            } catch (e: any) {
                setError(e.message);
                setLoading(false);
            }
        };

        fetchConsumerGroupDetails();
    }, [groupId]);

    if (loading) {
        return <div className="p-4">Loading consumer group details...</div>;
    }

    if (error) {
        return <div className="p-4 text-red-500">Error: {error}</div>;
    }

    if (!consumerGroup) {
        return <div className="p-4">Consumer group not found.</div>;
    }

    return (
        <div className="p-4">
            <h2 className="text-xl font-semibold mb-4 text-pastel-purple">Consumer Group Details: {consumerGroup.groupId}</h2>
            <p>State: {consumerGroup.state || 'Unknown'}</p>
            {consumerGroup.coordinator && (
                <p>Coordinator: Broker ID: {consumerGroup.coordinator.id}, Host: {consumerGroup.coordinator.host}:{consumerGroup.coordinator.port}</p>
            )}

            {consumerGroup.topics && consumerGroup.topics.length > 0 && (
                <>
                    <h3 className="text-lg font-semibold mt-4 mb-2">Topics:</h3>
                    <ul className="list-disc pl-5">
                        {consumerGroup.topics.map((topic) => (
                            <li key={topic}>{topic}</li>
                        ))}
                    </ul>
                </>
            )}

            {consumerGroup.members && consumerGroup.members.length > 0 && (
                <>
                    <h3 className="text-lg font-semibold mt-4 mb-2">Members:</h3>
                    <ul>
                        {consumerGroup.members.map((member) => (
                            <li key={member.clientId} className="mb-4 border p-2 rounded">
                                <p>Consumer ID: {member.consumerId}</p>
                                <p>Client ID: {member.clientId}</p>
                                <p>Host: {member.host}</p>
                                {member.assignments && member.assignments.length > 0 && (
                                    <>
                                        <h4 className="text-md font-semibold mt-2 mb-1">Assignments:</h4>
                                        <ul className="list-disc pl-5">
                                            {member.assignments.map((assignment) => (
                                                <li key={`${assignment.topic}-${assignment.partition}`}>
                                                    Topic: {assignment.topic}, Partition: {assignment.partition}
                                                </li>
                                            ))}
                                        </ul>
                                    </>
                                )}
                            </li>
                        ))}
                    </ul>
                </>
            )}
        </div>
    );
};

export default ConsumerGroupDetails;