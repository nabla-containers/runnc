default: build

# Temporarily have server to download other binaries from, 
# this will eventually be github release
RELEASE_SERVER=9.12.247.246
# Synced relelase version to downlaod from
RELEASE_VER=NONE

build: godep bin/runnc bin/runnc-cont bin/ukvm-bin

.PHONY: godep
godep: 
	dep ensure

bin/runnc: runnc.go
	GOOS=linux GOARCH=amd64 go build -o $@ .

bin/runnc-cont: runnc-cont/
	GOOS=linux GOARCH=amd64 go build -ldflags "-linkmode external -extldflags -static" -o $@ ./runnc-cont

bin/ukvm-bin: 
	wget -nc http://${RELEASE_SERVER}/nablet-build/ukvm-bin -O $@ && chmod +x $@

clean:
	rm -rf bin/
