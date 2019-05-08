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

package llcli

import (
	"fmt"

	"github.com/urfave/cli"
)

func newExecCmd() cli.Command {
	return cli.Command{
		Name:  "exec",
		Usage: "execute new process inside the container",
		ArgsUsage: `<container-id> <command> [command options]  || -p process.json <container-id>

Where "<container-id>" is the name for the instance of the container and
"<command>" is the command to be executed in the container.
"<command>" can't be empty unless a "-p" flag provided.

EXAMPLE:
For example, if the container is configured to run the linux ps command the
following will output a list of processes running in the container:

       # {{name}} exec <container-id> ps`,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "console-socket",
				Usage: "path to an AF_UNIX socket which will receive a file descriptor referencing the master end of the console's pseudoterminal",
			},
			cli.StringFlag{
				Name:  "cwd",
				Usage: "current working directory in the container",
			},
			cli.StringSliceFlag{
				Name:  "env, e",
				Usage: "set environment variables",
			},
			cli.BoolFlag{
				Name:  "tty, t",
				Usage: "allocate a pseudo-TTY",
			},
			cli.StringFlag{
				Name:  "user, u",
				Usage: "UID (format: <uid>[:<gid>])",
			},
			cli.Int64SliceFlag{
				Name:  "additional-gids, g",
				Usage: "additional gids",
			},
			cli.StringFlag{
				Name:  "process, p",
				Usage: "path to the process.json",
			},
			cli.BoolFlag{
				Name:  "detach,d",
				Usage: "detach from the container's process",
			},
			cli.StringFlag{
				Name:  "pid-file",
				Value: "",
				Usage: "specify the file to write the process id to",
			},
			cli.StringFlag{
				Name:  "process-label",
				Usage: "set the asm process label for the process commonly used with selinux",
			},
			cli.StringFlag{
				Name:  "apparmor",
				Usage: "set the apparmor profile for the process",
			},
			cli.BoolFlag{
				Name:  "no-new-privs",
				Usage: "set the no new privileges value for the process",
			},
			cli.StringSliceFlag{
				Name:  "cap, c",
				Value: &cli.StringSlice{},
				Usage: "add a capability to the bounding set for the process",
			},
			cli.BoolFlag{
				Name:   "no-subreaper",
				Usage:  "disable the use of the subreaper used to reap reparented processes",
				Hidden: true,
			},
		},
		Action: func(context *cli.Context) error {
			// TODO: implement
			return fmt.Errorf("OCI Exec Not Implemented")
		},
		SkipArgReorder: true,
	}
}
