[![Build Status](https://travis-ci.org/nabla-containers/runnc.svg?branch=master)](https://travis-ci.org/nabla-containers/runnc)
[![Go Report Card](https://goreportcard.com/badge/github.com/nabla-containers/runnc)](https://goreportcard.com/report/github.com/nabla-containers/runnc)

# Runnc

`runnc` is the nabla-container runtime which interfaces with the container OCI runtime spec to create a nabla-container runtime. The runtime currently re-uses functionality from `runc` for some setup steps, but will eventually be self-sufficient in providing nabla-container equivalent setups.

## Getting started with the go repo!

1. Ensure that your `GOPATH` is set. (https://github.com/golang/go/wiki/SettingGOPATH)
2. Go get the repo `go get github.com/nabla-containers/runnc`
3. Install genisoimage on host `sudo apt install genisoimage`
4. Ensure that docker is installed (docker-ce recent versions, i.e. v15 onwards)

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

0. Install genisoimage on host
```
sudo apt install genisoimage
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
sudo docker run --rm --runtime=runnc nablact/nabla-node-base
```
