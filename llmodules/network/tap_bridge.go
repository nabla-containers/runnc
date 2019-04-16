package network

import (
	"fmt"
	"syscall"

	ll "github.com/nabla-containers/runnc/llif"
	"github.com/nabla-containers/runnc/nabla-lib/network"
	"github.com/pkg/errors"
)

type tapBrNetworkHandler struct{}

func NewTapBrNetworkHandler() (ll.NetworkHandler, error) {
	return &tapBrNetworkHandler{}, nil
}

func (h *tapBrNetworkHandler) NetworkCreateFunc(i *ll.NetworkCreateInput) (*ll.LLState, error) {
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

func (h *tapBrNetworkHandler) NetworkRunFunc(i *ll.NetworkRunInput) (*ll.LLState, error) {
	tapName, ok := i.NetworkState.Options["TapName"]
	if !ok {
		return nil, errors.New("Unable to get tap name")
	}

	// The tap device will get the IP assigned to the k8s nabla
	// container veth pair.
	// XXX: This is a workaround due to an error with MacvTap, error was :
	// Could not create /dev/tap8863: open /sys/devices/virtual/net/macvtap8863/tap8863/dev: no such file or directory
	ipAddress, gateway, ipMask, mac, err := network.CreateTapInterfaceDocker(nablaTapName(i.ContainerId), "eth0")
	if err != nil {
		return nil, errors.Wrap(err, "Unable to configure network runtime")
	}
	cidr, totalBits := ipMask.Size()
	if totalBits != 32 {
		return nil, errors.New("Unexpected IP address number of bits")
	}

	ret := &ll.LLState{
		Options: map[string]string{
			"IPAddress": ipAddress.String(),
			"Gateway":   gateway.String(),
			"IPMask":    fmt.Sprintf("%d", cidr),
			"Mac":       mac,
			"TapName":   tapName,
		},
	}

	return ret, nil
}

func (h *tapBrNetworkHandler) NetworkDestroyFunc(i *ll.NetworkDestroyInput) (*ll.LLState, error) {
	tapName, ok := i.NetworkState.Options["TapName"]
	if !ok {
		return nil, errors.New("Unable to get tap name")
	}
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
