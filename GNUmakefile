default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	@if [ -z "$$BSKY_PDS_HOST" ]; then \
		echo "Skipping tests: BSKY_PDS_HOST is not set"; \
	else \
		go test -v -cover -timeout=120s -parallel=10 ./...; \
	fi


testacc:
	@if [ -z "$$BSKY_PDS_HOST" ]; then \
		echo "Skipping acceptance tests: BSKY_PDS_HOST is not set"; \
	else \
		TF_ACC=1 go test -v -cover -timeout 120m ./...; \
	fi

.PHONY: fmt lint test testacc build install generate
