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
	"encoding/json"
	"fmt"
	"github.com/nabla-containers/runnc/nabla-lib/storage"
	"github.com/nabla-containers/runnc/utils"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type config struct {
	RuncBinaryPath string
	Bundle         string
	PidFile        string
	Root           string
	Log            string
	LogFormat      string
	ConsoleSocket  string
}

var (
	// LogFile handlier to use for logging
	LogFile *os.File
	// LogFilePath to log to
	LogFilePath = "/tmp/runnc.log"
	// Config is the config for the runc command invoked
	Config = config{}
	// DefaultRuncBinaryPath is the default path to the docker-runc binary
	DefaultRuncBinaryPath = "/usr/bin/docker-runc"
)

func runRunc(args, env []string) {
	newargs := append([]string{Config.RuncBinaryPath}, args[1:]...)

	newenv := make([]string, len(env))
	for i, v := range env {
		if strings.HasPrefix(v, "_=") {
			newenv[i] = "_=" + Config.RuncBinaryPath
		} else {
			newenv[i] = v
		}
	}
	log.Printf("Calling runc with following args: %v, %v, %v", Config.RuncBinaryPath, newargs, newenv)

	if execErr := syscall.Exec(Config.RuncBinaryPath, newargs, newenv); execErr != nil {
		panic(execErr)
	}
}

// linuxPermCapNetAdmin is the string to signify the CAP_NET_ADMIN linux capability
var linuxPermCapNetAdmin = "CAP_NET_ADMIN"

// readSpec reads the runtime spec (config.json) from the bundle
func readSpec(bundlePath string) (*spec.Spec, error) {
	specFile := filepath.Join(bundlePath, "/config.json")
	specBytes, err := ioutil.ReadFile(specFile)
	if err != nil {
		return nil, err
	}

	var spec spec.Spec
	if err := json.Unmarshal(specBytes, &spec); err != nil {
		return nil, err
	}
	return &spec, nil
}

// writeSpec writes the runtimespec (config.json) to the bundle
func writeSpec(bundlePath string, s *spec.Spec) error {
	specFile := filepath.Join(bundlePath, "/config.json")

	specBytes, err := json.Marshal(s)
	if err != nil {
		return err
	}
	log.Printf("Write Spec bytes\n\n%v", string(specBytes))

	err = ioutil.WriteFile(specFile, specBytes, 0644)

	return err
}

// addRootfsISO creates an ISO from the rootfs of the target spec and adds it
// to the root of the rootfs.
func addRootfsISO(bundlePath string, s *spec.Spec) error {
	rootfsPath := ""

	if s.Root == nil {
		rootfsPath = filepath.Join(bundlePath, "rootfs")
	} else {
		rootfsPath = s.Root.Path
	}

	log.Printf("ISO: Rootfs path determined as %v", rootfsPath)
	isoPath, err := storage.CreateIso(rootfsPath)
	if err != nil {
		log.Printf("ISO: Failed to create ISO")
		return err
	}

	targetISOPath := filepath.Join(rootfsPath, "rootfs.iso")

	log.Printf("ISO: Created ISO %v", isoPath)
	log.Printf("ISO: Target ISO %v", targetISOPath)
	if err = utils.Copy(targetISOPath, isoPath); err != nil {
		return err
	}

	// TODO: Delete old tmp iso or modify storage iso to create iso in specific
	// directory

	return nil
}

// addNablaBinaries adds the nabla runtime binaries to the rootfs
func addNablaBinaries(bundlePath string, s *spec.Spec) error {
	rootfsPath := ""

	if s.Root == nil {
		rootfsPath = filepath.Join(bundlePath, "rootfs")
	} else {
		rootfsPath = s.Root.Path
	}

	binSrcPath := "/usr/local/bin"
	libSrcPath := "/opt/runnc/lib"

	ukvmBinSrcPath := filepath.Join(binSrcPath, "nabla-run")
	nablaRunSrcPath := filepath.Join(binSrcPath, "runnc-cont")

	ukvmBinDstPath := filepath.Join(rootfsPath, "nabla-run")
	nablaRunDstPath := filepath.Join(rootfsPath, "runnc-cont")
	libDstPath := filepath.Join(rootfsPath, "/lib64")

	if err := utils.Copy(ukvmBinDstPath, ukvmBinSrcPath); err != nil {
		return err
	}

	if err := utils.Copy(nablaRunDstPath, nablaRunSrcPath); err != nil {
		return err
	}

	err := utils.Copy(libDstPath, libSrcPath)

	return err
}

