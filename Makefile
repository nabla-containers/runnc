default: build

build: godep bin/runnc bin/runnc-cont

.PHONY: godep
godep: 
	dep ensure

bin/runnc: runnc.go
	GOOS=linux GOARCH=amd64 go build -o $@ .

bin/runnc-cont: runnc-cont/
	GOOS=linux GOARCH=amd64 go build -ldflags "-linkmode external -extldflags -static" -o $@ ./runnc-cont

clean:
	rm -rf bin/
