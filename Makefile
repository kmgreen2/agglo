.DEFAULT_GOAL := all

PROTOC=protoc  -I=api/proto

PKG_SOURCES=$(filter-out %_test.go,$(wildcard pkg/**/*.go))
INTERNAL_SOURCES=$(filter-out %_test.go,$(wildcard internal/**/*.go))
export GOOS ?= $(scripts/myos.sh)

define _mockgen
mockgen -package=test -source=$(1) -destination test/mocks/$(subst /,_,$(1))

endef

define _clean_empty_mocks
find test/mocks -name '*.go' | xargs grep -L 'struct' | xargs rm

endef

all: build

.PHONY: build
build: 
	CGO_ENABLED=0 go build -o bin/regexmap cmd/regexmap/main.go
	CGO_ENABLED=0 go build -o bin/printvals cmd/printvals/main.go
	go build -o bin/binge cmd/binge/main.go
	CGO_ENABLED=0 go build -o bin/genevents cmd/genevents/main.go
	CGO_ENABLED=0 go build -o bin/activitytracker cmd/activitytracker/main.go cmd/activitytracker/activitytracker.go
	CGO_ENABLED=0 go build -o bin/ticker cmd/ticker/main.go
	CGO_ENABLED=0 go build -o bin/entwinectl cmd/entwinectl/main.go
	CGO_ENABLED=0 go build -o bin/dumbserver cmd/dumbserver/main.go
	CGO_ENABLED=0 go build -o bin/ntpsync cmd/ntpsync/main.go

.PHONY: ci-build
ci-build: 
	GOOS=linux CGO_ENABLED=0 go build -o bin/regexmap cmd/regexmap/main.go
	GOOS=linux CGO_ENABLED=0 go build -o bin/printvals cmd/printvals/main.go
	GOOS=linux go build -o bin/binge cmd/binge/main.go
	GOOS=linux CGO_ENABLED=0 go build -o bin/genevents cmd/genevents/main.go
	GOOS=linux CGO_ENABLED=0 go build -o bin/activitytracker cmd/activitytracker/main.go cmd/activitytracker/activitytracker.go
	GOOS=linux CGO_ENABLED=0 go build -o bin/ticker cmd/ticker/main.go
	GOOS=linux CGO_ENABLED=0 go build -o bin/entwinectl cmd/entwinectl/main.go
	GOOD=linux CGO_ENABLED=0 go build -o bin/dumbserver cmd/dumbserver/main.go
	GOOS=linux CGO_ENABLED=0 go build -o bin/ntpsync cmd/ntpsync/main.go

.PHONY: genmocks
genmocks:
	$(foreach  mock_source,$(PKG_SOURCES), $(call _mockgen,$(mock_source)))
	$(foreach  mock_source,$(INTERNAL_SOURCES), $(call _mockgen,$(mock_source)))
	$(call _clean_empty_mocks)

.PHONY: clean
clean: clean-test clean-build

.PHONY: clean-build
clean-build: ## remove build artifacts
	find . -name '${SERVICE}-*.tgz' -exec rm -f {} +

.PHONY: clean-test
clean-test: ## remove test and coverage artifacts
	rm -f coverage.out coverage.html

.PHONY: lint
lint:
	go list ./... | xargs golint -set_exit_status

.PHONY: test
test: build genmocks ## run tests quickly
	deployments/local/minio/run-minio.sh  && deployments/local/dynamodb/run-dynamodb.sh && \
	deployments/local/kafka/run-kafka.sh && \
	go test ./... ; \
	deployments/local/minio/stop-minio.sh ; deployments/local/dynamodb/stop-dynamodb.sh ; deployments/local/kafka/stop-kafka.sh

coverage: setup genmocks ## check code coverage
	deployments/local/minio/run-minio.sh  && deployments/local/dynamodb/run-dynamodb.sh  && \
	deployments/local/kafka/run-kafka.sh && \
	go test ./... -cover -coverprofile=coverage.txt && \
	go tool cover -html=coverage.txt -o coverage.html && \
	deployments/local/minio/stop-minio.sh ; deployments/local/dynamodb/stop-dynamodb.sh ; deployments/local/kafka/stop-kafka.sh

.PHONY: ci-test
ci-test: genmocks ## run tests quickly
	GOOS=linux go test ./... ; \

.PHONY: proto
proto: api/proto/pipeline.proto api/proto/genevents.proto
	$(PROTOC) --go_out=generated/proto api/proto/pipeline.proto 
	$(PROTOC) --go_out=generated/proto api/proto/genevents.proto
	$(PROTOC) -I$(GOPATH)/src \
		-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis api/proto/ticker.proto \
		--go_out=generated/proto --go-grpc_out=generated/proto
	$(PROTOC) -I$(GOPATH)/src \
		-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis api/proto/ticker.proto \
		--go_out=generated/proto \
		--grpc-gateway_out=logtostderr=true:generated/proto

.PHONY: lambda-local
lambda-local:
	docker build --build-arg CONFIGFILE=$(CONFIGFILE) --rm --target lambda-local . -t binge-lambda-$(LAMBDANAME)-local:latest

.PHONY: lambda-production
lambda-production:
	docker build --build-arg CONFIGFILE=$(CONFIGFILE) --rm --target lambda-production . -t binge-lambda-$(LAMBDANAME)-production:latest
