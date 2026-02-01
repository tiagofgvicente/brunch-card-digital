# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Install git for dependencies if needed
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod ./
# RUN go mod download # Uncomment when you have external dependencies

# Copy the entire project
COPY . .

# Build the application with optimizations for a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o brunch-api ./cmd/server/main.go

# Stage 2: Final lightweight image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copia o binário
COPY --from=builder /app/brunch-api .
COPY --from=builder /app/web ./web

# GARANTE QUE A PASTA EXISTE ANTES DE COPIAR
RUN mkdir -p internal/database

# Copia o ficheiro SQL da pasta local para dentro da imagem
COPY internal/database/migrations.sql ./internal/database/migrations.sql
COPY --from=builder /app/internal/database ./internal/database

EXPOSE 8080
CMD ["./brunch-api"]