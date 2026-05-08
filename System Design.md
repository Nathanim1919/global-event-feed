# System Design: Global Event Feed (Senior-Level, 2027)

## 1. Introduction

This document outlines the system design for a `global-event-feed` service, intended as a robust, scalable, and observable backend solution. The goal is to demonstrate a senior engineer's understanding of modern distributed system principles, preparing the project to be a strong portfolio piece by 2027.

The service will ingest, store, and distribute a feed of global events in near real-time. Events can originate from various sources and consumers should be able to subscribe to these events with low latency.

## 2. Goals & Non-Functional Requirements (NFRs)

Beyond basic CRUD for events, this system aims to achieve:

*   **High Availability:** The system should remain operational even with component failures (e.g., database replica failure, single service instance crash).
*   **Scalability:**
    *   **High Ingestion Rate:** Capable of handling hundreds to thousands of events per second.
    *   **High Read Throughput:** Capable of serving thousands of read requests per second, including real-time subscriptions.
    *   **Horizontal Scalability:** Components should be easily scaled by adding more instances.
*   **Low Latency:**
    *   **Ingestion:** Events should be processed and available to consumers within sub-second latency.
    *   **Retrieval:** Event retrieval for recent events should be in milliseconds.
*   **Durability:** No event data loss under normal operating conditions.
*   **Observability:** Comprehensive monitoring, logging, and tracing to quickly diagnose issues.
*   **Security:** Authentication and authorization for API endpoints, secure data handling.
*   **Maintainability:** Clean code, modular design, automated testing, clear documentation.
*   **Cost-Effectiveness:** Design choices should consider operational costs.

## 3. High-Level Architecture

The system will follow a microservices-oriented architecture, leveraging asynchronous communication where appropriate.

```
+----------------+       +-------------------+       +-------------------+       +-------------------+
|  Event Sources |------>|  Ingestion Service  |------>|  Event Processing   |------>|  Event Storage    |
| (e.g., IoT,     |       |   (Go, REST/gRPC) |       |     (Kafka/NATS)    |       | (PostgreSQL,     |
|   APIs, Feeds) |       |                   |       |                     |       |    Redis)         |
+----------------+       +-------------------+       |                     |       |                   |
                                                     |                     |       |  +-------------+  |
                                                     |                     |<------|  | Cache (Redis)|  |
                                                     |                     |       |  +-------------+  |
                                                     +-------------------+       +-------------------+
                                                              |
                                                              | Publish
                                                              V
                                                      +-------------------+
                                                      |  Notification     |
                                                      |   Service         |
                                                      | (WebSockets, SSE) |
                                                      +-------------------+
                                                              |
                                                              V
                                                      +-------------------+
                                                      |  Client           |
                                                      | (Web/Mobile App)  |
                                                      +-------------------+
```

**Core Components:**

*   **Ingestion Service (Go):** Receives raw events from various sources.
*   **Message Broker (Kafka/NATS):** Decouples ingestion from processing and acts as an event bus.
*   **Event Processing Service (Go):** Consumes from the message broker, enriches/validates events, and writes to storage.
*   **Event Storage (PostgreSQL, Redis):** Persistent storage and fast retrieval/caching.
*   **Notification Service (Go):** Subscribes to processed events and pushes them to connected clients (WebSockets/SSE).
*   **API Gateway:** Handles routing, authentication, rate limiting.

## 4. Detailed Component Design

### 4.1. Ingestion Service

*   **Technology:** Go, RESTful API (primary), gRPC (for high-volume/internal sources).
*   **Responsibilities:**
    *   Receiving events from diverse external sources (HTTP POST, potentially gRPC for high-throughput trusted sources).
    *   Basic schema validation of incoming event payloads.
    *   Authentication and rate limiting (potentially handled by an API Gateway).
    *   Publishing raw events to the Message Broker.
*   **Data Model (Ingested Event):**
    ```go
    type IngestedEvent struct {
        SourceID    string    `json:"source_id"` // Identifier for the origin system
        EventType   string    `json:"event_type"`
        Payload     json.RawMessage `json:"payload"` // Raw event data
        IngestedAt  time.Time `json:"ingested_at"` // When received by this service
        CorrelationID string `json:"correlation_id"` // For tracing requests
    }
    ```
*   **Error Handling:**
    *   Synchronous errors (e.g., malformed request, auth failure) returned immediately via HTTP/gRPC.
    *   Asynchronous errors (e.g., Message Broker unavailable): Implement a retry mechanism with a Dead Letter Queue (DLQ) for events that cannot be published after multiple retries.

