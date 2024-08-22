.PHONY: all server cli clean

all: server cli

server:
	go build -v -o server cmd/server/***

cli:
	go build -v -o cli cmd/cli/***

clean:
	rm server cli
