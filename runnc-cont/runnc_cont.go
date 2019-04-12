// Copyright (c) 2018, IBM
// Author(s): Brandon Lum, Ricardo Koller, Dan Williams
//
// Permission to use, copy, modify, and/or distribute this software for
// any purpose with or without fee is hereby granted, provided that the
// above copyright notice and this permission notice appear in all
// copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL
// WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE
// AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL
// DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA
// OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
// TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package runnc_cont

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/nabla-containers/runnc/nabla-lib/network"
	"github.com/nabla-containers/runnc/nabla-lib/storage"
	spec "github.com/opencontainers/runtime-spec/specs-go"
)

type RunncCont struct {
	// NablaRunBin is the path to 'nabla-run' binary.
	NablaRunBin string

	NablaRunArgs []string

	// UniKernelBin is the path to 'unikernel' binary.
	UniKernelBin string

	// Tap tap device. (e.g. tap100)
	Tap string

	IPAddress net.IP
	IPMask    net.IPMask
	Gateway   net.IP
	Mac       string

	// Memory max memory size in MBs.
	Memory int64

	// Disk is the path to disk
	Disk string

	// WorkingDir current working directory.
	WorkingDir string

	// Env is a list of environment variables.
	Env []string

	// Mounts specify source and destination paths that will be copied
	// inside the container's rootfs.
	Mounts []spec.Mount
}

// NewRunncCont returns a brand new runnc-cont
func NewRunncCont(cfg Config) (*RunncCont, error) {
	if len(cfg.Disk) < 1 {
		return nil, fmt.Errorf("No disk provided")
	}

	var (
		ipAddress net.IP
		ipMask    net.IPMask
		gateway   net.IP
		tapName   string
		mac       string
		err       error
	)

	// If network details are specified, use them, if not do usual network plumbing
	if len(cfg.IPAddress) > 0 && len(cfg.Gateway) > 0 && len(tapName) > 0 {
		netstr := fmt.Sprintf("%s/%d", cfg.IPAddress, cfg.IPMask)
		ip, ipNet, err := net.ParseCIDR(netstr)
		if err != nil {
			return nil, fmt.Errorf("not a valid IP address: %s, err: %v", netstr, err)
		}

		ipAddress = ip
		ipMask = ipNet.Mask

		gateway = net.ParseIP(cfg.Gateway)
		if gateway == nil {
			return nil, fmt.Errorf("not a valid gateway address: %s", cfg.Gateway)
		}

		tapName = cfg.Tap

		if len(cfg.Mac) > 0 {
			if _, err := net.ParseMAC(cfg.Mac); err != nil {
				return nil, fmt.Errorf("not a valid mac addr: %s, err :%v", cfg.Mac, err)
			}
			mac = cfg.Mac
		}
	} else {
		// If network not provided
		// (This section will be removed and put in network module in runllc)
		if cfg.IsInDocker {
			// The tap device will get the IP assigned to the Docker
			// container veth pair.
			ipAddress, gateway, ipMask, mac, tapName, err = network.CreateMacvtapInterfaceDocker("eth0")
			if err != nil {
				return nil, fmt.Errorf("could not create %s: %v", tapName, err)
			}
		} else if cfg.IsInKubernetes {
			// The tap device will get the IP assigned to the k8s nabla
			// container veth pair.
			// XXX: This is a workaround due to an error with MacvTap, error was :
			// Could not create /dev/tap8863: open /sys/devices/virtual/net/macvtap8863/tap8863/dev: no such file or directory
			ipAddress, gateway, ipMask, mac, err = network.CreateTapInterfaceDocker(cfg.Tap, "eth0")
			if err != nil {
				return nil, fmt.Errorf("could not create %s: %v\n", cfg.Tap, err)
			}
		} else {
			// Defaults
			if len(cfg.IPAddress) == 0 {
				cfg.IPAddress = "10.0.0.2"
			}
			if cfg.IPMask == 0 {
				cfg.IPMask = 24
			}

			netstr := fmt.Sprintf("%s/%d", cfg.IPAddress, cfg.IPMask)
			ip, ipNet, err := net.ParseCIDR(netstr)
			if err != nil {
				return nil, fmt.Errorf("not a valid IP address: %s, err: %v", netstr, err)
			}

			ipAddress = ip
			ipMask = ipNet.Mask

			gateway = net.ParseIP(cfg.Gateway)
			if gateway == nil {
				return nil, fmt.Errorf("not a valid gateway address: %s", cfg.Gateway)
			}

			err = network.CreateTapInterface(cfg.Tap, &ipAddress, &ipMask)
			if err != nil {
				// Ignore networking related errors (i.e., like if the TAP
				// already exists).
				fmt.Fprintf(os.Stderr, "Could not create %s: %v\n", cfg.Tap, err)
			}
		}
	}

	return &RunncCont{
		NablaRunBin:  cfg.NablaRunBin,
		NablaRunArgs: cfg.NablaRunArgs,
		UniKernelBin: cfg.UniKernelBin,
		Tap:          cfg.Tap,
		IPAddress:    ipAddress,
		IPMask:       ipMask,
		Gateway:      gateway,
		Mac:          mac,
		Memory:       cfg.Memory,
		Disk:         cfg.Disk[0],
		WorkingDir:   cfg.WorkingDir,
		Env:          cfg.Env,
		Mounts:       cfg.Mounts,
	}, nil
}

