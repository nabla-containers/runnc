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
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/nabla-containers/runnc/libcontainer"
	ll "github.com/nabla-containers/runnc/llif"
	"github.com/urfave/cli"
)

func newDeleteCmd(llcHandler ll.RunllcHandler) cli.Command {
	return cli.Command{
		Name:  "delete",
		Usage: "delete any resources held by the container often used with detached container",
		ArgsUsage: `<container-id>

Where "<container-id>" is the name for the instance of the container.

EXAMPLE:
For example, if the container id is "ubuntu01" and runnc list currently shows the
status of "ubuntu01" as "stopped" the following will delete resources held for
"ubuntu01" removing "ubuntu01" from the runnc list of containers:

       # runnc delete ubuntu01`,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Forcibly deletes the container if it is still running (uses SIGKILL)",
			},
		},
		Action: func(context *cli.Context) error {
			id := context.Args().First()
			container, err := getContainer(context, llcHandler)
			if err != nil {
				if lerr, ok := err.(libcontainer.Error); ok && lerr.Code() == libcontainer.ContainerNotExists {
					// if there was an aborted start or something of the sort then the container's       directory could exist but
					// libcontainer does not see it because the state.json file inside that directory    was never created.
					path := filepath.Join(context.GlobalString("root"), id)
					if e := os.RemoveAll(path); e != nil {
						//	fmt.Fprintf(os.Stderr, "remove %s: %v\n", path, e)
					}
				}
				return err
			}
			s, err := container.Status()
			if err != nil {
				return err
			}
			switch s {
			case libcontainer.Stopped:
				destroy(container)
			case libcontainer.Created:
				return killContainer(container)
			default:
				if context.Bool("force") {
					return killContainer(container)
				}
				return fmt.Errorf("cannot delete container %s that is not stopped: %s\n", id, s)
			}

			return nil
		},
	}
}

func killContainer(container libcontainer.Container) error {
	_ = container.Signal(syscall.SIGKILL, false)
	for i := 0; i < 100; i++ {
		time.Sleep(100 * time.Millisecond)
		if err := container.Signal(syscall.Signal(0), false); err != nil {
			destroy(container)
			return nil
		}
	}
	return fmt.Errorf("container init still running")
}
