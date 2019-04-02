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
GO_BIN ?= go

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
RELEASE_VER=v0.3

RELEASE_SERVER=https://github.com/nabla-containers/nabla-base-build/releases/download/${RELEASE_VER}/

ifeq ($(GO111MODULE),on)
build: submodule_warning experimental-deps build/runnc build/nabla-run test_images
else
build: submodule_warning godep build/runnc build/nabla-run test_images
endif

container-build:
	sudo docker build . -f Dockerfile.build -t runnc-build
	sudo docker run --rm -v ${PWD}:/go/src/github.com/nabla-containers/runnc -w /go/src/github.com/nabla-containers/runnc runnc-build make

container-install:
	sudo docker build . -f Dockerfile.build -t runnc-build
	sudo docker run --rm -v /opt/runnc/:/opt/runnc/ -v /usr/local/bin:/usr/local/bin -v ${PWD}:/go/src/github.com/nabla-containers/runnc -w /go/src/github.com/nabla-containers/runnc runnc-build make install

container-uninstall:
	sudo docker rmi -f runnc-build
	make clean
	sudo hack/update_binaries.sh delete

.PHONY: godep
godep:
	dep ensure

upgrade:
	go get -u

experimental-deps:
	$(GO_BIN) build -v ./...
	make tidy

update:
	$(GO_BIN) get -u

tidy:
ifeq ($(GO111MODULE),on)
	$(GO_BIN) mod tidy
else
	echo skipping go mod tidy
endif

.PHONY: build/deps
ifeq ($(GO111MODULE),on)
build/deps: tidy
else
build/deps: godep
endif

build/runnc: build/deps create.go exec.go kill.go start.go util.go util_runner.go util_tty.go delete.go init.go runnc.go state.go util_nabla.go util_signal.go
	GOOS=linux GOARCH=${GOARCH} $(GO_BIN) build -o $@ .

solo5/tenders/spt/solo5-spt: FORCE
	make -C solo5

solo5/tests/test_hello/test_hello.spt: FORCE
	make -C solo5

.PHONY: FORCE

build/nabla-run: solo5/tenders/spt/solo5-spt
	install -m 775 -D $< $@

tests/integration/node.nabla:
	wget -nc ${RELEASE_SERVER}/node-${ARCH}.nabla -O $@ && chmod +x $@

tests/integration/test_hello.nabla: solo5/tests/test_hello/test_hello.spt
	install -m 664 -D $< $@

tests/integration/test_curl.nabla:
	wget -nc ${RELEASE_SERVER}/test_curl-${ARCH}.nabla -O $@ && chmod +x $@

install: build/runnc build/nabla-run
	sudo hack/update_binaries.sh

.PHONY: test,container-integration-test,local-integration-test,integration,integration-make
test: integration

test_images: \
tests/integration/node.nabla \
tests/integration/test_hello.nabla \
tests/integration/test_curl.nabla

integration: local-integration-test

integration-make:
	make -C tests/integration

local-integration-test: integration-make
	sudo tests/bats-core/bats -p tests/integration

#container-integration-test: test/integration/node_tests.iso
#	sudo docker run -it --rm \
#		-v $(CURDIR)/build:/build \
#		-v $(CURDIR)/tests:/tests \
#		--cap-add=NET_ADMIN \
#		-e INCONTAINER=1 \
#		ubuntu:16.04 /tests/bats-core/bats -p /tests/integration

clean:
	sudo rm -rf build/
	sudo rm -f tests/integration/node.nabla \
		tests/integration/test_hello.nabla \
		tests/integration/test_curl.nabla \
		tests/integration/node_tests.iso
	sudo make -C solo5 clean

SHELLCHECK=docker run --rm -v "$(CURDIR)":/v -w /v koalaman/shellcheck

.PHONY: shellcheck
shellcheck:
	$(SHELLCHECK) tests/integration/*.bats tests/integration/*.bash
