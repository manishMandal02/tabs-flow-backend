# TypeScript Linting
lint-ts:
	npm run lint

# Go Linting
lint-go:
	golangci-lint run ./...

# All Linting
lint: lint-ts lint-go

# Build Go services
build-go:
	go build -o bin/app ./cmd/app/main.go

# CDK Deployment
deploy-cdk:
	npx cdk deploy

# Clean up
clean:
	rm -rf ./bin

# Install Husky for Pre-commit
install-hooks:
	npx husky install

# Run Go application
run-go:
	go run ./cmd/app/main.go

# Default command
default: lint