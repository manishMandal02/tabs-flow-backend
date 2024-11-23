include .env

# Default 
default: dev

# local development
dev:
	air -- -local_dev=true

#  Linting
lint-ts:
	cd infra/ && npm run lint

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


# tests
# cdl/infra test
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

test-cleanup: cdk-destroy-all s3-bucket-empty
	echo "Cleanup Complete"

# show html report for go test coverage in browser
test-coverage:
	go test -coverpkg=./... ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html && open -a "Google Chrome" coverage.html

# script to clean aws s3 bucket tests
s3-bucket-empty:
	chmod +x scripts/empty-cdk-bucket.sh && scripts/empty-cdk-bucket.sh -p ${AWS_ACCOUNT_PROFILE} -r ${AWS_REGION}
