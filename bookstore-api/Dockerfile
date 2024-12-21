FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git and build tools
RUN apk add --no-cache git build-base

# Install swag
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Generate swagger docs
RUN swag init -g cmd/main.go

# Build the applications
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main ./cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o seed ./scripts/seed.go

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install certificates and wget for healthcheck
RUN apk --no-cache add ca-certificates wget

# Copy binaries and docs from builder
COPY --from=builder /app/main .
COPY --from=builder /app/seed .
COPY --from=builder /app/docs ./docs

# Expose port
EXPOSE 8080

# Set the default command
CMD ["./main"]