.PHONY: build run test clean

build:
	go build -o multilog cmd/multilog/main.go

run: build
	./multilog

test:
	go test ./...

clean:
	rm -rf multilog