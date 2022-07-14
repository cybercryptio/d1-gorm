##### Help message #####
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target> \033[36m\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

##### Build targets #####
.PHONY: build
build: ## Build the d1gorm library
	go build -v ./...

.PHONY: lint
lint: ## Lint the codebase
	gofmt -l -w .
	go mod tidy
	golangci-lint run -E gosec,asciicheck,bodyclose,gocyclo,unconvert,gocognit,misspell,revive,whitespace --timeout 5m

##### Test #####
.PHONY: unit-tests
unit-tests: build  ## Run unit tests
	go test -count=1 ./...

.PHONY: coverage
coverage: build  ## Generate coverage report
	go test -count=1 -coverprofile coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

##### Cleanup targets #####
.PHONY: clean  ## Remove build artifacts
clean :
	rm -f coverage.out coverage.html
