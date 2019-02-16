
GO        := go
GOBUILD   := CGO_ENABLED=0 $(GO) build -v $(BUILD_FLAG)

ARCH      := "`uname -s`"
LINUX     := "Linux"
MAC       := "Darwin"

.PHONY: all clean server

default: server buildsucc

buildsucc:
	@echo Build FeDB Server successfully!

all: server

server:
	$(GOBUILD) -o bin/fedb-server fedb-server/main.go

clean:
	go clean -i ./...
	rm -rf *.out bin/*
