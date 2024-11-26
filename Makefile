# Include .env file if it exists
ifneq (,$(wildcard .env))
    include .env
endif

# Default 
default: dev


# install dependencies
install-deps-go: 
	go mod download && go mod tidy

install-deps-ts: 
	cd infra/ && pnpm install 
	
install-deps: install-deps-go install-deps-ts

# local development
dev:
	air -- -local_dev=true

#  Linting
lint-ts:
	cd infra/ && pnpm run lint

lint-go:
	go vet ./... && golangci-lint run ./... && testifylint --enable-all ./...

lint-all: lint-go lint-ts 

# cdk 
cdk-synth: 
	cd infra/ && cdk synth "*" --profile ${AWS_ACCOUNT_PROFILE}
	
cdk-diff:
	cd infra/ && cdk diff "*" --profile ${AWS_ACCOUNT_PROFILE}
cdk-bootstrap:
	cd infra/ && cdk bootstrap --profile ${AWS_ACCOUNT_PROFILE}

cdk-deploy-main: 
	cd infra/ && cdk deploy StatefulStack ServiceStack --profile ${AWS_ACCOUNT_PROFILE}

cdk-deploy-stack-service:  
	cd infra/ && cdk deploy ServiceStack --profile ${AWS_ACCOUNT_PROFILE}  --no-previous-parameters

cdk-deploy-main-no-approval:
	cd infra/ && cdk deploy StatefulStack ServiceStack --profile ${AWS_ACCOUNT_PROFILE} --require-approval never

cdk-deploy-stack-stateful:  
	cd infra/ && cdk deploy StatefulStack --profile ${AWS_ACCOUNT_PROFILE}

cdk-deploy-stack-oidc:
	cd infra/ && cdk deploy GithubOIDCStack --profile ${AWS_ACCOUNT_PROFILE}

cdk-deploy-stack-acm:
	cd infra/ && cdk deploy ACMStack --profile ${AWS_ACCOUNT_PROFILE}

#! cdk destroy
cdk-destroy-stack-service:  
	cd infra/ && cdk destroy ServiceStack --profile ${AWS_ACCOUNT_PROFILE}

cdk-destroy-main:
	cd infra/ && cdk destroy StatefulStack ServiceStack --profile ${AWS_ACCOUNT_PROFILE}

cdk-destroy-main-no-approval:
	cd infra/ && cdk destroy StatefulStack ServiceStack --profile ${AWS_ACCOUNT_PROFILE} --require-approval never


# tests
# cdk test
test-infra:
	cd infra/ && pnpm test

test-unit:
	go test -v ./... -short

test-integration:
	go test -v ./test/integration/... --failfast 

test-e2e: 
	go test -v ./test/e2e --failfast 

test-all: test-infra test-unit test-integration test-e2e
	echo "All tests passed"

test-cleanup: cdk-destroy-main-no-approval s3-bucket-empty
	echo "Cleanup Complete"

# show html report for go test coverage in browser
test-coverage:
	go test -coverpkg=./... ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html && open -a "Google Chrome" coverage.html

# script to clean aws s3 bucket tests
s3-bucket-empty:
	chmod +x scripts/empty-cdk-bucket.sh && scripts/empty-cdk-bucket.sh -p ${AWS_ACCOUNT_PROFILE} -r ${AWS_REGION}
