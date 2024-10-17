# TypeScript Linting
lint-ts:
	npm run lint

# Go Linting
lint-go:
	golangci-lint run ./...

# All Linting
lint: lint-ts lint-go


# local development
dev:
	air -- -local_dev=true
	
# CDK commands
cdk-synth: 
	cd infra/ && cdk synth "*" --profile tabsflow-dev
cdk-diff:
	cd infra/ && cdk diff "*" --profile tabsflow-dev
cdk-bootstrap:
	cd infra/ && cdk bootstrap --profile tabsflow-dev

cdk-deploy-all:
	cd infra/ && cdk deploy --all --profile tabsflow-dev

cdk-deploy-stack-service:  
	cd infra/ && cdk deploy ServiceStack --profile tabsflow-dev

cdk-deploy-stack-stateful:  
	cd infra/ && cdk deploy StatefulStack --profile tabsflow-dev

cdk-destroy-stack-service:  
	cd infra/ && cdk destroy ServiceStack --profile tabsflow-dev

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