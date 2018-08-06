# Temporary makefile for debugging
default: build

build: runnc

runnc:
	GOOS=linux GOARCH=amd64 go build -o $@ .

clean:
	rm -rf runnc