### 4.2. Message Broker (e.g., Apache Kafka / NATS Streaming)

*   **Choice:**
    *   **Kafka:** High-throughput, durable, persistent, excellent for historical event replay. More complex to operate.
    *   **NATS Streaming:** Simpler, very fast, good for real-time streaming, less focused on long-term persistence/reprocessing.
*   **Purpose:**
    *   Decouples the ingestion path from the processing path.
    *   Provides durability for events before they are stored in the database.
    *   Enables multiple consumers to process the same events independently.
    *   Facilitates event-driven architecture.
*   **Topics/Subjects:**
    *   `events.raw`: For events coming directly from the Ingestion Service.
    *   `events.processed`: For events after enrichment/validation by the Event Processing Service.
*   **Partitioning:** Events should be partitioned (e.g., by `SourceID` or `EventType`) to enable parallel processing and ordering guarantees within a partition.

### 4.3. Event Processing Service

*   **Technology:** Go, consumes from Message Broker.
*   **Responsibilities:**
    *   Consuming events from the `events.raw` topic.
    *   **Validation:** Deeper semantic validation of the event payload.
    *   **Enrichment:** Adding metadata (e.g., geographical context based on `Location`, severity, normalized categories).
    *   **Transformation:** Mapping raw event structures to the canonical `model.Event` structure.
    *   **Persistence:** Writing the canonical `model.Event` to PostgreSQL.
    *   **Caching:** Updating/invalidating relevant caches (Redis) after successful persistence.
    *   Publishing the *processed* event to the `events.processed` topic.
*   **Error Handling:**
    *   Processing errors (e.g., validation failure, database write error): Log extensively, potentially publish to a dedicated `events.dead_letter` topic for manual inspection/reprocessing.
    *   Database connection issues: Implement retry logic with exponential backoff.

### 4.4. Event Storage

#### 4.4.1. Primary Database (PostgreSQL)

*   **Purpose:** Durable, transactional storage for all canonical event data.
*   **Schema (simplified `model.Event` from code):**
    ```sql
    CREATE TABLE events (
        id BIGSERIAL PRIMARY KEY,
        type VARCHAR(50) NOT NULL,
        title VARCHAR(255) NOT NULL,
        description TEXT,
        location VARCHAR(100) NOT NULL,
        timestamp TIMESTAMPTZ NOT NULL, -- When the event occurred
        source_id VARCHAR(100),         -- From IngestedEvent.SourceID
        correlation_id VARCHAR(255),    -- From IngestedEvent.CorrelationID
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- When stored in DB
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

    CREATE INDEX idx_events_timestamp ON events (timestamp DESC);
    CREATE INDEX idx_events_type ON events (type);
    CREATE INDEX idx_events_location ON events (location);
    ```
*   **High Availability:**
    *   **Replication:** Streaming replication (e.g., primary-standby) for read replicas and failover.
    *   **Connection Pooling:** Use `pgxpool` as already implemented.
*   **Scalability:**
    *   **Read Replicas:** Distribute read load to replicas.
    *   **Sharding (Future consideration):** If event volume becomes exceptionally high, sharding by `SourceID`, `EventType`, or time range could be explored.

#### 4.4.2. Caching Layer (Redis)

*   **Purpose:** Reduce load on PostgreSQL and provide ultra-low latency reads for frequently accessed data (e.g., "latest N events", "events by type for last hour").
*   **Data Structures:**
    *   **Sorted Sets:** Store recent events, ordered by timestamp, to efficiently retrieve "latest N events."
    *   **Hashes:** Store individual event details for quick lookup by ID.
*   **Cache Invalidation/Updates:** Handled by the Event Processing Service after a successful write to PostgreSQL.
*   **Time-to-Live (TTL):** Events in Redis can have a shorter TTL than in PostgreSQL, reflecting their transient "hot" nature.

### 4.5. Notification Service

*   **Technology:** Go, WebSockets (for real-time push), Server-Sent Events (SSE) (simpler push, unidirectional).
*   **Responsibilities:**
    *   Consuming processed events from the `events.processed` topic.
    *   Managing client connections (WebSockets/SSE).
    *   Filtering events based on client subscriptions (e.g., "only show events of type 'ALERT' in 'NYC'").
    *   Pushing relevant events to subscribed clients in real-time.
*   **Scalability:** Use a distributed pub/sub mechanism for scaling (e.g., Redis Pub/Sub if not already using Kafka for client-facing subscriptions). A stateful service like this can be challenging to scale, potentially requiring sticky sessions or a shared state layer for connections.

### 4.6. API Gateway (e.g., Nginx, Envoy, AWS API Gateway, Kong)

