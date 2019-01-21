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

ARCH=$(shell uname --m)
ifeq ($(ARCH), aarch64)
GOARCH="arm64"
else ifeq ($(ARCH), x86_64)
GOARCH="amd64"
endif

default: build

SUBMOD_NEEDS_UPDATE=$(shell [ -z "`git submodule | grep -v "^ "`" ] && echo 0 || echo 1)

ifeq ($(SUBMOD_NEEDS_UPDATE), 1)
submodule_warning:
	$(info #####################################################)
	$(info # Warning: git submodule out of date!!!!            #)
	$(info #          Please run `git submodule update --init` #)
	$(info #####################################################)
	$(info )
	$(info Continuing in 5 seconds...)
	$(shell	sleep 5)
else
submodule_warning:

endif

# Synced release version to download from
RELEASE_VER=v0.2

RELEASE_SERVER=https://github.com/nabla-containers/nabla-base-build/releases/download/${RELEASE_VER}/

build: submodule_warning godep build/runnc build/runnc-cont build/nabla-run

container-build:
	sudo docker build . -f Dockerfile.build -t runnc-build
	sudo docker run --rm -v ${PWD}:/go/src/github.com/cloudkernels/runnc -w /go/src/github.com/cloudkernels/runnc runnc-build make

container-install:
	sudo docker build . -f Dockerfile.build -t runnc-build
	sudo docker run --rm -v /opt/runnc/:/opt/runnc/ -v /usr/local/bin:/usr/local/bin -v ${PWD}:/go/src/github.com/cloudkernels/runnc -w /go/src/github.com/cloudkernels/runnc runnc-build make install

.PHONY: godep
godep:
	dep ensure

build/runnc: godep create.go exec.go kill.go start.go util.go util_runner.go util_tty.go delete.go  init.go runnc.go state.go util_nabla.go util_signal.go
	GOOS=linux GOARCH=$(GOARCH) go build -o $@ .

build/runnc-cont: godep runnc-cont/*
	GOOS=linux GOARCH=$(GOARCH) go build -ldflags "-linkmode external -extldflags -static" -o $@ ./runnc-cont

solo5/tenders/spt/solo5-spt: FORCE
	make -C solo5

.PHONY: FORCE

build/nabla-run: solo5/tenders/spt/solo5-spt
	install -m 775 -D $< $@

install: build/runnc build/runnc-cont build/nabla-run
	sudo hack/copy_binaries.sh

clean:
	sudo rm -rf build/
	sudo rm -f tests/integration/node.nabla \
		tests/integration/test_hello.nabla \
		tests/integration/test_curl.nabla \
		tests/integration/node_tests.iso
	sudo make -C solo5 clean
