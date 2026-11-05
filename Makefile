.PHONY: build run up down clean test

APP_NAME=keyforge

build:
	go build -ldflags="-w -s" -o bin/$(APP_NAME) ./cmd/$(APP_NAME)

run: build
	./bin/$(APP_NAME)

up:
	docker-compose up --build -d

down:
	docker-compose down -v

clean:
	rm -rf bin/
	go clean -testcache

test:
	go test -v -race ./...