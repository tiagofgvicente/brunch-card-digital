# Variables
CLUSTER_NAME=brunch-project
IMAGE_NAME=localhost/brunch-api:v3
TAR_NAME=brunch-api_v3.tar
KIND_CONFIG=kind-config.yaml 
K8S_DIR=deployments/k8s

# Injects Podman as the provider for Kind
export KIND_EXPERIMENTAL_PROVIDER=podman

.PHONY: all cluster build load secrets deploy clean status help nuke update forward

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

all: cluster build load secrets deploy seed forward ## Setup everything including DB data

cluster: ## Create the Kind cluster using Podman
	@echo "🚀 Creating Kind cluster..."
	# Se o cluster já existir, ignora o erro
	kind create cluster --name $(CLUSTER_NAME) || echo "Cluster already exists."

build: ## Build image and save to tar using a safe filename
	@echo "🔨 Building Podman image..."
	podman build -t $(IMAGE_NAME) -f Dockerfile .
	@echo "📦 Saving image to $(TAR_NAME)..."
	podman save $(IMAGE_NAME) -o $(TAR_NAME)

load: ## Load the image from the tar file into Kind
	@echo "🚚 Loading image archive into cluster..."
	kind load image-archive $(TAR_NAME) --name $(CLUSTER_NAME)
	@echo "🧹 Cleaning up tar file..."
	rm -f $(TAR_NAME)

secrets: ## Create Kubernetes secrets from local .env file
	@echo "🔐 Injecting secrets from .env into Kubernetes..."
	@if [ ! -f .env ]; then echo "❌ .env file not found! Please create one."; exit 1; fi
	# O dry-run permite correr este comando várias vezes sem dar erro se já existir
	kubectl create secret generic brunch-secrets --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f -

deploy: secrets ## Apply Kubernetes manifests
	@echo "🚀 Deploying to Kubernetes..."
	kubectl apply -f $(K8S_DIR)/postgres.yaml
	@echo "⏳ Waiting for Postgres Pod to be ready..."
	kubectl wait --for=condition=ready pod -l app=postgres --timeout=120s
	# Dá 5 segundos extra para o Postgres arrancar o processo interno
	@sleep 5 
	kubectl apply -f $(K8S_DIR)/app.yaml
	@echo "✅ App Deployed."

seed: ## 💉 Force inject SQL data directly into DB
	@echo "💉 Seeding Database with initial data..."
	# 1. Lê o ficheiro SQL local
	# 2. Encontra o pod do Postgres
	# 3. Executa o comando psql lá dentro com o utilizador 'admin' e base 'loyalty_db'
	cat internal/database/migrations.sql | kubectl exec -i $$(kubectl get pods -l app=postgres -o name) -- psql -U admin -d loyalty_db
	
	@echo "🔍 Verifying Store Existence..."
	# Faz um SELECT para te provar no terminal que a loja foi criada
	kubectl exec -i $$(kubectl get pods -l app=postgres -o name) -- psql -U admin -d loyalty_db -c "SELECT name, slug FROM stores;"

update: build load ## Update the app with new code without deleting the cluster
	@echo "🔄 Updating deployment..."
	kubectl rollout restart deployment brunch-api-deployment
	@echo "⏳ Waiting for rollout to finish..."
	kubectl rollout status deployment brunch-api-deployment

status: ## Check the status of the nodes and pods
	@echo "📊 Cluster Status:"
	kubectl get nodes
	kubectl get pods
	kubectl get services

forward:
	# Run in background and redirect output to avoid cluttering the terminal
	kubectl port-forward svc/brunch-api-service 8080:8080 > /dev/null 2>&1 &
	@echo "✅ Forwarding started in background. Access at http://localhost:8080"

clean: ## Delete the cluster and cleanup local files
	@echo "🗑️ Deleting cluster..."
	kind delete cluster --name $(CLUSTER_NAME)
	rm -f $(TAR_NAME)

nuke: clean ## Remove everything: cluster, images, and podman cache
	@echo "☢️ Deep cleaning Podman cache..."
	podman builder prune -f
	podman rmi $(IMAGE_NAME) --force || true
	@echo "✨ System is clean."