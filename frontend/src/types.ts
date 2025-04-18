export interface BrokerInfo {
  id: number;
  host: string;
  port: number;
}

export interface TopicInfo {
  name: string;
  numPartitions: number;
  replicationFactor: number;
  config?: { [key: string]: string };
  partitions?: PartitionInfo[];
}

export interface PartitionInfo {
  id: number;
  leader: number;
  replicas: number[];
  isr: number[];
}

export interface ConsumerGroupInfo {
  groupId: string;
  state?: string;
}

export interface ConsumerGroupDetails {
  groupId: string;
  state: string;
  coordinator: BrokerInfo;
  members?: ConsumerGroupMemberInfo[];
  topics?: string[];
}

export interface ConsumerGroupMemberInfo {
  clientId: string;
  consumerId: string;
  host: string;
  assignments?: TopicPartitionAssignment[];
}

export interface TopicPartitionAssignment {
  topic: string;
  partition: number;
}