func addNetAdmin(s *spec.Spec) error {
	if s.Process == nil {
		return fmt.Errorf("Spec process is nil")
	}

	if s.Process.Capabilities == nil {
		return fmt.Errorf("Spec process capabilities is nil")
	}

	s.Process.Capabilities.Bounding = utils.AddAbsentSlice(s.Process.Capabilities.Bounding, linuxPermCapNetAdmin)
	s.Process.Capabilities.Effective = utils.AddAbsentSlice(s.Process.Capabilities.Effective, linuxPermCapNetAdmin)
	s.Process.Capabilities.Inheritable = utils.AddAbsentSlice(s.Process.Capabilities.Inheritable, linuxPermCapNetAdmin)
	s.Process.Capabilities.Permitted = utils.AddAbsentSlice(s.Process.Capabilities.Permitted, linuxPermCapNetAdmin)
	s.Process.Capabilities.Ambient = utils.AddAbsentSlice(s.Process.Capabilities.Ambient, linuxPermCapNetAdmin)

	return nil
}

func modEntrypoint(s *spec.Spec) error {
	if s.Process == nil {
		return fmt.Errorf("Spec process is nil")
	}

	if len(s.Process.Args) == 0 {
		return fmt.Errorf("OCI process args are empty")
	}

	if !strings.HasSuffix(s.Process.Args[0], ".nabla") {
		return fmt.Errorf("Entrypoint is not a .nabla file")
	}

	args := []string{"/runnc-cont", "-docker",
		"-cwd", s.Process.Cwd,
		"-volume", "/rootfs.iso:/",
		"-unikernel", s.Process.Args[0]}

	for _, e := range s.Process.Env {
		args = append(args, "-env", e)
	}

	args = append(args, "--")
	args = append(args, s.Process.Args[1:]...)

	s.Process.Args = args

	return nil

}

func checkHostNamespace(s *spec.Spec) error {
	if s.Linux == nil {
		return fmt.Errorf("Not a Linux Process")
	}

	if s.Linux.Namespaces == nil {
		return fmt.Errorf("No namespace object")
	}

	for _, v := range s.Linux.Namespaces {
		if v.Type == "network" && strings.Contains(v.Path, "default") {
			return fmt.Errorf("Unable to use host network namespace")
		}
	}

	return nil
}

func modDevicePermissions(s *spec.Spec) error {

	if s.Linux == nil || s.Linux.Resources == nil {
		return fmt.Errorf("Spec linux.resources is empty")
	}

	devs := []spec.LinuxDeviceCgroup{
		{
			Allow:  true,
			Access: "rwm",
		},
	}

	s.Linux.Resources.Devices = devs

	return nil
}

// bundleMod modifies the bundle (config.json and associated rootfs)
// of the OCI runtime spec for the use with nabla containers
func bundleMod(bundlePath string) error {
	// Read the spec
	spec, err := readSpec(bundlePath)
	if err != nil {
		return err
	}

	// Check for use of certain host namespaces
	if err = checkHostNamespace(spec); err != nil {
		return err
	}

	// Modify the spec to add caps
	log.Printf("Adding NET_ADMIN to spec")
	if err = addNetAdmin(spec); err != nil {
		return err
	}

	if err = modDevicePermissions(spec); err != nil {
		return err
	}

	// Modify the spec to change entrypoint to launcher
	log.Printf("Modifying entrypoint")
	if err = modEntrypoint(spec); err != nil {
		return err
	}

	// Create an ISO from the rootfs
	log.Printf("Adding ISO to Rootfs")
	if err = addRootfsISO(bundlePath, spec); err != nil {
		return err
	}

	// Copy the necessary runtime binaries into the rootfs
	log.Printf("Copying runtime binaries into rootfs")
	if err = addNablaBinaries(bundlePath, spec); err != nil {
		return err
	}

	// Output the spec
	err = writeSpec(bundlePath, spec)

	return err
}

func main() {
	// Create logfile (temp fix)
	LogFile, err := os.OpenFile(LogFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer LogFile.Close()

	log.SetOutput(LogFile)
	log.Println("This is a test log entry")

	// Get arguments and env to pass onto runc later
	args := os.Args
	log.Printf("Runnc called with args: %v\n", args)
	env := os.Environ()

	Config.RuncBinaryPath = DefaultRuncBinaryPath

	for i, v := range args {
		if v == "--bundle" {
			if i+1 < len(args) {
				bundlePath := args[i+1]
				log.Printf("Bundle: %v\n\n", bundlePath)
				if err := bundleMod(bundlePath); err != nil {
					log.Printf("ERROR: %v", err)
					panic(err)
				}
				break
			} else {
				panic("Unable to parse bundle")
			}
		}
	}

	runRunc(args, env)
}
