# Ledger Link

A robust financial transaction management system built with Go, featuring secure user management, transaction processing, audit logging, and comprehensive monitoring.

## Features

- User Management with Role-based Access
- Secure Transaction Processing
- Real-time Balance Tracking
- Comprehensive Audit Logging
- MySQL Database with GORM
- Docker Support
- Monitoring Stack (Prometheus & Grafana)
- Metrics Collection and Visualization
- Rate Limiting
- Request Tracing

## Tech Stack

- Go 1.22+
- MySQL 8.0
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
│   │   └── transaction_handler.go
│   ├── services/
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── transaction_service.go
│   │   └── balance_service.go
│   └── models/
│       └── models.go
├── config/
│   ├── grafana/
│   │   └── dashboards/
│   └── prometheus/
│       └── prometheus.yml
├── pkg/
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

# Monitoring
PROMETHEUS_ENABLED=true
TRACING_ENABLED=true
```

## Monitoring Stack

### Prometheus

The application exposes metrics at `/metrics` endpoint, which Prometheus scrapes. Key metrics include:
- Request latencies
- Error rates
- Transaction volumes
- System metrics

### Grafana

Pre-configured dashboards are available for:
- System Overview
- Transaction Metrics
- User Activity
- Error Rates

Default credentials:
- Username: admin
- Password: admin

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
- `POST /api/v1/transactions` - Create transaction
- `GET /api/v1/transactions` - List transactions
- `GET /api/v1/transactions/:id` - Get transaction details

### Balances
- `GET /api/v1/balances/:user_id` - Get user balance
- `POST /api/v1/balances/transfer` - Transfer funds

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
docker-compose up -d mysql prometheus grafana

# Run the application
go run main.go
```

## Docker Support

```bash
# Build and start all services
docker-compose up --build

# Start specific services
docker-compose up mysql prometheus grafana

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
