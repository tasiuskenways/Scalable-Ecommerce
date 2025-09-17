# Gemini Project Context: Scalable-Ecommerce

## Project Overview

This is a Go-based microservices project for a scalable e-commerce platform. The architecture is designed to be modular and scalable, with different services responsible for specific business functionalities. The services communicate with each other over an internal network and are managed using Docker Compose. An API Gateway (Kong) is used to expose the services to the public.

### Key Technologies

*   **Backend:** Go
*   **Framework:** Fiber
*   **API Gateway:** Kong
*   **Database:** PostgreSQL
*   **Cache:** Redis
*   **Containerization:** Docker, Docker Compose

### Service Architecture

The project consists of the following microservices:

*   **Kong:** The API Gateway, which acts as a single entry point for all incoming requests and routes them to the appropriate service.
*   **Crypto Service:** Handles cryptographic operations like encryption and decryption.
*   **User Service:** Manages user accounts, authentication, and authorization.
*   **Product Service:** Manages the product catalog.
*   **Shopping Cart Service:** Manages user shopping carts.

Each service (`user-service`, `product-service`, `shopping-cart-service`) has its own dedicated PostgreSQL database and Redis cache, ensuring data isolation and independent scalability.

## Building and Running

The entire application can be built and run using Docker Compose.

**To start all services:**

```bash
docker-compose up -d
```

**To stop all services:**

```bash
docker-compose down
```

**To view logs for a specific service:**

```bash
docker-compose logs -f <service_name>
```

*(TODO: Add instructions for running tests if available.)*

## Development Conventions

The codebase for each microservice follows a conventional layered architecture pattern:

*   **`cmd`:** Contains the main application entry point.
*   **`internal`:** Contains the core application logic, separated into the following layers:
    *   **`config`:** Configuration loading and management.
    *   **`domain`:** Core business logic, including entities, repositories, and services.
    *   **`application`:** Application-level logic, including DTOs and application services.
    *   **`infrastructure`:** Implementation details, such as database access, external service clients, and repositories implementation.
    *   **`interfaces`:** The presentation layer, which handles HTTP requests and responses (e.g., handlers, routes).
*   **`pkg`:** Shared libraries and utilities.

This structure promotes separation of concerns and makes the codebase easier to understand, maintain, and test.
