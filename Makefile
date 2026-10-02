BINARY := bin/capturecrawl

.PHONY: build test fmt vet clean

build:
	go build -o $(BINARY) ./cmd/capturecrawl

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -rf bin
