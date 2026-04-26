IMAGE      := worklog
VHS_IMAGE  := worklog-vhs
GO_IMAGE   := golang:1.26
LINT_IMAGE := golangci/golangci-lint:latest-alpine
DATA_DIR   := $(HOME)/.worklog
# Shared Go module cache so repeat `make test` / `make lint` don't re-download.
GO_CACHE   := worklog-gomod-cache

DOCKER_GO = docker run --rm \
	-v $(PWD):/src -w /src \
	-v $(GO_CACHE):/go/pkg/mod \
	$(GO_IMAGE)

.PHONY: help build lint test ci tidy run serve demo demo-tui clean

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Docker image
	docker build --target app -t $(IMAGE) .

lint: ## Run golangci-lint (same as CI)
	docker run --rm -v $(PWD):/src -w /src $(LINT_IMAGE) golangci-lint run

test: ## Run Go tests with the race detector (same as CI)
	$(DOCKER_GO) go test -race ./...

ci: lint test ## Run everything CI runs

tidy: ## Run go mod tidy
	$(DOCKER_GO) go mod tidy

run: build ## Build and launch the TUI (mounts ~/.worklog)
	@mkdir -p $(DATA_DIR)
	docker run -it --rm --user $$(id -u):$$(id -g) \
		-e TERM=xterm-256color \
		-v $(DATA_DIR):/home/wl/.worklog $(IMAGE)

serve: build ## Build and start the web UI on http://localhost:7171
	@mkdir -p $(DATA_DIR)
	docker run --rm -p 7171:7171 \
		-v $(DATA_DIR):/home/wl/.worklog $(IMAGE) \
		serve --host 0.0.0.0 --no-open

demo: ## Record docs/demo.gif using VHS
	docker build --target vhs -t $(VHS_IMAGE) .
	@mkdir -p docs $(DATA_DIR)
	docker run --rm \
		-v $(PWD):/vhs \
		-v $(DATA_DIR):/root/.worklog \
		$(VHS_IMAGE) docs/demo.tape

demo-tui: ## Record docs/tui.gif using VHS
	docker build --target vhs -t $(VHS_IMAGE) .
	@mkdir -p docs $(DATA_DIR)
	docker run --rm \
		-v $(PWD):/vhs \
		-v $(DATA_DIR):/root/.worklog \
		$(VHS_IMAGE) docs/tui.tape

clean: ## Remove the Docker image and Go module cache volume
	-docker rmi $(IMAGE)
	-docker volume rm $(GO_CACHE)
