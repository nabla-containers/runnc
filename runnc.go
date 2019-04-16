package main

import (
	"github.com/nabla-containers/runnc/llcli"
	ll "github.com/nabla-containers/runnc/llif"
	llfs "github.com/nabla-containers/runnc/llmodules/fs"
	llnet "github.com/nabla-containers/runnc/llmodules/network"
	llnabla "github.com/nabla-containers/runnc/llruntimes/nabla"
)

func main() {
	fsH, err := llfs.NewISOFsHandler()
	if err != nil {
		panic(err)
	}
	networkH, err := llnet.NewTapBrNetworkHandler()
	if err != nil {
		panic(err)
	}
	execH, err := llnabla.NewNablaExecHandler()
	if err != nil {
		panic(err)
	}

	nablaLLCHandler := ll.RunllcHandler{
		FsH:      fsH,
		NetworkH: networkH,
		ExecH:    execH,
	}

	llcli.Runllc(nablaLLCHandler)
}
