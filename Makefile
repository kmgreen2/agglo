.DEFAULT_GOAL := all

PROTOC=protoc  -I=api/proto

PKG_SOURCES = $(filter-out %_test.go,$(wildcard pkg/**/*.go))
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

.PHONY: genmocks
genmocks:
	$(foreach  mock_source,$(PKG_SOURCES), $(call _mockgen,$(mock_source)))
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
lint: setup ## check style with flake8
	go list ./... | xargs golint -set_exit_status

.PHONY: test
test: build genmocks ## run tests quickly
	go test ./... ; \

.PHONY: coverage
coverage: setup genmocks ## check code coverage
	go test ./... -cover -coverprofile=coverage.txt
	go tool cover -html=coverage.txt -o coverage.html


.PHONY: proto
proto: api/proto/pipeline.proto
	$(PROTOC) --go_out=pkg/proto api/proto/pipeline.proto 
