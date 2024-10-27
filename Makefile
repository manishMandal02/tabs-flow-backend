include .env


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
	cd infra/ && cdk synth "*" --profile ${AWS_ACCOUNT_PROFILE}
cdk-diff:
	cd infra/ && cdk diff "*" --profile ${AWS_ACCOUNT_PROFILE}
cdk-bootstrap:
	cd infra/ && cdk bootstrap --profile ${AWS_ACCOUNT_PROFILE}

cdk-deploy-all:
	cd infra/ && cdk deploy --all --profile ${AWS_ACCOUNT_PROFILE}

cdk-deploy-stack-service:  
	cd infra/ && cdk deploy ServiceStack --profile ${AWS_ACCOUNT_PROFILE} --no-previous-parameters

cdk-deploy-stack-stateful:  
	cd infra/ && cdk deploy StatefulStack --profile ${AWS_ACCOUNT_PROFILE}

cdk-destroy-stack-service:  
	cd infra/ && cdk destroy ServiceStack --profile ${AWS_ACCOUNT_PROFILE}

cdk-destroy-all:
	cd infra/ && cdk destroy --all --profile ${AWS_ACCOUNT_PROFILE}
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