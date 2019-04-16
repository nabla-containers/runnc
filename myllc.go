package main

import (
	ll "github.com/nabla-containers/runnc/llif"
	llfs "github.com/nabla-containers/runnc/llmodules/fs"
	llnet "github.com/nabla-containers/runnc/llmodules/network"
	llnabla "github.com/nabla-containers/runnc/llruntimes/nabla"
)

var MyLLC = ll.RunllcHandler{
	// TODO(runllc) use the New func
	FsH:      &llfs.ISOFsHandler{},
	NetworkH: &llnet.TapBrNetworkHandler{},
	ExecH:    &llnabla.NablaExecHandler{},
}
