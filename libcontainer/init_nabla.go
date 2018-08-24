package libcontainer

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"
)

type initConfig struct {
	Args []string `json:"args"`
}

func initNabla() error {
	// Ricardo special
	time.Sleep(1 * time.Second)
	fmt.Println("HELLO CALLING FROM START_INITIALIZATION")

	var (
		pipefd      int
		envInitPipe = os.Getenv("_LIBCONTAINER_INITPIPE")
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

	fmt.Printf("Got config: %v\n", config)

	// clear the current process's environment to clean any libcontainer
	// specific env vars.
	os.Clearenv()

	// TODO: WAIT FOR EXEC.FIFO

	if err := syscall.Exec(config.Args[0], config.Args, os.Environ()); err != nil {
		return err
	}

	return nil

}
