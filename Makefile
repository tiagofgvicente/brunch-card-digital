# Variables
CLUSTER_NAME=brunch-project
IMAGE_NAME=localhost/brunch-api:v1
# We use an underscore here to avoid the Podman ":" filename error
TAR_NAME=brunch-api_v1.tar
KIND_CONFIG=deployments/kind-config.yaml
K8S_DIR=deployments/k8s

# Injects Podman as the provider for Kind
export KIND_EXPERIMENTAL_PROVIDER=podman

.PHONY: all cluster build load deploy clean status help

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

all: cluster build load deploy ## Setup everything from scratch

cluster: ## Create the Kind cluster using Podman
	@echo "Creating Kind cluster..."
	kind create cluster --name $(CLUSTER_NAME) --config $(KIND_CONFIG)

build: ## Build image and save to tar using a safe filename
	@echo "Building Podman image..."
	podman build -t $(IMAGE_NAME) -f Dockerfile .
	@echo "Saving image to $(TAR_NAME)..."
	podman save $(IMAGE_NAME) -o $(TAR_NAME)

load: ## Load the image from the tar file into Kind
	@echo "Loading image archive into cluster..."
	kind load image-archive $(TAR_NAME) --name $(CLUSTER_NAME)
	@echo "Cleaning up tar file..."
	rm $(TAR_NAME)

deploy: ## Apply Kubernetes manifests
	@echo "Deploying to Kubernetes..."
	kubectl apply -f $(K8S_DIR)/postgres-db.yaml
	@echo "Waiting for Postgres to be ready..."
	kubectl wait --for=condition=ready pod -l app=postgres --timeout=90s
	kubectl apply -f $(K8S_DIR)/app-deploy.yaml

status: ## Check the status of the nodes and pods
	kubectl get nodes
	kubectl get pods

clean: ## Delete the cluster and cleanup local files
	kind delete cluster --name $(CLUSTER_NAME)
	rm -f *.tar

.PHONY: nuke
nuke: clean ## Remove everything: cluster, images, and podman cache
	@echo "Deep cleaning Podman cache..."
	podman builder prune -f
	podman rmi $(IMAGE_NAME) --force || true
	@echo "System is clean. Ready for 'make all'."