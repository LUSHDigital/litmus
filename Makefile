.PHONY: test

test:
	go test ./... -v

build:
	go build -o litmus_local

run: build
	./litmus_local --config=examples