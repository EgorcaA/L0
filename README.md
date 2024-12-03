# L0 Project Overview

This project is designed to integrate Kafka, PostgreSQL, and an in-memory cache to handle and serve order data efficiently. Below is an overview of the components and functionality implemented in this project.
 
Run with:
```
docker compose -f zk-single-kafka-single.yml up
go run ./cmd/app/main.go
docker compose -f zk-single-kafka-single.yml down
```

---

## Features

### 1. PostgreSQL Database
- **Database Setup**: A PostgreSQL database is used to store order data.
- **User Configuration**: A dedicated user is created with appropriate permissions.
- **Tables**: The database schema includes tables specifically designed to store order information received from Kafka.

### 2. Kafka Integration
- **Connection**: The service connects to Kafka brokers.
- **Topic Subscription**: It subscribes to the `orders` topic to consume messages.
- **Message Processing**: Kafka messages containing order data are parsed and stored in the database.

### 3. Caching
- **Redis Cache**: Redis stores recently received order data for quick retrieval.
- **Cache Recovery**: Upon service restart, the cache is repopulated from the database to ensure data consistency.

### 4. HTTP Server
- **Endpoint**: The service includes an HTTP server that exposes an endpoint to fetch order data by ID.
- **Data Source**: The endpoint retrieves data from the in-memory cache for performance.

### 5. User Interface
- **Basic Display**: A simple user interface is provided to display order details by ID.
- **Ease of Use**: Users can input an order ID and view the corresponding data.

---

## How It Works

1. **Data Flow**:
   - Kafka messages are consumed from the `orders` topic.
   - Parsed messages are stored in the PostgreSQL database and cached in memory.
   - The HTTP server serves cached data via a REST endpoint.

2. **Startup Behavior**:
   - On startup, the service restores the in-memory cache from the PostgreSQL database to maintain continuity.

3. **Fault Tolerance**:
   - The service handles errors gracefully, ensuring no data is lost and providing meaningful logs for debugging.

---


