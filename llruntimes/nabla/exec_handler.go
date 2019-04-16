package fs

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nabla-containers/runnc/libcontainer/configs"
	ll "github.com/nabla-containers/runnc/llif"
	"github.com/nabla-containers/runnc/llruntimes/nabla/runnc-cont"
	"github.com/pkg/errors"
)

var (
	NablaBinDir = "/opt/runnc/bin/"
	NablaRunBin = NablaBinDir + "nabla-run"
)

type nablaExecHandler struct{}

func NewNablaExecHandler() (ll.ExecHandler, error) {
	return &nablaExecHandler{}, nil
}

func (h *nablaExecHandler) ExecCreateFunc(i *ll.ExecCreateInput) (*ll.LLState, error) {
	ret := &ll.LLState{}
	return ret, nil
}

func (h *nablaExecHandler) ExecRunFunc(i *ll.ExecRunInput) error {
	networkOptions := i.NetworkState.Options
	fsOptions := i.FsState.Options
	config := i.Config
	contRoot := i.ContainerRoot

	runncCont, err := newRunncCont(contRoot, *config, networkOptions, fsOptions)
	if err != nil {
		return errors.Wrap(err, "Unable to construct nabla run args")
	}

	// Shouldn't return
	return runncCont.Run()
}

func (h *nablaExecHandler) ExecDestroyFunc(i *ll.ExecDestroyInput) (*ll.LLState, error) {
	ret := &ll.LLState{}
	return ret, nil
}

func newRunncCont(containerRoot string, cfg configs.Config, networkMap map[string]string, fsMap map[string]string) (*runnc_cont.RunncCont, error) {
	if len(cfg.Args) == 0 {
		return nil, fmt.Errorf("OCI process args are empty")
	}

	if !strings.HasSuffix(cfg.Args[0], ".nabla") {
		return nil, fmt.Errorf("entrypoint is not a .nabla file")
	}

	cidr, err := strconv.Atoi(networkMap["IPMask"])
	if err != nil {
		return nil, fmt.Errorf("Unablae to parse IPMask: %v", cidr)
	}

	c := runnc_cont.Config{
		NablaRunBin:  NablaRunBin,
		UniKernelBin: filepath.Join(containerRoot, cfg.Args[0]),
		Memory:       cfg.Memory,
		Tap:          networkMap["TapName"],
		Disk:         []string{fsMap["FsPath"]},
		WorkingDir:   cfg.Cwd,
		Env:          cfg.Env,
		NablaRunArgs: cfg.Args[1:],
		Mounts:       cfg.Mounts,
		IPAddress:    networkMap["IPAddress"],
		Mac:          networkMap["Mac"],
		Gateway:      networkMap["Gateway"],
		IPMask:       cidr,
	}

	cont, err := runnc_cont.NewRunncCont(c)
	if err != nil {
		return nil, err
	}

	return cont, nil
}
