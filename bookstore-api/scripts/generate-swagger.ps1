# Add Go bin to PATH
$env:Path += ";$env:USERPROFILE\go\bin"

# Check if swag is installed
if (!(Get-Command swag -ErrorAction SilentlyContinue)) {
    Write-Host "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest

    # Refresh PATH after installation
    $env:Path += ";$env:USERPROFILE\go\bin"

    # Verify installation
    if (!(Get-Command swag -ErrorAction SilentlyContinue)) {
        Write-Host "Error: Failed to install swag. Please install it manually using: go install github.com/swaggo/swag/cmd/swag@latest" -ForegroundColor Red
        exit 1
    }
}

# Generate swagger documentation
Write-Host "Generating Swagger documentation..."
try {
    & swag init -g cmd/main.go
    if ($LASTEXITCODE -ne 0) {
        throw "Swagger generation failed"
    }
    Write-Host "Swagger documentation generated successfully!" -ForegroundColor Green
    Write-Host "You can access the documentation at http://localhost:8080/swagger/index.html when the server is running." -ForegroundColor Cyan
} catch {
    Write-Host "Error generating Swagger documentation: $_" -ForegroundColor Red
    exit 1
}