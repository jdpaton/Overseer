.PHONY: test build server all

build:
	go build .

test:
	go test

server:
	./overseer -server

clean:
	rm -f ./overseer
	go fmt *.go
	go clean

all: clean build test server
