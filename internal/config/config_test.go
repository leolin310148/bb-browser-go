package config

import (
	"path/filepath"
	"testing"
)

func TestHomeDir_EnvOverride(t *testing.T) {
	t.Setenv("BB_BROWSER_HOME", "/tmp/bb-override")
	if got := HomeDir(); got != "/tmp/bb-override" {
		t.Fatalf("HomeDir with override = %q, want /tmp/bb-override", got)
	}
}

func TestHomeDir_Default(t *testing.T) {
	t.Setenv("BB_BROWSER_HOME", "")
	t.Setenv("HOME", "/tmp/fakehome")
	want := filepath.Join("/tmp/fakehome", ".bb-browser")
	if got := HomeDir(); got != want {
		t.Fatalf("HomeDir default = %q, want %q", got, want)
	}
}

func TestDerivedPaths(t *testing.T) {
	t.Setenv("BB_BROWSER_HOME", "/tmp/bb")

	cases := []struct {
		name string
		fn   func() string
		want string
	}{
		{"DaemonJSONPath", DaemonJSONPath, "/tmp/bb/daemon.json"},
		{"SitesDir", SitesDir, "/tmp/bb/sites"},
		{"CommunitySitesDir", CommunitySitesDir, "/tmp/bb/bb-sites"},
		{"ManagedBrowserDir", ManagedBrowserDir, "/tmp/bb/browser"},
		{"ManagedPortFile", ManagedPortFile, "/tmp/bb/browser/cdp-port"},
		{"ManagedUserDataDir", ManagedUserDataDir, "/tmp/bb/browser/user-data"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.fn(); got != c.want {
				t.Fatalf("%s = %q, want %q", c.name, got, c.want)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if DaemonPort != 19824 {
		t.Errorf("DaemonPort = %d, want 19824", DaemonPort)
	}
	if DaemonHost != "127.0.0.1" {
		t.Errorf("DaemonHost = %q, want 127.0.0.1", DaemonHost)
	}
	if CommandTimeout != 30 {
		t.Errorf("CommandTimeout = %d, want 30", CommandTimeout)
	}
	if DefaultCDPPort != 19825 {
		t.Errorf("DefaultCDPPort = %d, want 19825", DefaultCDPPort)
	}
}
