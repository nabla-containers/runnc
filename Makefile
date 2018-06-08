default: build

build: runnc

runnc: runnc.go
	GOOS=linux GOARCH=amd64 go build -o $@ .

clean:
	rm -f runnc
