// Copyright 2014 Docker, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"github.com/opencontainers/runc/libcontainer/utils"
	"github.com/urfave/cli"
	"os"
	"time"
)

var stateCommand = cli.Command{
	Name:  "state",
	Usage: "output the state of a container",
	ArgsUsage: `<container-id>

Where "<container-id>" is your name for the instance of the container.`,
	Description: `The state command outputs current state information for the
instance of a container.`,
	Action: func(context *cli.Context) error {
		// TODO(runllc): Inject LLC
		llc := MyLLC

		// TODO: implement
		container, err := getContainer(context, llc)
		if err != nil {
			fatal(err)
		}

		state, err := container.State()
		if err != nil {
			fatal(err)
		}

		status, err := container.Status()
		if err != nil {
			fatal(err)
		}

		//bundle, annotations := utils.Annotations(state.Config.Labels)
		cs := containerState{
			Version:        state.BaseState.Config.Version,
			ID:             state.BaseState.ID,
			InitProcessPid: state.BaseState.InitProcessPid,
			Status:         status.String(),
			Bundle:         utils.SearchLabels(state.Config.Labels, "bundle"),
			Rootfs:         state.BaseState.Config.Rootfs,
			Created:        state.BaseState.Created,
		}
		data, err := json.MarshalIndent(cs, "", "  ")
		if err != nil {
			fatal(err)
		}
		os.Stdout.Write(data)

		// DEBUG
		//os.Stderr.Write(data)

		return nil
	},
}

// containerState represents the platform agnostic pieces relating to a
// running container's status and state
type containerState struct {
	// Version is the OCI version for the container
	Version string `json:"ociVersion"`
	// ID is the container ID
	ID string `json:"id"`
	// InitProcessPid is the init process id in the parent namespace
	InitProcessPid int `json:"pid"`
	// Status is the current status of the container, running, paused, ...
	Status string `json:"status"`
	// Bundle is the path on the filesystem to the bundle
	Bundle string `json:"bundle"`
	// Rootfs is a path to a directory containing the container's root filesystem.
	Rootfs string `json:"rootfs"`
	// Created is the unix timestamp for the creation time of the container in UTC
	Created time.Time `json:"created"`
	// Annotations is the user defined annotations added to the config.
	Annotations map[string]string `json:"annotations,omitempty"`
}
