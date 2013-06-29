.PHONY: test build server all

build:
	go build .

test:
	go test -i

server:
	./overseer -server

clean:
	rm -f ./overseer
	go fmt *.go
	go clean

run: server

all: clean build test server