*   **Purpose:**
    *   **Routing:** Directs incoming requests to the appropriate backend service (Ingestion, Retrieval API, Notification).
    *   **Authentication/Authorization:** Centralized enforcement of security policies.
    *   **Rate Limiting:** Protects backend services from abuse.
    *   **SSL Termination:** Offloads encryption from backend services.
    *   **Load Balancing:** Distributes traffic across multiple instances of backend services.

## 5. Security Considerations

*   **Authentication:** JWT-based authentication for API access. Generate and validate tokens at the API Gateway.
*   **Authorization:** Role-Based Access Control (RBAC) if different users have different permissions (e.g., only admins can ingest certain types of events).
*   **Data Encryption:**
    *   **In Transit:** TLS/SSL for all network communication (API, database connections, message broker).
    *   **At Rest:** Database encryption (handled by cloud provider or OS/filesystem level).
*   **Input Validation:** Strict validation at the ingestion and processing layers to prevent injection attacks and malformed data.
*   **Secrets Management:** Use environment variables for configuration, but for sensitive credentials (DB passwords, API keys), integrate with a secrets manager (e.g., HashiCorp Vault, AWS Secrets Manager).

## 6. Observability

*   **Logging:**
    *   **Structured Logging:** Use a library like `zap` or `logrus` to output JSON-formatted logs.
    *   **Centralized Logging:** Aggregate logs from all services into a central system (e.g., ELK stack, Grafana Loki).
    *   **Log Levels:** Appropriate use of INFO, DEBUG, WARN, ERROR.
*   **Metrics:**
    *   **Prometheus:** Expose application metrics (request rates, error rates, latency, goroutine count, memory usage) via Prometheus endpoints.
    *   **Grafana:** Visualize metrics for dashboards and alerting.
    *   **Key Metrics:** Request/response times, error counts, events processed/published/persisted, database connection pool stats, cache hit/miss rates.
*   **Tracing:**
    *   **OpenTelemetry:** Instrument services to generate traces.
    *   **Jaeger/Zipkin:** Backend for collecting and visualizing traces to understand request flow across services and pinpoint bottlenecks.
    *   **Correlation IDs:** Pass a unique `CorrelationID` through all services for end-to-end request tracking.

## 7. Deployment & Infrastructure

*   **Containerization:** All services will be containerized using Docker.
*   **Orchestration:** Kubernetes (K8s) for deploying, scaling, and managing containers.
    *   **Deployment:** `Deployment` resources for stateless services (Ingestion, Processing, Notification).
    *   **StatefulSet:** For stateful components like PostgreSQL (or use a managed DB service).
    *   **Services:** `Service` resources for internal communication and exposing endpoints.
    *   **Ingress:** For external HTTP/S access to the API Gateway.
    *   **Helm Charts:** Package and deploy the application components easily.
*   **Infrastructure as Code (IaC):** Terraform/Pulumi to define and provision cloud resources (K8s cluster, managed databases, load balancers, message brokers).
*   **CI/CD Pipeline:** GitHub Actions/GitLab CI/Jenkins for automated:
    *   Code linting and static analysis (`go vet`, `golangci-lint`).
    *   Unit and integration tests.
    *   Docker image builds and pushes to a container registry.
    *   Kubernetes deployment updates (canary deployments, rolling updates).

## 8. Development Workflow

*   **Local Development:** Use `docker-compose` to spin up local versions of Kafka/NATS, PostgreSQL, Redis for isolated testing.
*   **Testing:**
    *   **Unit Tests:** For individual functions and components.
    *   **Integration Tests:** Verify interactions between services and external dependencies (DB, message broker).
    *   **End-to-End Tests:** Simulate user flows.
*   **Code Review:** Mandatory code reviews.
*   **Documentation:** Maintain `README.md` for setup, `CONTRIBUTING.md`, and this `System Design.md`.

## 9. Future Enhancements

*   **Event Filtering/Querying:** More advanced API for querying events based on multiple criteria (time range, type, location, keywords).
*   **Analytics Dashboard:** A front-end application to visualize event trends and statistics.
*   **User Preferences:** Allow users to save their event feed preferences.
*   **Machine Learning:** Anomaly detection on event streams, predictive analytics.
*   **Multi-Region Deployment:** For global resilience and lower latency for geographically dispersed users.
*   **GraphQL API:** Provide a flexible API for clients to request exactly the data they need.

---

This design provides a roadmap for building a sophisticated `global-event-feed` system. Implementing these aspects and documenting the challenges and solutions will significantly elevate this project to a senior-level portfolio piece.
