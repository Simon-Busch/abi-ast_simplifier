build:
	@go build -o bin/abiSimplifier

run: build
	@./bin/abiSimplifier

test:
	@go test -v ./...
