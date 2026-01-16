.PHONY: build clean test install dev

BINARY_NAME=zeus
VERSION=1.0.0

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)

test:
	go test -v ./...

install:
	go install -ldflags "-X main.version=$(VERSION)" .

dev:
	go run . $(ARGS)
