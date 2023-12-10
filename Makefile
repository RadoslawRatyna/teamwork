run:
	go run cmd/customerimporter/main.go

build:
	go build -o build/customerimporter cmd/customerimporter/main.go

test:
	go test --race .