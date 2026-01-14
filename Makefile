.PHONY: build clean test install dev

BINARY_NAME=zeus
VERSION=1.0.0

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)

test:
	go test -v ./...

install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

dev:
	go run . $(ARGS)
