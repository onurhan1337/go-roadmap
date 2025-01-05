# Ledger Link

A robust financial transaction management system built with Go, featuring secure user management, transaction processing, audit logging, and comprehensive monitoring.

## Features

- User Management with Role-based Access
- Secure Transaction Processing with Atomic Operations
- Real-time Balance Tracking with Caching
- Comprehensive Audit Logging
- MySQL Database with GORM
- Redis Caching for Performance
- Docker Support
- Monitoring Stack (Prometheus & Grafana)
- Metrics Collection and Visualization
- Rate Limiting
- Request Tracing
- Transaction History Tracking
- Balance History

## Tech Stack

- Go 1.22+
- MySQL 8.0
- Redis (Caching)
- GORM (ORM)
- Docker & Docker Compose
- golang-migrate (Database Migrations)
- Prometheus (Metrics)
- Grafana (Visualization)
- OpenTelemetry (Tracing)

## Project Structure

```
ledger-link/
├── cmd/
├── internal/
│   ├── database/
│   │   ├── migrations/
│   │   ├── config.go
│   │   └── migrate.go
│   ├── handlers/
│   │   ├── user_handler.go
│   │   ├── transaction_handler.go
│   │   └── balance_handler.go
│   ├── services/
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── transaction_service.go
│   │   └── balance_service.go
│   ├── processor/
│   │   └── transaction_processor.go
│   └── models/
│       ├── models.go
│       ├── balance.go
│       ├── transaction.go
│       └── interfaces.go
├── config/
│   ├── grafana/
│   │   └── dashboards/
│   └── prometheus/
│       └── prometheus.yml
├── pkg/
│   ├── cache/
│   ├── auth/
│   ├── logger/
│   └── middleware/
├── docker-compose.yml
├── .env
└── main.go
```

## Quick Start

1. Clone the repository:

```bash
git clone <repository-url>
cd ledger-link
```

2. Copy the example environment file:

```bash
cp .env.example .env
```

3. Start the services using Docker:

```bash
docker-compose up -d
```

4. Access the services:
- Application API: http://localhost:8080
- Grafana Dashboard: http://localhost:3000
- Prometheus: http://localhost:9090

## Environment Variables

```env
# Application
APP_PORT=8080
APP_ENV=development
LOG_LEVEL=debug

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=ledger_user
DB_PASSWORD=ledger_pass
DB_NAME=ledger_link

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Monitoring
PROMETHEUS_ENABLED=true
TRACING_ENABLED=true
```

## Transaction Processing

The system implements a robust transaction processing mechanism with the following features:

### Balance Management
- Atomic operations for balance updates
- Optimistic locking for concurrent transactions
- Cache invalidation on balance updates
- Balance history tracking
- 5-minute cache TTL for read operations

### Transfer Process
1. Create pending transaction
2. Lock sender and receiver balances
3. Verify sufficient funds
4. Update sender balance
5. Update receiver balance
6. Record balance history
7. Update transaction status
8. Release locks

### Error Handling
- Automatic rollback on failed transfers
- Detailed error logging
- Transaction status tracking
- Audit trail for all operations

## API Endpoints

### User Management
- `POST /api/v1/users` - Create user
- `GET /api/v1/users` - List users
- `GET /api/v1/users/:id` - Get user details
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh token

### Transactions
- `POST /api/v1/transactions/transfer` - Transfer funds
- `POST /api/v1/transactions/deposit` - Deposit funds
- `POST /api/v1/transactions/withdraw` - Withdraw funds
- `GET /api/v1/transactions` - List transactions
- `GET /api/v1/transactions/:id` - Get transaction details

### Balances
- `GET /api/v1/balances/current` - Get current balance
- `GET /api/v1/balances/history` - Get balance history

## Monitoring Stack

### Prometheus Metrics
- Transaction counts by type and status
- Balance operation metrics
- Cache hit/miss ratios
- Request latencies
- Error rates
- Active transactions

### Grafana Dashboards
- System Overview
- Transaction Metrics
- Balance Operations
- Cache Performance
- Error Rates
- User Activity

Default credentials:
- Username: admin
- Password: admin

## Development

### Running Tests

```bash
go test ./... -v
```

### Database Migrations

```bash
# Run migrations
go run cmd/migrate/main.go up

# Rollback migrations
go run cmd/migrate/main.go down
```

### Local Development

```bash
# Start dependencies
docker-compose up -d mysql redis prometheus grafana

# Run the application
go run main.go
```

### Docker Commands

```bash
# Build and start all services
docker-compose up --build -d

# Start specific services
docker-compose up mysql redis prometheus grafana

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Clean up unused resources
docker system prune -a --volumes
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request