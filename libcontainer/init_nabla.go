package libcontainer

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"syscall"
)

type initConfig struct {
	Args []string `json:"args"`
}

func initNabla() error {
	var (
		pipefd, rootfd int
		envInitPipe    = os.Getenv("_LIBCONTAINER_INITPIPE")
		envStateDir    = os.Getenv("_LIBCONTAINER_STATEDIR")
	)

	// Get the INITPIPE.
	pipefd, err := strconv.Atoi(envInitPipe)
	if err != nil {
		return fmt.Errorf("unable to convert _LIBCONTAINER_INITPIPE=%s to int: %s", envInitPipe, err)
	}

	fmt.Println("Reading config from parent:")
	pipe := os.NewFile(uintptr(pipefd), "pipe")
	defer pipe.Close()

	var config *initConfig
	if err := json.NewDecoder(pipe).Decode(&config); err != nil {
		return err
	}

	// Only init processes have STATEDIR.
	if rootfd, err = strconv.Atoi(envStateDir); err != nil {
		return fmt.Errorf("unable to convert _LIBCONTAINER_STATEDIR=%s to int: %s", envStateDir, err)
	}

	// clear the current process's environment to clean any libcontainer
	// specific env vars.
	os.Clearenv()

	// wait for the fifo to be opened on the other side before
	// exec'ing the users process.
	fd, err := syscall.Openat(rootfd, execFifoFilename, os.O_WRONLY|syscall.O_CLOEXEC, 0)
	if err != nil {
		return newSystemErrorWithCause(err, "openat exec fifo")
	}
	if _, err := syscall.Write(fd, []byte("0")); err != nil {
		return newSystemErrorWithCause(err, "write 0 exec fifo")
	}
	syscall.Close(rootfd)

	if err := syscall.Exec(config.Args[0], config.Args, os.Environ()); err != nil {
		return err
	}

	return nil

}
