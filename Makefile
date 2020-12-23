export

# Setup Go Variables
GOPATH := $(shell go env GOPATH)
GOBIN := $(PWD)/bin

# Invoke shell with new path to enable access to bin
PATH := $(GOBIN):$(PATH)
SHELL := env PATH=$(PATH) bash

# Other Variables
REVIEWDOG_ARG ?= -diff="git diff master"
REVIEWDOG_CONF := ./build/ci/lint/reviewdog.yml
GolangCI_CONF := ./build/ci/lint/golangci.yml

# Test Variables
ALL_TESTS := $(shell go list ./...)
UNIT_TESTS := $(shell go list ./... | grep -v "/tests")

# ######################################
# Make Commands
# ######################################

.PHONY: update-dep
update-dep:
	@./scripts/update_dep.sh

.PHONY: lint-fix
lint-fix:
	@gofmt -w -s $(shell find . -name "*.go" -type f | grep -v "_gen.go$$")
	@gofumports -w -local github.com/BinaryHexer/nbw $(shell find . -name "*.go" -type f | grep -v "_gen.go$$")
	@gofumports -w -local github.com/BinaryHexer $(shell find . -name "*.go" -type f | grep -v "_gen.go$$")
	@make lint args='--fix -v' cons_args='-v'

.PHONY:
test:
	@make unit-test $(args)

.PHONY: test-v
test-v:
	@make unit-test args='-v'

.PHONY: test-long
test-long:
	@make unit-test args='-count=5'

.PHONY: unit-test
unit-test:
	go test $(UNIT_TESTS) -race $(args)

.PHONY: coverage
coverage:
## https://www.ory.sh/golang-go-code-coverage-accurate
	go-acc $(ALL_TESTS)

.PHONY: reviewdog
reviewdog:
	reviewdog -conf=$(REVIEWDOG_CONF) $(REVIEWDOG_ARG)

# ######################################
# END Make Commands
# ######################################
