[![Build Status](https://travis-ci.org/nabla-containers/runnc.svg?branch=master)](https://travis-ci.org/nabla-containers/runnc)
[![Go Report Card](https://goreportcard.com/badge/github.com/nabla-containers/runnc)](https://goreportcard.com/report/github.com/nabla-containers/runnc)

# Runnc

`runnc` is the nabla-container runtime which interfaces with the container OCI runtime spec to create a nabla-container runtime. The runtime currently re-uses functionality from `runc` for some setup steps, but will eventually be self-sufficient in providing nabla-container equivalent setups.

There is initial aarch64 support. For more information please check the [README.aarch64](README.aarch64.md) file

## Getting started with the go repo!

1. Ensure that your `GOPATH` is set. (https://github.com/golang/go/wiki/SettingGOPATH)
2. Go get the repo `go get github.com/nabla-containers/runnc`
3. Install genisoimage on host `sudo apt install genisoimage`
4. Install jq on host `sudo apt install jq`
5. Ensure that docker is installed (docker-ce recent versions, i.e. v15 onwards)

Docker major versions tested with:

- docker-ce 17

## Build and Install Runnc

We have created two ways to build and install `runnc`. You may build inside a container, or perform a local build.


### Build with a container
```
# Go to the repo
cd $GOPATH/src/github.com/nabla-containers/runnc

# make container-build to build runnc.
make container-build

# make container-install to install runnc
make container-install
```

### Build locally
```
# Go to the repo
cd $GOPATH/src/github.com/nabla-containers/runnc

# Get the neceesary binaries for the runtime
make build

# Install libseccomp on the host
sudo apt install libseccomp-dev

# Install the appropriate binaries/libraries
make install
```

## Configure Docker to use new Runtime

0. Install genisoimage and libseccomp on host
```
sudo apt install genisoimage
sudo apt install libseccomp-dev
```

1. Modify to add runtime to `/etc/docker/daemon.json`, for example:
```
{
    "runtimes": {
        "runnc": {
                "path": "/usr/local/bin/runnc"
        }
    }
}
```

2. Restart docker

```
systemctl restart docker
```

3. Run with runtime:

```
sudo docker run --rm --runtime=runnc nablact/nabla-node-base:v0.3
```

## Limitations

There are many. Some are fixable and being worked on, some are fixable but harder and will take some time, and some others are ones that we don't really know how to fix (or possibly not worth fixing).

Container runtime limitations:
- Unable to properly handle /32 IP address assignments. Current hack converts cidr from 32 to 1

Here are some missing features that we are currently working on:
- ~~a golang base image~~
- MirageOS and IncludeOS base images
- base images for all the known apps that can run on rumprun (from rumprun-packages), like openjdk.
- a writable file system. Currently only `/tmp` is writable.
- support for committing the image
- volumes (as in `docker -v /a:/a`)
- not ignoring cgroups (start with the memory ones)
- multiple network interfaces
- ~~not using `runc` as an intermediate step. Right now, `runnc` calls `runc` which then calls `nabla-run`~~
- `runnc` use of interactive console/tty (i.e. `docker run -it`)

These are some harder features (sorted from more to less important):
- allow dynamic loading of libraries. The nabla runtime can only start static binaries and that seems to be OK for most things, but one big limitation is that python can't load modules with `.so`'s in them.
- use something other than the rumprun netbsd libc: we could use LKL, or IncludeOS recent support for musl libc, or wrap the netbsd libc on something that looks like glibc
- `mmap()` for sharing memory to/from another process (nabla and not nabla)
- GPU support
- support for custom/host namespaces
- `docker exec`. What exactly would it run? what do people do for microcontainers (like an image with just one statically built go binary)
- "real" TLS (Thread Local Storage) support. Right now, pthread-key based thread specific data is supported (`pthread_key_create` / `pthread_setspecific`), but it does not use the real segment-based TLS. So you would get the correct behavior, but not the best-performing implementation. Also, `__thread` is not supported.

Harder limitations that we don't know how to fix (nor we don't know if they should be fixed):
- support for running vanilla images. Currently nabla can only run nabla based images.
- `fork()`. Should a nabla process fork another nabla process (unikernel)? a single unikernel can't run multiple address spaces
