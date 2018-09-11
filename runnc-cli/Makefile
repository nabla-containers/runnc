# Temporary makefile for debugging
default: build

build: runnc

runnc:
	GOOS=linux GOARCH=amd64 go build -o $@ .

install: runnc
	sudo cp runnc /usr/local/bin/runnc2

clean:
	rm -rf runnc
