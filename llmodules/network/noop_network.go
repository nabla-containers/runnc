package network

import (
	ll "github.com/nabla-containers/runnc/llif"
)

type noopNetworkHandler struct{}

func NewnoopNetworkHandler() (ll.NetworkHandler, error) {
	return &noopNetworkHandler{}, nil
}

func (h *noopNetworkHandler) NetworkCreateFunc(i *ll.NetworkCreateInput) (*ll.LLState, error) {
	ret := &ll.LLState{}
	return ret, nil
}

func (h *noopNetworkHandler) NetworkRunFunc(i *ll.NetworkRunInput) (*ll.LLState, error) {
	return i.NetworkState, nil
}

func (h *noopNetworkHandler) NetworkDestroyFunc(i *ll.NetworkDestroyInput) (*ll.LLState, error) {
	return i.NetworkState, nil
}
