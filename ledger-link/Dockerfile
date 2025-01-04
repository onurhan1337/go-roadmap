FROM golang:1.22-alpine

# Install curl for healthcheck
RUN apk add --no-cache curl

WORKDIR /app

# Set GOPROXY with fallbacks and disable module verification
ENV GOPROXY=https://proxy.golang.org,direct \
    GOSUMDB=off \
    GO111MODULE=on

COPY go.mod go.sum ./
RUN go mod download || go mod download || go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080

CMD ["./main"] 