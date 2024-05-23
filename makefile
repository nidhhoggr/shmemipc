GO ?= go
GOFMT ?= gofmt "-s"
GOFILES := $(shell find . -name "*.go")

all: build

.PHONY: build
build: 
	$(GO) mod tidy
	$(GO) build -o bin/fixed example/race/fixed/fixed.go
	$(GO) build -o bin/duplex example/duplex/duplex.go

.PHONY: run
run: 
	$(GO) mod tidy
	$(GO) run -race example/race/fixed/fixed.go
	$(GO) run -race example/duplex/duplex.go


.PHONY: test
test: 
	$(GO) clean -testcache
	$(GO) test -race -v .

.PHONY: fmt
fmt:
	$(GOFMT) -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	@diff=$$($(GOFMT) -d $(GOFILES)); \
  if [ -n "$$diff" ]; then \
    echo "Please run 'make fmt' and commit the result:"; \
    echo "$${diff}"; \
    exit 1; \
  fi;
