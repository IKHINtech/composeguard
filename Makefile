APP_NAME=composeguard
CMD=./cmd/composeguard

.PHONY: fmt test build run clean

fmt:
	go fmt ./...

test:
	go test ./...

build:
	go build -o $(APP_NAME) $(CMD)

run:
	go run $(CMD) check

clean:
	rm -f $(APP_NAME)
	rm -rf dist
