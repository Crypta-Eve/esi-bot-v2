build:
	go build -o bot ./cmd/eb2/

run:
	go run ./cmd/eb2/*.go	

run-built: build
	./bot