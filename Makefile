
.PHONY: all clean server

default: server buildsucc

buildsucc:
	@echo Build FeDB Server successfully!

all: server

server:
	go build -o bin/fedb-server fedb-server/main.go

clean:
	go clean -i ./...
	rm -rf *.out
