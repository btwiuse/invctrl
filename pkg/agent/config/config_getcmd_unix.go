//go:build !windows

package config

import "os/exec"

func (c *Config) getCmd() []string {
	shell := "bash"
	if _, err := exec.LookPath(shell); err != nil {
		shell = "sh"
	}
	args := []string{"env", "TERM=xterm-256color", shell}
	if c.Cmd == "" {
		return args
	}
	return append(args, "-c", c.Cmd)
}
