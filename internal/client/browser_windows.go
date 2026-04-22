package client

import (
	"os"
	"path/filepath"
)

func findBrowserExecutable() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	candidates := []string{
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
	}
	if localAppData != "" {
		candidates = append(candidates,
			filepath.Join(localAppData, `Google\Chrome Dev\Application\chrome.exe`),
			filepath.Join(localAppData, `Google\Chrome SxS\Application\chrome.exe`),
		)
	}
	candidates = append(candidates,
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\BraveSoftware\Brave-Browser\Application\brave.exe`,
	)
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}
