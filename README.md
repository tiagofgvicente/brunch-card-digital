# Brunch Card Digital

## Overview
Brunch Card Digital is a digital loyalty card application designed to enhance customer engagement through a rewards system. Customers can earn stamps for their purchases and redeem rewards once they reach a certain threshold.

## Features

## Technologies Used

## Getting Started
1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/brunch-card-digital.git
   cd brunch-card-digital
   ```
2. Build the Docker image:
   ```bash
   make build
   ```
3. Deploy to Kubernetes:
   ```bash
   make deploy
   ```

## API Endpoints

## License
This project is licensed under the MIT License.
# Brunch Card Digital

Brunch Card Digital is a small Go-based backend that implements a digital loyalty card system. It provides APIs to create and manage customer cards, generates QR codes for cards, and persists data in PostgreSQL. The project includes Docker and Kubernetes manifests for local development and testing (Kind + Podman recommended).

## Quick summary
- Language: Go (module: `brunch-card-digital`)
- Database: PostgreSQL
- QR generation: `github.com/skip2/go-qrcode`
- Packaging: Docker
- Orchestration: Kubernetes (manifests in `deployments/k8s`)

## Repository layout

Top-level layout (important files/folders):

- `cmd/server/main.go` — application entrypoint and HTTP server bootstrap
- `internal/api/` — HTTP handlers and API logic (`handlers.go`, `qrcode.go`, `status.go`, `stamps.go`)
- `internal/database/` — DB connection and repository code (`postgres.go`, `repository.go`, `migrations.sql`)
- `internal/models/` — domain models for cards
- `models/` — templates and card model helpers
- `web/` — static frontend templates (`index.html`, `card.html`)
- `deployments/` — Kubernetes/Kind configuration and manifests (`kind-config.yaml`, `k8s/`)
- `Dockerfile`, `Makefile`, `go.mod`, `README.md`

Example tree (truncated):

```
Dockerfile
go.mod
Makefile
cmd/
   server/main.go
internal/
   api/
      handlers.go
      qrcode.go
      stamps.go
      status.go
   database/
      postgres.go
      repository.go
      migrations.sql
   models/
      card.go
      templates.go
web/
   index.html
   card.html
deployments/
   kind-config.yaml
   k8s/
      app-configmap.yaml
      app-deploy.yaml
      postgres-db.yaml
```

## Prerequisites
- Go 1.24
- Docker or Podman (for building images)
- Kind + Podman (optional) for local Kubernetes testing
- A running PostgreSQL instance (or use the `deployments/k8s/postgres-db.yaml` manifest)

## Build & Run (local)

Build and run the binary locally:

```bash
# from repo root
go build -o brunch-api ./cmd/server/main.go
./brunch-api
```

The server expects PostgreSQL to be reachable. Defaults used in `cmd/server/main.go` (for Kubernetes local deployment):

- host: `postgres-service`
- user: `admin`
- password: `brunch_pass`
- dbname: `loyalty_db`

You can override these values by modifying `cmd/server/main.go` or wiring environment variables and a small wrapper script.

## Build Docker image

Use the `Makefile` targets (see `Makefile`) or build manually:

```bash
# builds image (Makefile target)
make build

# or with podman/docker
docker build -t localhost/brunch-api:v1 .
```

## Deploy to Kubernetes (Kind + Podman)

The repository includes Kubernetes manifests in `deployments/k8s`. A simple flow using the provided `Makefile` is:

```bash
make cluster   # create kind cluster (podman provider)
make load      # load the image into the cluster
make deploy    # apply k8s manifests
```

If you prefer to run only the Postgres manifest for local development:

```bash
kubectl apply -f deployments/k8s/postgres-db.yaml
```

## API

Create a card (example):

```bash
curl -X POST http://localhost:8080/api/v1/cards \
   -H "Content-Type: application/json" \
   -d '{"customer_id": "tester_01", "design": "retro"}'
```

Get a QR code for a card (PNG response):

```bash
curl "http://localhost:8080/api/v1/qrcode?id=<card-uuid>" --output card.png
```

API endpoints (overview):

- `POST /api/v1/cards` — create a new loyalty card
- `GET  /api/v1/qrcode?id={cardID}` — generate QR code PNG for card
- Other handlers are implemented under `internal/api` for stamps and status

## Development notes
- Database migrations are in `internal/database/migrations.sql`.
- Repository code is in `internal/database/repository.go` and uses `database/sql` with the `lib/pq` driver.
- QR generation uses `github.com/skip2/go-qrcode`.

## Contributing
Open issues or PRs with improvements or bug fixes. Keep changes focused and include tests where appropriate.

## License
MIT
