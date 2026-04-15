//go:build windows

package client

import "os/exec"

func setDetached(cmd *exec.Cmd) {
	// On Windows, no Setpgid equivalent needed for basic detachment.
}
