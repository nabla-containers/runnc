package main

import (
	ll "github.com/nabla-containers/runnc/llif"
	llfs "github.com/nabla-containers/runnc/llmodules/fs"
)

var MyLLC = ll.RunllcHandler{
	// TODO(runllc) use the New func
	FSH: &llfs.ISOFSHandler{},
}
