# Order Processing System (Event-Driven Architecture in Go)

## 📌 Project Goal

This is a learning-oriented backend project focused on:

* Event-Driven Architecture
* Pub/Sub messaging
* Resilient system design
* Message queues (Kafka version + in-memory version)
* Idempotency, retries, timeouts
* Worker-based async processing

The goal is to deeply understand how distributed systems work internally, not just "use Kafka".

---

## 🧠 System Idea

The system processes **orders** using asynchronous event-driven communication.

### High-level flow:

1. Client sends `POST /orders`
2. Order is saved in DB
3. `OrderCreated` event is published
4. Other services react asynchronously:

    * Payment processing
    * Notification service
    * Analytics
    * Order state updates

The services **do not call each other directly**.
They only communicate through events.

---

## 🏗 Architecture Overview

### Style:

Event-Driven + Pub/Sub

### Core Components:

* HTTP API
* Order Service (business logic)
* Event Bus (interface abstraction)
* InMemory Event Bus implementation
* Kafka Event Bus implementation (later)
* Workers (event consumers)
* PostgreSQL (orders storage)
* Optional: Redis (idempotency/deduplication)

### Implementations:

* ✅ InMemory Bus (for local dev & understanding flow)
* 🚧 Kafka Bus (distributed version)

Switching between them should not require changes in business logic.

---

## 🔁 Failure Scenarios & Resilience Patterns (Planned)

The project aims to implement and experiment with:

* Retries with backoff
* Dead Letter Queue
* Idempotent handlers
* At-least-once delivery simulation
* Message deduplication
* Timeout handling
* Graceful shutdown of workers
* Context cancellation

---

## 🧵 Workers

Each subscriber runs in a worker goroutine.

Responsibilities:

* Receive event
* Process business logic
* Handle retries
* Log errors
* Ensure idempotency

---

## 🗄 Database

PostgreSQL stores:

### Orders table

* id
* user_id
* amount
* status (created / paid / failed / canceled)
* created_at

Optional:

* processed_events table (for idempotency)

---

## 🚀 Planned Roadmap

### Phase 1 — Core Flow (InMemory Only)

* [ ] Basic HTTP server
* [ ] Order creation
* [ ] InMemory Event Bus
* [ ] Payment worker
* [ ] Order status updates

### Phase 2 — Resilience

* [ ] Retry mechanism
* [ ] Idempotency support
* [ ] Dead-letter simulation
* [ ] Graceful shutdown

### Phase 3 — Kafka Integration

* [ ] Kafka-based EventBus
* [ ] Topic per event type
* [ ] Consumer groups
* [ ] Offset management

### Phase 4 — Production-Like Improvements

* [ ] Structured logging
* [ ] Observability
* [ ] Metrics
* [ ] Tracing
* [ ] Config via env

---

## 🎯 Learning Objectives

By building this project, I aim to deeply understand:

* How pub/sub really works internally
* How worker pools operate in Go
* Concurrency primitives (channels, goroutines)
* Delivery guarantees
* Idempotency strategies
* Decoupling via interfaces
* Clean architecture in Go
* How Kafka differs from in-memory messaging

---

## ⚠️ Non-Goals

This project is not focused on:

* Frontend
* Authentication
* Production-level security

The main purpose is architectural and system design exploration.

---

## 🛠 Tech Stack

* Go
* PostgreSQL
* Kafka (later stage)
* Docker (planned)
* Docker Compose (planned)
