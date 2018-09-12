# Temporary makefile for debugging
default: build

build: runnc

runnc:
	GOOS=linux GOARCH=amd64 go build -o $@ .

# TODO: Currently install involves copying runnc-cont and nabla-run into /opt/runnc/bin/
install: runnc
	sudo cp runnc /usr/local/bin/runnc2

clean:
	rm -rf runnc
