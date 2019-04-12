package fs

import (
	"os"
	"path/filepath"

	"github.com/nabla-containers/runnc/libcontainer/configs"
	ll "github.com/nabla-containers/runnc/llif"
	"github.com/nabla-containers/runnc/nabla-lib/storage"
	"github.com/nabla-containers/runnc/utils"
	"github.com/pkg/errors"
)

type ISOFSHandler struct{}

func NewISOFSHandler() (ll.FSHandler, error) {
	return &ISOFSHandler{}, nil
}

func (h *ISOFSHandler) FSCreateFunc(i *ll.FSCreateInput) (*ll.LLState, error) {
	fsPath, err := createRootfsISO(i.Config, i.ContainerRoot)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create rootfs ISO")
	}

	ret := &ll.LLState{}
	ret.Options = map[string]string{
		"FsPath": fsPath,
	}

	return ret, nil
}

func (h *ISOFSHandler) FSRunFunc(i *ll.FSRunInput) (*ll.LLState, error) {
	return i.FSState, nil
}

func (h *ISOFSHandler) FSDestroyFunc(i *ll.FSDestroyInput) (*ll.LLState, error) {
	if err := os.RemoveAll(i.ContainerRoot); err != nil {
		return nil, err
	}
	return i.FSState, nil
}

func createRootfsISO(config *configs.Config, containerRoot string) (string, error) {
	rootfsPath := config.Rootfs
	targetISOPath := filepath.Join(containerRoot, "rootfs.iso")
	for _, mount := range config.Mounts {
		if (mount.Destination == "/etc/resolv.conf") ||
			(mount.Destination == "/etc/hosts") ||
			(mount.Destination == "/etc/hostname") {
			dest := filepath.Join(rootfsPath, mount.Destination)
			source := mount.Source
			if err := utils.Copy(dest, source); err != nil {
				return "", errors.Wrap(err, "Unable to copy "+source+" to "+dest)
			}
		}
	}
	_, err := storage.CreateIso(rootfsPath, &targetISOPath)
	if err != nil {
		return "", errors.Wrap(err, "Error creating iso from rootfs")
	}
	return targetISOPath, nil
}
