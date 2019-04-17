package fs

import (
	"os"

	ll "github.com/nabla-containers/runnc/llif"
)

type noopFsHandler struct{}

func NewnoopFsHandler() (ll.FsHandler, error) {
	return &noopFsHandler{}, nil
}

func (h *noopFsHandler) FsCreateFunc(i *ll.FsCreateInput) (*ll.LLState, error) {
	ret := &ll.LLState{}
	return ret, nil
}

func (h *noopFsHandler) FsRunFunc(i *ll.FsRunInput) (*ll.LLState, error) {
	return i.FsState, nil
}

func (h *noopFsHandler) FsDestroyFunc(i *ll.FsDestroyInput) (*ll.LLState, error) {
	if err := os.RemoveAll(i.ContainerRoot); err != nil {
		return nil, err
	}
	return i.FsState, nil
}
