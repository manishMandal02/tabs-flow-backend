# TypeScript Linting
lint-ts:
	npm run lint

# Go Linting
lint-go:
	golangci-lint run ./...

# All Linting
lint: lint-ts lint-go

# CDK dev environment setup
cdk-synth: 
	cd infra/ &&  cdk synth --profile tabsflow-dev
cdk-bootstrap:
	cd infra/ &&  cdk bootstrap --profile tabsflow-dev

cdk-deploy-all:
	cd infra/ &&  cdk deploy --all --profile tabsflow-dev

# CDK stack deployment
cdk-deploy-stack:  
	cd infra/ && cdk deploy ${stack} --profile tabsflow-dev

go-test:
	echo test ${name}

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