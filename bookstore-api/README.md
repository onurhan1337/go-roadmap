# 📚 Bookstore API

A modern RESTful API for managing a bookstore's inventory, built with Go, Gin, GORM, and PostgreSQL.

## 🌟 Features

- **RESTful API Endpoints** for book management (CRUD operations)
- **PostgreSQL Database** with GORM ORM
- **Swagger Documentation** for API exploration
- **Docker Support** for easy deployment
- **ISBN Validation** for book entries
- **Structured Project Layout**
- **Health Check Endpoint**

## 🛠️ Tech Stack

- **[Go](https://golang.org/)** - Programming language
- **[Gin](https://gin-gonic.com/)** - Web framework
- **[GORM](https://gorm.io/)** - ORM library
- **[PostgreSQL](https://www.postgresql.org/)** - Database
- **[Swagger](https://swagger.io/)** - API documentation
- **[Docker](https://www.docker.com/)** - Containerization

## 🚀 Quick Start

### Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose
- PostgreSQL (if running locally)

### Running with Docker

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/bookstore-api.git
   cd bookstore-api
   ```

2. Start the services:

   ```bash
   docker-compose up --build
   ```

3. Seed the database (optional):
   ```bash
   docker-compose --profile seeder up
   ```

The API will be available at http://localhost:8080

### Running Locally

1. Install dependencies:

   ```bash
   go mod download
   ```

2. Set up environment variables:

   ```bash
   export DB_HOST=localhost
   export DB_USER=postgres
   export DB_PASSWORD=postgres
   export DB_NAME=bookstore
   export DB_PORT=5432
   ```

3. Generate Swagger documentation:

   ```bash
   # Windows
   .\scripts\generate-swagger.ps1

   # Linux/Mac
   ./scripts/generate-swagger.sh
   ```

4. Run the application:
   ```bash
   go run cmd/main.go
   ```

## 📖 API Documentation

Access the Swagger documentation at: http://localhost:8080/swagger/index.html

### Available Endpoints

- `GET /api/v1/books` - List all books
- `GET /api/v1/books/:isbn` - Get book by ISBN
- `POST /api/v1/books` - Create a new book
- `PUT /api/v1/books/:isbn` - Update a book
- `DELETE /api/v1/books/:isbn` - Delete a book
- `GET /ping` - Health check

## 📁 Project Structure

```
bookstore-api/
├── cmd/
│   └── main.go                 # Application entry point
├── config/
│   └── database.go            # Database configuration
├── controllers/
│   └── book_controller.go     # Request handlers
├── models/
│   └── book.go               # Data models
├── pkg/
│   └── validator/
│       └── validator.go      # Custom validators
├── internal/
│   └── router/
│       └── router.go         # Router setup
├── scripts/
│   ├── generate-swagger.ps1  # Swagger generation (Windows)
│   ├── generate-swagger.sh   # Swagger generation (Unix)
│   └── seed.go              # Database seeding
├── docs/                     # Swagger documentation
├── Dockerfile               # Docker configuration
└── docker-compose.yml      # Docker Compose configuration
```

## 🧪 Testing

soon.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 🔮 Future Enhancements

- Authentication & Authorization
- Advanced Search & Filtering
- Book Categories/Genres
- User Reviews & Ratings
- Caching Layer
- Monitoring & Logging
