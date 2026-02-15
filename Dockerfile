# Stage 1: Build the Go binary
FROM golang:1.24.3-alpine AS builder

# Install git for dependencies if needed
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o brunch-api ./cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/brunch-api .

COPY --from=builder /app/web ./web

RUN mkdir -p internal/database
COPY --from=builder /app/internal/database/migrations.sql ./internal/database/migrations.sql

EXPOSE 8080

CMD ["./brunch-api"]