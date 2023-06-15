// Package hotreload is func library that implement reload service on development stage.
// # This manifest was generated by ymir. DO NOT EDIT.
package hotreload

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

func (w *worker) execCmd(command string, options ...string) (
	*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	var err error
	// Checking for nil.
	if command == "" || options == nil {
		return nil, nil, nil, fmt.Errorf("no command to execute")
	}

	// Collect command line.
	var args = append([]string{"-c", command}, options...)
	cmd := exec.Command("/bin/sh", args...) // #nosec G204
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Start(); err != nil {
		return nil, nil, nil, err
	}
	return cmd, stdOut, stdErr, err
}

func (w *worker) killCmd(cmd *exec.Cmd) (pid int, err error) {
	pid = cmd.Process.Pid
	err = syscall.Kill(-pid, syscall.SIGKILL)
	_, _ = cmd.Process.Wait()
	return
}
