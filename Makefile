# Copyright (c) 2018, IBM
# Author(s): Brandon Lum, Ricardo Koller, Dan Williams
#
# Permission to use, copy, modify, and/or distribute this software for
# any purpose with or without fee is hereby granted, provided that the
# above copyright notice and this permission notice appear in all
# copies.
#
# THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL
# WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED
# WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE
# AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL
# DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA
# OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
# TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
# PERFORMANCE OF THIS SOFTWARE.

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
