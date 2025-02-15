run:
	go run main.go

test:
	go test -timeout 5s ./...

profile:
	go test -timeout 5s -cpuprofile cpu.prof -blockprofile block.prof -memprofile mem.prof ./...
