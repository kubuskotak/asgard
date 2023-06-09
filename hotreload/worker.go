// Package hotreload is func library that implement reload service on development stage.
// # This manifest was generated by ymir. DO NOT EDIT.
package hotreload

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog/log"

	"github.com/kubuskotak/asgard/common"
)

// Worker abstract of worker instances.
type Worker interface {
	Run() error
	Reload()
}

type worker struct {
	workdir string
	args    []string
	command *exec.Cmd

	reload chan bool
	round  uint64
	mu     sync.RWMutex
}

// NewWorker create new worker instance.
func NewWorker(dir string, args ...string) Worker {
	return &worker{
		workdir: dir,
		args:    args,
		reload:  make(chan bool),
	}
}

// Reload reload worker.
func (w *worker) Reload() {
	w.mu.Lock()
	close(w.reload)
	w.reload = make(chan bool)
	w.mu.Unlock()
}

// Run return error to start worker.
func (w *worker) Run() error {
	_, err := exec.LookPath("task")
	if err != nil {
		panic(err.Error())
	}
	var (
		pidKill = make(chan struct{})
		wg      = sync.WaitGroup{}
		waitFor = 500 * time.Millisecond
	)
	go func() {
		<-w.reload
		close(pidKill)
		wg.Wait()
	}()
	var funKill = func(
		cmd *exec.Cmd,
		stdout io.ReadCloser, stderr io.ReadCloser,
		pkill chan struct{}, pExit chan struct{},
		wg *sync.WaitGroup,
	) {
		defer wg.Done()
		select {
		case <-pkill:
			break
		case <-pExit:
			return
		}
		defer func() {
			_ = stdout.Close()
			_ = stderr.Close()
		}()
		fmt.Println(common.Colorize("+ Command Killing...", common.ColorBlue))
		pid, err := w.killCmd(cmd)
		if err != nil {
			if cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
				os.Exit(1)
			}
		}
		fmt.Println(common.Colorize(fmt.Sprintf("+ Command Killed PID: %d", pid), common.ColorBlue))
	}
	fmt.Println(common.Colorize("+ Hot Reload Running...", common.ColorBlue))
	go func() {
		for {
			select {
			case <-pidKill:
				return
			default:
				var (
					processExit = make(chan struct{})
					args        = make([]string, 0)
				)
				args = append(args, "dev")
				cmd, stdOut, stdErr, err := w.execCmd("task", args...)
				if err != nil {
					log.Error().Err(err).Msg("+ Dev Failed")
				}
				fmt.Println(common.Colorize(fmt.Sprintf("+ Running Process PID: %d ; args: %v", cmd.Process.Pid, args), common.ColorBlue))
				wg.Add(1)
				atomic.AddUint64(&w.round, 1)
				go funKill(cmd, stdOut, stdErr, pidKill, processExit, &wg)
				go func() {
					_, _ = io.Copy(colorable.NewColorable(os.Stdout), stdOut)
				}()
				_, _ = io.Copy(colorable.NewColorable(os.Stderr), stdErr)
				_, _ = cmd.Process.Wait()
				close(processExit)
				time.Sleep(waitFor)
			}
		}
	}()
	return nil
}
