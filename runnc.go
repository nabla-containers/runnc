// Copyright (c) 2018, IBM
// Author(s): Brandon Lum, Ricardo Koller
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
	spec "github.com/opencontainers/runtime-spec/specs-go"
	"github.ibm.com/nabla-containers/nabla-lib/storage"
	"github.ibm.com/nabla-containers/runnc/utils"
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
	LogFile               *os.File = nil
	LogFilePath           string   = "/tmp/runnc.log"
	Config                         = config{}
	DefaultRuncBinaryPath          = "/usr/bin/docker-runc"
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

var CAP_NET_ADMIN = "CAP_NET_ADMIN"

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

	if err = ioutil.WriteFile(specFile, specBytes, 0644); err != nil {
		return err
	}

	return nil
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
	ukvmBinSrcPath := filepath.Join(binSrcPath, "ukvm-bin")
	nablaRunSrcPath := filepath.Join(binSrcPath, "runnc-cont")

	// TODO: Add checks for file exists?
	ukvmBinDstPath := filepath.Join(rootfsPath, "ukvm-bin")
	nablaRunDstPath := filepath.Join(rootfsPath, "runnc-cont")

	if err := utils.Copy(ukvmBinDstPath, ukvmBinSrcPath); err != nil {
		return err
	}

	if err := utils.Copy(nablaRunDstPath, nablaRunSrcPath); err != nil {
		return err
	}

	return nil
}

func addAbsentSlice(slice []string, add string) []string {
	for _, v := range slice {
		if add == v {
			return slice
		}
	}

	return append(slice, add)
}

func addNetAdmin(s *spec.Spec) error {
	if s.Process == nil {
		return fmt.Errorf("Spec process is nil")
	}

	if s.Process.Capabilities == nil {
		return fmt.Errorf("Spec process capabilities is nil")
	}

	s.Process.Capabilities.Bounding = addAbsentSlice(s.Process.Capabilities.Bounding, CAP_NET_ADMIN)
	s.Process.Capabilities.Effective = addAbsentSlice(s.Process.Capabilities.Effective, CAP_NET_ADMIN)
	s.Process.Capabilities.Inheritable = addAbsentSlice(s.Process.Capabilities.Inheritable, CAP_NET_ADMIN)
	s.Process.Capabilities.Permitted = addAbsentSlice(s.Process.Capabilities.Permitted, CAP_NET_ADMIN)
	s.Process.Capabilities.Ambient = addAbsentSlice(s.Process.Capabilities.Ambient, CAP_NET_ADMIN)

	return nil
}

func modEntrypoint(s *spec.Spec) error {
	if s.Process == nil {
		return fmt.Errorf("Spec process is nil")
	}

	if len(s.Process.Args) == 0 {
		return fmt.Errorf("OCI process args are empty")
	}

	// Set cwd to root
	if s.Process.Cwd != "/" {
		log.Printf("Currently, CWD is not supported, ignoring and setting to /")
	}
	s.Process.Cwd = "/"

	args := append([]string{"/runnc-cont", "-docker",
		"-volume", "/rootfs.iso:/",
		"-unikernel", s.Process.Args[0], "--"},
		s.Process.Args[1:]...)

	s.Process.Args = args

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

	// Modify the spec to add caps
	log.Printf("Adding NET_ADMIN to spec")
	if err = addNetAdmin(spec); err != nil {
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
	if err = writeSpec(bundlePath, spec); err != nil {
		return err
	}

	return nil
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
