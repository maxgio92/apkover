go := $(shell command -v go)

.PHONY: build
build:
	@$(go) build -v .

.PHONY: test
test: test/e2e

.PHONY: test/e2e
test/e2e:
	@$(go) test -cover -v main_test.go

