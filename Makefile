build:
	@go build -o jerome.out .

run:
	@go run main.go

test:
	@go test -v ./...
