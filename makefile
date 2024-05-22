GO ?= go
GOFMT ?= gofmt "-s"
GOFILES := $(shell find . -name "*.go")

all: build

.PHONY: build
build: 
	$(GO) mod tidy
	$(GO) build -o bin/example example/main.go

.PHONY: run
run: 
	$(GO) mod tidy
	$(GO) run -race example/main.go


.PHONY: test
test: 
	$(GO) test -race -v

.PHONY: examples
examples: build 
	./bin/example

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

.PHONY: protogen
protogen:
	protoc protos/*.proto  --go_out=.
