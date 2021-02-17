.DEFAULT_GOAL := all

PROTOC=protoc  -I=api/proto

PKG_SOURCES = $(filter-out %_test.go,$(wildcard pkg/**/*.go))
INTERNAL_SOURCES = $(filter-out %_test.go,$(wildcard internal/**/*.go))
export GOOS ?= darwin

define _mockgen
mockgen -package=test -source=$(1) -destination test/mocks/$(subst /,_,$(1))

endef

define _clean_empty_mocks
find test/mocks -name '*.go' | xargs grep -L 'struct' | xargs rm

endef

all: build

.PHONY: build
build: 
	CGO_ENABLED=1 go build -o bin/regexmap cmd/regexmap/main.go
	CGO_ENABLED=1 go build -o bin/printvals cmd/printvals/main.go
	CGO_ENABLED=1 go build -o bin/binge cmd/binge/main.go
	CGO_ENABLED=1 go build -o bin/genevents cmd/genevents/main.go
	CGO_ENABLED=1 go build -o bin/activitytracker cmd/activitytracker/main.go cmd/activitytracker/activitytracker.go

.PHONY: ci-build
ci-build: 
	GOOS=linux CGO_ENABLED=1 go build -o bin/regexmap cmd/regexmap/main.go
	GOOS=linux CGO_ENABLED=1 go build -o bin/printvals cmd/printvals/main.go
	GOOS=linux CGO_ENABLED=1 go build -o bin/binge cmd/binge/main.go
	GOOS=linux CGO_ENABLED=1 go build -o bin/genevents cmd/genevents/main.go
	GOOS=linux CGO_ENABLED=1 go build -o bin/activitytracker cmd/activitytracker/main.go cmd/activitytracker/activitytracker.go

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
	go test ./... ; \

.PHONY: ci-test
ci-test: genmocks ## run tests quickly
	GOOS=linux go test ./... ; \

.PHONY: coverage
coverage: genmocks ## check code coverage
	go test ./... -cover -coverprofile=coverage.txt
	go tool cover -html=coverage.txt -o coverage.html


.PHONY: proto
proto: api/proto/pipeline.proto api/proto/genevents.proto
	$(PROTOC) --go_out=generated/proto api/proto/pipeline.proto 
	$(PROTOC) --go_out=generated/proto api/proto/genevents.proto

