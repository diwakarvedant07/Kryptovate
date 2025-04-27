# Ledger Service API

A robust and scalable backend service for managing customers and transactions. This service provides a RESTful API for handling customer data and transaction processing with MongoDB integration and a transaction queue system.

## Features

- Customer management (CRUD operations)
- Transaction processing with queue system
- MongoDB integration for data persistence
- Swagger documentation
- Health check endpoint
- Comprehensive test coverage

## Prerequisites

- Go 1.16 or higher
- MongoDB (local or Atlas cluster)
- Git

## Setup

1. Clone the repository:

```bash
git clone <repository-url>
cd ledger-service
```

2. Install dependencies:

```bash
go mod download
```

3. Create a `.env` file in the root directory with the following variables:

```
MONGO_CLUSTER=<your-mongodb-connection-string>
```

4. Run the application:

```bash
go run ledger-service.go
```

The service will start on `http://localhost:3005`

## API Documentation

The API documentation is available through Swagger UI at:

```
http://localhost:3005/swagger/index.html
```

### Endpoints

#### Customers

- `POST /customers` - Create a new customer
- `GET /customers` - Get all customers
- `GET /customers/:id` - Get a specific customer
- `PUT /customers/:id` - Update a customer
- `DELETE /customers/:id` - Delete a customer

#### Transactions

- `POST /transactions` - Create a new transaction
- `GET /transactions` - Get all transactions
- `GET /transactions/:id` - Get a specific transaction
- `GET /transactions/customer/:customerId` - Get transactions for a specific customer

#### Health Check

- `GET /health` - Check service health status

## Testing

Run the test suite:

```bash
go test ./...
```

For verbose test output:

```bash
go test -v ./...
```

## Project Structure

```
.
├── handlers/           # API handlers
├── models/            # Data models
├── queue/             # Transaction queue implementation
├── docs/              # Swagger documentation
├── ledger-service.go  # Main application file
└── go.mod             # Go module file
```

## Live Demo

The service is currently running at:

```
http://43.204.217.119/LEDGER/

```

Check service health

```
http://43.204.217.119/LEDGER/health

```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Contact

For any questions or support, please contact:

- Email: dev@kryptovate.com
