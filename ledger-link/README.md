# Ledger Link

A robust financial transaction management system built with Go, featuring secure user management, transaction processing, and audit logging.

## Features

- User Management with Role-based Access
- Secure Transaction Processing
- Real-time Balance Tracking
- Comprehensive Audit Logging
- MySQL Database with GORM
- Docker Support

## Tech Stack

- Go 1.22+
- MySQL 8.0
- GORM (ORM)
- Docker & Docker Compose
- golang-migrate (Database Migrations)

## Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose
- Make (optional)

## Project Structure

```
ledger-link/
├── cmd/
├── internal/
│   ├── database/
│   │   ├── migrations/
│   │   ├── config.go
│   │   └── migrate.go
│   └── models/
│       └── models.go
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

3. Start the MySQL database using Docker:

```bash
docker-compose up -d
```

4. Run the application:

```bash
go run main.go
```

## Database Setup

The application uses MySQL as its database. You can run it in two ways:

### Using Docker (Recommended)

```bash
# Start MySQL container
docker-compose up -d

# Check container status
docker-compose ps
```

### Local MySQL

```bash
# Create database
mysql -u root -p
mysql> CREATE DATABASE ledger_link CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## Environment Variables

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=ledger_user
DB_PASSWORD=ledger_pass
DB_NAME=ledger_link
```

## Database Schema

### Users

- ID (Primary Key)
- Username (Unique)
- Email (Unique)
- Password Hash
- Role
- Created/Updated At

### Transactions

- ID (Primary Key)
- From User ID (Foreign Key)
- To User ID (Foreign Key)
- Amount
- Type (transfer/deposit/withdrawal)
- Status (pending/completed/failed)
- Created At

### Balances

- User ID (Primary Key)
- Amount
- Last Updated At

### Audit Logs

- ID (Primary Key)
- Entity Type
- Entity ID
- Action
- Details
- Created At

## Development

### Running Tests

```bash
go test ./...
```

### Database Migrations

Migrations are automatically handled by GORM auto-migrate feature when the application starts.

### Batch Processing System

The project includes a concurrent batch processing system for handling multiple tasks efficiently. Here's how to use it:

```go
// Create a new batch processor
processor := batch.NewBatchProcessor(batch.Config{
    WorkerCount: 5,    // Number of concurrent workers
    QueueSize:   100,  // Size of the task queue
    Logger:      logger,
})

// Start the processor
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
processor.Start(ctx)

// Submit tasks
task := NewCustomTask()
err := processor.Submit(task)
if err != nil {
    log.Printf("Failed to submit task: %v", err)
}

// Get processing statistics
stats := processor.GetStats()

// Graceful shutdown
processor.Stop()
```

#### Implementing Custom Tasks

To create a custom task, implement the `Task` interface:

```go
// Task interface
type Task interface {
    Process(ctx context.Context) error
    ID() string
}

// Example implementation
type CustomTask struct {
    id       string
    data     interface{}
    duration time.Duration
}

func NewCustomTask(id string, data interface{}, duration time.Duration) *CustomTask {
    return &CustomTask{
        id:       id,
        data:     data,
        duration: duration,
    }
}

func (t *CustomTask) Process(ctx context.Context) error {
    select {
    case <-time.After(t.duration):
        // Your task processing logic here
        return nil
    case <-ctx.Done():
        return fmt.Errorf("task cancelled: %w", ctx.Err())
    }
}

func (t *CustomTask) ID() string {
    return t.id
}
```

Key features of the batch processor:

- Concurrent processing with configurable worker pool
- Built-in statistics tracking
- Graceful shutdown support
- Context-aware task processing
- Generic interface for custom task implementations

## Docker Support

Build and run the entire application stack using Docker Compose:

```bash
# Build and start all services
docker-compose up --build

# Stop all services
docker-compose down

# View logs
docker-compose logs -f
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
