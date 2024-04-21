build:
	@go build -o bin/redis_clone
run: build
	@./bin/redis_clone
test:
	@go test -v ./...
