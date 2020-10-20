package tikdog_watcher

import (
	"github.com/creack/pty"
	"io"
	"os/exec"
	"syscall"
)

func (t *watcherManager) killCmd(cmd *exec.Cmd) (pid int, err error) {
	pid = cmd.Process.Pid

	// https://stackoverflow.com/questions/22470193/why-wont-go-kill-a-child-process-correctly
	err = syscall.Kill(-pid, syscall.SIGKILL)
	// Wait releases any resources associated with the Process.
	_, _ = cmd.Process.Wait()
	return pid, err
}

func (t *watcherManager) startCmd(cmd string) (*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	c := exec.Command("/bin/sh", "-c", cmd)
	f, err := pty.Start(c)
	return c, f, f, err
}
