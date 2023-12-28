run: build
	@./bin/gstream

testp:
	@go run cmd/testconsumer/main.go

build: 
	@go build -o bin/gstream