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

# Synced relelase version to downlaod from
RELEASE_VER=v0.1
SOLO5_RELEASE_VER=v0.2.2.1-seccomp

RELEASE_SERVER=https://github.com/nabla-containers/nabla-base-build/releases/download/${RELEASE_VER}/
SOLO5_RELEASE_SERVER=https://github.com/nabla-containers/solo5/releases/download/${SOLO5_RELEASE_VER}/

build: godep build/runnc build/runnc-cont build/nabla-run

container-build: 
	sudo docker build . -f Dockerfile.build -t runnc-build
	sudo docker run --rm -v ${PWD}:/go/src/github.com/nabla-containers/runnc -w /go/src/github.com/nabla-containers/runnc runnc-build make 

container-install: 
	sudo docker build . -f Dockerfile.build -t runnc-build
	sudo docker run --rm -v /opt/runnc/lib/:/opt/runnc/lib/ -v /usr/local/bin:/usr/local/bin -v ${PWD}:/go/src/github.com/nabla-containers/runnc -w /go/src/github.com/nabla-containers/runnc runnc-build make install



.PHONY: godep
godep: 
	dep ensure

build/runnc: runnc.go
	GOOS=linux GOARCH=amd64 go build -o $@ .

build/runnc-cont: runnc-cont/*
	GOOS=linux GOARCH=amd64 go build -ldflags "-linkmode external -extldflags -static" -o $@ ./runnc-cont

build/nabla-run: Makefile
	wget -N ${SOLO5_RELEASE_SERVER}/nabla-run -O $@ && chmod +x $@

tests/integration/node.nabla: 
	wget -nc ${RELEASE_SERVER}/node.nabla -O $@ && chmod +x $@

tests/integration/test_hello.nabla: 
	wget -nc ${SOLO5_RELEASE_SERVER}/test_hello.nabla -O $@ && chmod +x $@

tests/integration/test_curl.nabla: 
	wget -nc ${RELEASE_SERVER}/test_curl.nabla -O $@ && chmod +x $@

install: build/runnc build/runnc-cont build/nabla-run
	sudo hack/copy_binaries.sh
	sudo hack/copy_libraries.sh

.PHONY: test,container-integration-test,local-integration-test,integration
test: integration

test_images: \
tests/integration/node.nabla \
tests/integration/test_hello.nabla \
tests/integration/test_curl.nabla

integration: local-integration-test container-integration-test

test/integration/node_tests.iso:
	make -C tests/integration

local-integration-test: test_images test/integration/node_tests.iso
	sudo tests/bats-core/bats -p tests/integration

container-integration-test: test_images test/integration/node_tests.iso
	sudo docker run -it --rm \
		-v $(CURDIR)/build:/build \
		-v $(CURDIR)/tests:/tests \
		--cap-add=NET_ADMIN \
		-e INCONTAINER=1 \
		ubuntu:16.04 /tests/bats-core/bats -p /tests/integration

clean:
	sudo rm -rf build/
	rm -f tests/integration/node.nabla \
		tests/integration/test_hello.nabla \
		tests/integration/test_curl.nabla

