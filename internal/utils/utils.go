package utils

import (
	"bufio"
	"github.com/pkg/errors"
	"os/exec"
	"sync"
)

// RunCmd runs a command and returns stdout and stderr channels. It returns an error channel
// that is filled on cmd.Wait() so that consuming from it waiting for the command to complete
// is safe.
func RunCmd(command string, args ...string) (<-chan string, <-chan string, <-chan error) {
	cmd := exec.Command(command, args...)

	errCh := make(chan error)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errCh <- errors.Wrap(err, "error creating stdout pipe")
		return nil, nil, errCh
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		errCh <- errors.Wrap(err, "error creating stderr pipe")
		return nil, nil, errCh
	}

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	var wg sync.WaitGroup
	stdoutCh := make(chan string)
	stderrCh := make(chan string)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for stdoutScanner.Scan() {
			text := stdoutScanner.Text()
			stdoutCh <- text
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for stderrScanner.Scan() {
			text := stderrScanner.Text()
			stderrCh <- text
		}
	}()

	if err := cmd.Start(); err != nil {
		close(stdoutCh)
		close(stderrCh)
		errCh <- errors.Wrapf(err, "error starting the %s command", command)
		close(errCh)
		return nil, nil, errCh
	}

	go func() {
		errCh <- cmd.Wait()
		close(errCh)
		wg.Wait()
		close(stdoutCh)
		close(stderrCh)
	}()

	return stdoutCh, stderrCh, errCh
}
