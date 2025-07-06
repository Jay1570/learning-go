run: build
	@./bin/ecom

build:
	@go build -o bin/ecom cmd/main.go

test:
	@go test -v ./...
