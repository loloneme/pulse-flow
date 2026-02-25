# Order Processing System (Event-Driven Architecture in Go)

## Project Goal

Это проект, ориентированный на обучение таким технологиям, как:

* Event-Driven архитектура
* Паттерн Publisher/Subscriber
* Паттерны отказоустойчивости
* Очереди сообщений (Kafka version + in-memory version)
* Idempotency, retries, timeouts
* Worker-based async processing
* ELK logging and observation

---

## System Idea

Система обрабатывает **заказы**, используя асинхронную event-driven коммуникацию


### High-level flow:

1. Клиент отправляет HTTP запрос `POST /orders`
2. Заказ сохраняется в БД (в данном случае PostgreSQL)
3. `OrderCreated` event публикуется
4. Остальные сервисы реагируют:

    * Валидация заказа (асинхронно используя сторонние сервисы, в данном проекте - моки)
    * Произведение оплаты (так же в реальном проекте используя сторонний сервис)

---

## Architecture Overview

### Style:

Event-Driven + Pub/Sub

### Core Components:

* HTTP API
* Order Service (business logic)
* Event Bus (interface abstraction)
* InMemory / Kafka Event Bus имплементация
* Workers (event consumers)
* PostgreSQL (orders storage)
* Redis (идемпотентность/дедупликация)

### Implementations:

* InMemory Bus (для понимания работы очередей сообщений)
* Kafka Bus 

Переключение между ними не требует изменений в бизнес-логике

---

## Failure Scenarios & Resilience Patterns

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

## Workers

Каждый подписчик работает в своей горутине. На обработку каждого события есть свои воркеры или воркер-пул

Ответственность:

* Получить событие
* Обработать бизнес-логику
* Ретрай, если при обращении к сторонним сервисам получен таймаут
* Логирование шагов и ошибок
* Идемпотентность события

---

## Non-Goals

Проект на ориентирован на:

* Фронтенд
* Аутентификацию
* Безопаность продакшн-уровня 

Основаная цель - изучить архитектуру и проектирование системы


---

## Tech Stack

* Go
* PostgreSQL
* Kafka
* Docker
* ELK Stack - Filebeat, ElasticSearch, Kibana