#!/bin/sh

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Generate swagger documentation
echo "Generating Swagger documentation..."
swag init -g cmd/main.go

echo "Swagger documentation generated successfully!"
echo "You can access the documentation at http://localhost:8080/swagger/index.html when the server is running."