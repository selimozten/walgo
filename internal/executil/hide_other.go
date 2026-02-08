//go:build !windows

package executil

import "os/exec"

func hideWindow(_ *exec.Cmd) {}
