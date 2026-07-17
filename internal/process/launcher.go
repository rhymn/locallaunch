package process

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type Request struct {
	Path string   `json:"path"`
	Args []string `json:"args"`
	Cwd  string   `json:"cwd"`
}

func Launch(req *Request) (int, error) {
	if req.Path == "" {
		return 0, fmt.Errorf("path is required")
	}

	var cmd *exec.Cmd

	if runtime.GOOS == "darwin" && strings.HasSuffix(req.Path, ".app") {
		args := append([]string{req.Path}, req.Args...)
		cmd = exec.Command("open", args...)
	} else {
		cmd = exec.Command(req.Path, req.Args...)
	}

	if req.Cwd != "" {
		cmd.Dir = req.Cwd
	}

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("starting process: %w", err)
	}

	return cmd.Process.Pid, nil
}
