build:
	@go build -o bin/SolAstParser

run: build
	@./bin/SolAstParser

test:
	@go test -v ./...
