run:
	go run cmd/customerimporter/main.go

build:
	go build -o build/customerimporter cmd/customerimporter/main.go

test:
	go test --race .

clean:
	rm -r build

benchmark:
	go test -bench=CountEmailDomains -benchmem -memprofile mem.prof -cpuprofile cpu.prof