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

package main

import (
	"flag"
	"fmt"
	"github.ibm.com/nabla-containers/nabla-lib/network"
	"github.ibm.com/nabla-containers/nabla-lib/storage"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type arrayEnvVars []string

func (i *arrayEnvVars) String() string {
	return "my string representation"
}

func (i *arrayEnvVars) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var envVars arrayEnvVars

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
		return "", fmt.Errorf(
			"Input storage %s is not an ISO", path)
	}

	// "path" is a file, so we treat it like a disk
	return path, nil
}

func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
	}

	nablarun := flag.String("nabla-run", "./nabla-run",
		"Path to desired nabla-run to use.")
	unikernel := flag.String("unikernel", "./node.nablet",
		"Unikernel executable file. It will be looked for in $PATH.")
	tap := flag.String("tap", "tap100",
		"Tap device (e.g., tap100). Defaults to tap100.")
	ipv4 := flag.String("ipv4", "10.0.0.2",
		"IP v4 address (defaults to 10.0.0.2)")
	maskv4 := flag.Int("maskv4", 24,
		"Mask v4 (defaults to 24)")
	gwv4 := flag.String("gwv4", "10.0.0.1",
		"Gateway v4 (defaults to 10.0.0.1")
	inDocker := flag.Bool("docker", false,
		"Is this running in a Docker container")
	volume := flag.String("volume", ":",
		"'--volume <SRC>:<DST>'. "+
			"<SRC> is the directory or device to mount, and <DST> "+
			"is the path where it's going to be mounted in the unikernel.")
	flag.Var(&envVars, "env",
		"Environment variable; add as many '-env A -env B' as needed")
	flag.Parse()

	vol := strings.Split(*volume, ":")
	if len(vol) != 2 {
		panic("Volume should be '--volume <SRC>:<DST>'.")
	}

	netstr := fmt.Sprintf("%s/%d", *ipv4, *maskv4)
	ip, ipNet, err := net.ParseCIDR(netstr)
	if err != nil {
		panic("Not a valid IP address")
	}
	gw := net.ParseIP(*gwv4)

	cmdargs := strings.Join(flag.Args(), " ")

	os.Exit(run(*nablarun, *unikernel, *tap, ip, ipNet.Mask, gw,
		*inDocker, vol, cmdargs, envVars))
}

func run(nablarun string, unikernel string, tapName string,
	ip net.IP, mask net.IPMask, gw net.IP,
	inDocker bool, volume []string,
	cmdargs string, envVars []string) int {

	disk, err := setupDisk(volume[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not setup the disk: %v\n", err)
		return 1
	}

	if inDocker {
		// The tap device will get the IP assigned to the Docker
		// container veth pair.
		ip, gw, mask, err = network.CreateTapInterfaceDocker(tapName, "eth0")
	} else {
		err = network.CreateTapInterface(tapName, gw, mask)
	}
	if err != nil {
		// Ignore networking related errors (i.e., like if the TAP
		// already exists).
		fmt.Fprintf(os.Stderr, "Could not create the TAP: %v\n", err)
	}

	_, err = os.Stat(unikernel)
	if err != nil {
		// If the unikernel path doesn't exist, look in $PATH
		_unikernel, err := exec.LookPath(unikernel)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Could not find the nabla file %s: %v\n",
				unikernel, err)
			return 1
		}
		unikernel = _unikernel
	}

	unikernelArgs, err := CreateRumprunArgs(ip, mask, gw, volume[1],
		envVars, unikernel, cmdargs)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Could not create the unikernel cmdline: %v\n", err)
		return 1
	}

	args := []string{nablarun,
		"--net=" + tapName,
		"--disk=" + disk,
		unikernel,
		unikernelArgs}

	fmt.Printf("%s\n", args)

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
    newenv =  append(newenv, "LD_LIBRARY_PATH=/lib64")

    err = syscall.Exec(nablarun, args, newenv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err from execve: %v\n", err)
		return 1
	}

	return 0
}
