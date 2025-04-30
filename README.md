# Maestro - Apache Kafka Management UI

Maestro is a modern, lightweight management interface for Apache Kafka clusters. It provides a simple and intuitive way to monitor and manage Kafka topics, consumer groups, and brokers through a responsive web interface.

<img alt="Maestro UI" src="./frontend/src/logos/maestro.png" width="300" />

## Features

- **Cluster Overview**: View broker information and cluster health at a glance
- **Topic Management**: Create, configure, delete, and browse Kafka topics
- **Consumer Group Monitoring**: Track consumer groups, their members, and assignments
- **Configuration Management**: Modify topic configurations with a user-friendly interface
- **Message Explorer**: Browse and inspect messages within Kafka topics
- **Responsive Design**: Modern UI that works on desktop and mobile devices

## Architecture

Maestro follows a client-server architecture with two main components:

- **Backend**: A Go service that communicates with Kafka using the Confluent Kafka Go client
- **Frontend**: A React application built with TypeScript, Tailwind CSS, and Vite

The backend exposes a RESTful API that allows the frontend to query and manage Kafka resources. This separation of concerns makes Maestro extensible and maintainable.

## Getting Started

### Prerequisites

- Docker and Docker Compose (for the quickest setup)
- Go 1.22+ (for backend development)
- Node.js 20+ (for frontend development)
- Apache Kafka 3.0+ cluster

Quick Start with Docker Compose
The easiest way to run Maestro with a local Kafka cluster for testing:

```shell
# Clone the repository
git clone https://github.com/valeriouberti/maestro.git
cd maestro

# Start Kafka Cluster
docker-compose up -d
```

Once running, navigate to http://localhost:5173 to access the Maestro UI.

### Manual Setup

#### Backend Setup

```shell
# Navigate to backend directory
cd backend

# Set required environment variables
export KAFKA_BROKERS=localhost:9092

# Optional environment variables (with defaults shown)
# export PORT=8080
# export READ_TIMEOUT=5s
# export WRITE_TIMEOUT=10s
# export KAFKA_TIMEOUT=5s
# export LOG_LEVEL=info
# export ENABLE_TLS=false
# export CERT_FILE=
# export KEY_FILE=
# export ENVIRONMENT=development

# Run the backend
go run cmd/maestro/main.go
```

#### Frontend Setup

```shell
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Update API configuration if needed
# Edit src/apiConfig.ts to point to your backend

# Start development server
npm run dev
```

# Api Documentation

The backend exposes a RESTful API with the following endpoints:

#### Cluster Operations

- `GET /api/v1/clusters` - List all brokers in the cluster

#### Topic Operations

- `GET /api/v1/topics` - List all topics
- `GET /api/v1/topics/:topicName` - Get details for a specific topic
- `POST /api/v1/topics` - Create a new topic
- `DELETE /api/v1/topics/:topicName` - Delete a topic
- `PUT /api/v1/topics/:topicName/config` - Update topic configuration
- `GET /api/v1/topics/:topicName/messages` - Retrieve messages from a topic

#### Message Exploration

- `GET /api/v1/topics/:topicName/messages` - Retrieve messages from a topic
  - Query parameters:
    - `partition` - Partition to read from (default: 0)
    - `offset` - Starting offset (default: beginning, use "latest" for newest messages)
    - `limit` - Maximum number of messages to retrieve (default: 100)

#### Consumer Group Operations

- `GET /api/v1/consumer-groups` - List all consumer groups
- `GET /api/v1/consumer-groups/:groupId` - Get details for a specific consumer group

## Configuration

#### Backend Configuration

The backend is configured using environment variables:

| Variable      | Description                              | Default                   |
| ------------- | ---------------------------------------- | ------------------------- |
| KAFKA_BROKERS | Comma-separated list of Kafka brokers    | (required)                |
| PORT          | HTTP server port                         | 8080                      |
| READ_TIMEOUT  | HTTP read timeout                        | 5s                        |
| WRITE_TIMEOUT | HTTP write timeout                       | 10s                       |
| KAFKA_TIMEOUT | Kafka operations timeout                 | 5s                        |
| LOG_LEVEL     | Logging level (debug, info, warn, error) | info                      |
| ENABLE_TLS    | Enable HTTPS                             | false                     |
| CERT_FILE     | TLS certificate file path                | (required if TLS enabled) |
| KEY_FILE      | TLS key file path                        | (required if TLS enabled) |
| ENVIRONMENT   | Environment name                         | development               |

#### Frontend Configuration

The frontend uses Vite's environment variables system for configuration:

| Variable          | Description              | Default                      |
| ----------------- | ------------------------ | ---------------------------- |
| VITE_API_BASE_URL | Base URL for backend API | http://localhost:8080/api/v1 |

Environment-specific configuration files:

- `.env` - Loaded in all environments
- `.env.development` - Development environment only
- `.env.production` - Production environment only

Example `.env` file:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

# Development

#### Backend Development

The backend is structured following Go best practices:

```
backend/
├── cmd/
│   └── maestro/          # Application entry point
├── internal/
│   ├── config/           # Configuration management
│   └── kafka/            # Kafka client implementation
├── pkg/
│   ├── api/              # HTTP handlers and routing
│   └── domain/           # Domain models and interfaces
└── tests/
    └── unit/             # Unit tests
```

#### Frontend Development

The frontend follows a component-based architecture using React:

```
frontend/
├── public/               # Static assets
└── src/
    ├── components/       # React components
    ├── App.tsx           # Application root component
    ├── apiConfig.ts      # API configuration
    ├── types.ts          # TypeScript type definitions
    └── main.tsx          # Application entry point
```

### Future improvements

#### Backend

- <input disabled="" type="checkbox"> Add authentication and authorization
- <input disabled="" type="checkbox"> Add schema registry integration
- <input disabled="" type="checkbox"> Support for Kafka Connect management
- <input disabled="" type="checkbox"> Enhanced broker management capabilities
- <input disabled="" type="checkbox"> Metrics collection and visualization
- <input disabled="" type="checkbox"> Support for SASL/SCRAM and SSL authentication methods
- <input disabled="" type="checkbox"> ACL management
- <input disabled="" type="checkbox"> Integration with multiple Kafka clusters
- <input disabled="" type="checkbox"> Prometheus metrics export

#### Frontend

- <input disabled="" type="checkbox"> Dark mode support
- <input disabled="" type="checkbox"> User preferences and settings
- <input disabled="" type="checkbox"> Interactive topic data visualization
- <input disabled="" type="checkbox"> Message publishing interface
- <input disabled="" type="checkbox"> Real-time updates using WebSockets
- <input disabled="" type="checkbox"> Role-based access control UI
- <input disabled="" type="checkbox"> Improved mobile experience
- <input disabled="" type="checkbox"> Advanced filtering and searching
- <input disabled="" type="checkbox"> Export functionality (CSV, JSON)
- <input disabled="" type="checkbox"> Topic data browsing interface

# Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

# License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

# Acknowledgements

- [Confluent Kafka Go](https://github.com/confluentinc/confluent-kafka-go) - Kafka client for Go
- [Gin Web Framework](https://github.com/gin-gonic/gin) - HTTP web framework for Go
- [React](https://react.dev/) - JavaScript library for building user interfaces
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework
- [Vite](https://vite.dev/) - Next generation frontend tooling

---

Created and maintained by [Valerio Uberti](https://github.com/valeriouberti)
