package tikdog_watcher

import (
	"io"
	"os/exec"
	"strconv"
	"strings"
)

func (t *watcherManager) killCmd(cmd *exec.Cmd) (pid int, err error) {
	pid = cmd.Process.Pid
	// https://stackoverflow.com/a/44551450
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(pid))
	return pid, kill.Run()
}

func (t *watcherManager) startCmd(cmd string) (*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	var err error

	if !strings.Contains(cmd, ".exe") {
		t.runnerLog("CMD will not recognize non .exe file for execution, path: %s", cmd)
	}

	c := exec.Command("cmd", "/c", cmd)
	stderr, err := c.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	err = c.Start()
	if err != nil {
		return nil, nil, nil, err
	}
	return c, stdout, stderr, err
}
