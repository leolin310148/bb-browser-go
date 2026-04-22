package client

import "os/exec"

func findBrowserExecutable() string {
	candidates := []string{"google-chrome", "google-chrome-stable", "chromium-browser", "chromium"}
	for _, c := range candidates {
		if path, err := exec.LookPath(c); err == nil {
			return path
		}
	}
	return ""
}
