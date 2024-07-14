# This could do with a LOT of tidying up :)

APP_NAME := itsoverthere
DOCKER_REPO := teq0v2/$(APP_NAME)

# App version is read from the VERSION file
VERSION := $(shell cat VERSION)

# For now append the git SHA to the version so we can keep each build
SHA := $(shell git rev-parse --short HEAD)

DOCKER_TAG_ARM64 := $(VERSION)-arm64
DOCKER_TAG_AMD64 := $(VERSION)-amd64
DOCKER_TAG_DARWIN_AMD64 := $(VERSION)-darwin-amd64
DOCKER_TAG_DARWIN_ARM := $(VERSION)-darwin-arm64

# Binary build folders
BUILD_FOLDER := build
BUILD_FOLDER_ARM64 := $(BUILD_FOLDER)/arm64
BUILD_FOLDER_AMD64 := $(BUILD_FOLDER)/amd64
BUILD_FOLDER_DARWIN_ARM64 := $(BUILD_FOLDER)/darwin/arm64
BUILD_FOLDER_DARWIN_AMD64 := $(BUILD_FOLDER)/darwin/amd64

# Go build args
BUILD_ARGS_COMMON := CGO_ENABLED=0
BUILD_ARGS_ARM64 := GOOS=linux GOARCH=arm64
BUILD_ARGS_AMD64 := GOOS=linux GOARCH=amd64
BUILD_ARGS_DARWIN_AMD64 := GOOS=darwin GOARCH=amd64
BUILD_ARGS_DARWIN_ARM := GOOS=darwin GOARCH=arm64

DOCKER_ARGS_ARM64 := --build-arg="ARCH=arm64"
DOCKER_ARGS_AMD64 := --build-arg="ARCH=amd64"
DOCKER_ARGS_DARWIN_AMD64 := --build-arg="ARCH=darwin/amd64"

.PHONY: build-folder-arm64 build-folder-darwin-amd64 buld-folder-darwin-arm docker-push-arm64 clean

# Create build folders
build-folder-arm64:
	mkdir -p $(BUILD_FOLDER_ARM64)

build-folder-amd64:
	mkdir -p $(BUILD_FOLDER_AMD64)

build-folder-darwin-amd64:
	mkdir -p $(BUILD_FOLDER_DARWIN_AMD64)

build-folder-darwin-arm:
	mkdir -p $(BUILD_FOLDER_DARWIN_ARM64)

# Build the Go app
build-arm64: build-folder-arm64
	$(BUILD_ARGS_COMMON) $(BUILD_ARGS_ARM64) go build -o $(BUILD_FOLDER_ARM64)/$(APP_NAME) .

build-amd64: build-folder-amd64
	$(BUILD_ARGS_COMMON) $(BUILD_ARGS_ARM64) go build -o $(BUILD_FOLDER_AMD64)/$(APP_NAME) .

build-darwin-amd64: build-folder-darwin-amd64
	$(BUILD_ARGS_COMMON) $(BUILD_ARGS_DARWIN_AMD64) go build -o $(BUILD_FOLDER_DARWIN_AMD64)/$(APP_NAME) .

# Create the Docker images
docker-build-arm64: build-arm64
	docker build $(DOCKER_ARGS_ARM64) -t $(DOCKER_REPO):$(DOCKER_TAG_ARM64) .

docker-build-amd64: build-amd64
	docker build $(DOCKER_ARGS_AMD64) -t $(DOCKER_REPO):$(DOCKER_TAG_AMD64) .

docker-build-darwin-amd64: build-darwin-amd64
	docker build $(DOCKER_ARGS_DARWIN_AMD64) -t $(DOCKER_REPO):$(DOCKER_TAG_DARWIN_AMD64) .

# Push the Docker images to Docker Hub
docker-push-arm64: docker-build-arm64
	docker login --username $(DOCKER_USERNAME) --password $(DOCKER_PASSWORD)
	docker push $(DOCKER_REPO):$(DOCKER_TAG_ARM64)
	docker tag $(DOCKER_REPO):$(DOCKER_TAG_ARM64) $(DOCKER_REPO):latest
	docker push $(DOCKER_REPO):latest

docker-push-darwin-amd64: docker-build-darwin-amd64
	docker login --username $(DOCKER_USERNAME) --password $(DOCKER_PASSWORD)
	docker push $(DOCKER_REPO):$(DOCKER_TAG_DARWIN_AMD64)
	docker tag $(DOCKER_REPO):$(DOCKER_TAG_DARWIN_AMD64) $(DOCKER_REPO):latest
	docker push $(DOCKER_REPO):latest

# Clean up the built Go app and Docker image
clean:
	rm -f $(APP_NAME)
	docker rmi $(DOCKER_REPO):$(DOCKER_TAG)

run-local-amd64: docker-build-amd64
	docker run -p6050:8080 -e DB_HOST=host.docker.internal -e DB_PORT=5432 -e DB_USER=postgres -e DB_PASSWORD=postgres -e DB_NAME=itsoverthere teq0v2/itsoverthere:1.0.0-amd64

certbot:
	sudo certbot certonly --manual --preferred-challenges dns -d '*.itsoverthere.lol' -d 'itsoverthere.lol'

helm-install:
	helm install lol ./helm -f ./helm/values.yaml

helm-upgrade:
	helm upgrade lol ./helm -f ./helm/values.yaml

helm-dry-run:
	helm install itsoverthere ./helm -f ./helm/values.yaml --dry-run --debug 
