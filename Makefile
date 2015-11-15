SOURCES=$(shell find . -name "*.go" | grep -v Godeps)

build:
	go build  ./...

deps-get:
	go get -t ./...

run-node-1:
	go run main.go -addr=127.0.0.1:12000 -cluster=127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002

run-node-2:
	go run main.go -addr=127.0.0.1:12001 -cluster=127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002

run-node-3:
	go run main.go -addr=127.0.0.1:12002 -cluster=127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002

test:
	go test -v ./...

test-race:
	GOMAXPROCS=2 GORACE="halt_on_error=1" go test -race  ./...
