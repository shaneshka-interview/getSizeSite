PROJECTNAME=$(shell basename "$(PWD)")

## run:
run:
	go run ./main.go -f ./example/simple.txt

## test:
test:
	go test ./...

## race:
race:
	go test -v -race ./...

## bench:
bench:
	go test -bench=. -benchmem -cpuprofile=cpu.out -memprofile=mem.out > bench.txt

## cover:
cover:
	go test -coverprofile cover.out
	go tool cover -html=cover.out


.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
