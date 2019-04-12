package network

import (
	"syscall"

	ll "github.com/nabla-containers/runnc/llif"
	"github.com/nabla-containers/runnc/nabla-lib/network"
	"github.com/pkg/errors"
)

type TapBrNetworkHandler struct{}

func NewTapBrNetworkHandler() (ll.NetworkHandler, error) {
	return &TapBrNetworkHandler{}, nil
}

func (h *TapBrNetworkHandler) NetworkCreateFunc(i *ll.NetworkCreateInput) (*ll.LLState, error) {
	tapName := nablaTapName(i.ContainerId)
	if err := network.CreateTapInterface(tapName, nil, nil); err != nil {
		return nil, errors.Wrap(err, "Unable to create tap in NetworkCreate")
	}

	ret := &ll.LLState{
		Options: map[string]string{
			"TapName": tapName,
		},
	}
	return ret, nil
}

func (h *TapBrNetworkHandler) NetworkRunFunc(i *ll.NetworkRunInput) (*ll.LLState, error) {
	return nil, nil
}

func (h *TapBrNetworkHandler) NetworkDestroyFunc(i *ll.NetworkDestroyInput) (*ll.LLState, error) {
	// TODO(runllc): Use options passed instead to test message passing
	tapName := nablaTapName(i.ContainerId)
	if err := network.RemoveTapDevice(tapName); err != nil {
		return nil, err
	}
	return i.NetworkState, nil
}

//err = network.CreateTapInterface(nablaTapName(id), nil, nil)

// nablaTapName returns the tapname of a given container ID
func nablaTapName(id string) string {
	if len(id) < 8 {
		panic("Insufficient uniqueness in ID")
	}
	return ("tap" + id)[:syscall.IFNAMSIZ-1]
}