func setupDisk(path string) (string, error) {
	if path == "" {
		return storage.CreateDummy()
	}

	pathInfo, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf(
			"can not find the disk or directory %s", path)
	}

	if pathInfo.Mode()&os.ModeDir != 0 {
		// path is a dir, so we flat it to an iso disk
		return "", fmt.Errorf("input storage %s is not an ISO", path)
	}

	// "path" is a file, so we treat it like a disk
	return path, nil
}

func (r *RunncCont) Run() error {
	var (
		mac string
		err error
	)

	disk, err := setupDisk(r.Disk)
	if err != nil {
		return fmt.Errorf("could not setup the disk: %v", err)
	}

	_, err = os.Stat(r.UniKernelBin)
	if err != nil {
		// If the unikernel path doesn't exist, look in $PATH
		unikernel, err := exec.LookPath(r.UniKernelBin)
		if err != nil {
			return fmt.Errorf("could not find the nabla file %s: %v", r.UniKernelBin, err)
		}
		r.UniKernelBin = unikernel
	}

	unikernelArgs, err := CreateRumprunArgs(r.IPAddress, r.IPMask, r.Gateway, "/",
		r.Env, r.WorkingDir, r.UniKernelBin, r.NablaRunArgs)
	if err != nil {
		return fmt.Errorf("could not create the unikernel cmdline: %v\n", err)
	}

	var args []string
	if mac != "" {
		args = []string{r.NablaRunBin,
			"--x-exec-heap",
			"--mem=" + strconv.FormatInt(r.Memory, 10),
			"--net-mac=" + mac,
			"--net=" + r.Tap,
			"--disk=" + disk,
			r.UniKernelBin,
			unikernelArgs}
	} else {
		args = []string{r.NablaRunBin,
			"--x-exec-heap",
			"--mem=" + strconv.FormatInt(r.Memory, 10),
			"--net=" + r.Tap,
			"--disk=" + disk,
			r.UniKernelBin,
			unikernelArgs}
	}

	fmt.Printf("nabla-run arg %s\n", args)

	// Set LD_LIBRARY_PATH to our dynamic libraries
	env := os.Environ()

	newenv := make([]string, 0, len(env))
	for _, v := range env {
		if strings.HasPrefix(v, "LD_LIBRARY_PATH=") {
			continue
		} else {
			newenv = append(newenv, v)
		}
	}
	newenv = append(newenv, "LD_LIBRARY_PATH=/lib64")

	err = syscall.Exec(r.NablaRunBin, args, newenv)
	if err != nil {
		return fmt.Errorf("Err from execve: %v\n", err)
	}

	return nil
}
