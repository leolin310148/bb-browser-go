package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/leolin310148/bb-browser-go/internal/client"
	"github.com/leolin310148/bb-browser-go/internal/config"
	"github.com/leolin310148/bb-browser-go/internal/protocol"
)

// doctorCheck is one row of the doctor output.
type doctorCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "ok", "warn", "fail"
	Detail string `json:"detail,omitempty"`
}

// runDoctor performs end-to-end diagnostics for the CLI/daemon/browser stack.
// Exits 0 when all checks pass, 1 when any check fails (warn does not fail).
func runDoctor(jsonOutput bool) {
	checks := []doctorCheck{}

	checks = append(checks, doctorCheck{
		Name:   "Binary",
		Status: "ok",
		Detail: fmt.Sprintf("bb-browser-go %s", version),
	})

	checks = append(checks, checkHomeDir())

	info, infoCheck := checkDaemonJSON()
	checks = append(checks, infoCheck)

	if info != nil {
		checks = append(checks, checkDaemonProcess(info))
		statusRaw, statusCheck := checkDaemonHTTP()
		checks = append(checks, statusCheck)
		if statusRaw != nil {
			checks = append(checks, checkCDPConnected(statusRaw))
			checks = append(checks, checkTabs())
		}
	}

	checks = append(checks, checkCDPDiscovery())

	if jsonOutput {
		emitDoctorJSON(checks)
	} else {
		emitDoctorText(checks)
	}

	for _, c := range checks {
		if c.Status == "fail" {
			os.Exit(1)
		}
	}
}

func checkHomeDir() doctorCheck {
	dir := config.HomeDir()
	st, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return doctorCheck{Name: "Home dir", Status: "warn", Detail: dir + " (does not exist yet)"}
		}
		return doctorCheck{Name: "Home dir", Status: "fail", Detail: err.Error()}
	}
	if !st.IsDir() {
		return doctorCheck{Name: "Home dir", Status: "fail", Detail: dir + " is not a directory"}
	}
	return doctorCheck{Name: "Home dir", Status: "ok", Detail: dir}
}

func checkDaemonJSON() (*protocol.DaemonInfo, doctorCheck) {
	info, err := client.ReadDaemonJSON()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, doctorCheck{Name: "Daemon JSON", Status: "warn", Detail: "daemon not started (no daemon.json)"}
		}
		return nil, doctorCheck{Name: "Daemon JSON", Status: "fail", Detail: err.Error()}
	}
	return info, doctorCheck{
		Name:   "Daemon JSON",
		Status: "ok",
		Detail: fmt.Sprintf("%s:%d (pid %d)", info.Host, info.Port, info.PID),
	}
}

func checkDaemonProcess(info *protocol.DaemonInfo) doctorCheck {
	if !client.IsProcessAlive(info.PID) {
		return doctorCheck{
			Name:   "Daemon process",
			Status: "fail",
			Detail: fmt.Sprintf("pid %d is gone — stale daemon.json (run 'bb-browser daemon stop' or delete %s)", info.PID, config.DaemonJSONPath()),
		}
	}
	return doctorCheck{Name: "Daemon process", Status: "ok", Detail: fmt.Sprintf("pid %d alive", info.PID)}
}

func checkDaemonHTTP() (json.RawMessage, doctorCheck) {
	raw, err := client.GetDaemonStatus()
	if err != nil {
		return nil, doctorCheck{Name: "Daemon HTTP", Status: "fail", Detail: err.Error()}
	}
	return raw, doctorCheck{Name: "Daemon HTTP", Status: "ok", Detail: "/status responsive"}
}

func checkCDPConnected(raw json.RawMessage) doctorCheck {
	var st protocol.DaemonStatus
	if err := json.Unmarshal(raw, &st); err != nil {
		return doctorCheck{Name: "CDP connected", Status: "warn", Detail: "could not parse /status payload"}
	}
	if !st.CDPConnected {
		return doctorCheck{
			Name:   "CDP connected",
			Status: "fail",
			Detail: "daemon is up but not attached to Chrome — start the browser or check BB_BROWSER_CDP_URL",
		}
	}
	return doctorCheck{Name: "CDP connected", Status: "ok", Detail: "daemon attached to Chrome"}
}

func checkCDPDiscovery() doctorCheck {
	ep, err := client.DiscoverCDPPort()
	if err != nil {
		return doctorCheck{
			Name:   "CDP discovery",
			Status: "warn",
			Detail: "no Chrome reachable from this CLI (" + err.Error() + ")",
		}
	}
	return doctorCheck{
		Name:   "CDP discovery",
		Status: "ok",
		Detail: fmt.Sprintf("%s:%d reachable", ep.Host, ep.Port),
	}
}

func checkTabs() doctorCheck {
	req := &protocol.Request{ID: newID(), Action: protocol.ActionTabList}
	resp, err := client.SendCommand(req)
	if err != nil {
		return doctorCheck{Name: "Tabs", Status: "warn", Detail: err.Error()}
	}
	if !resp.Success {
		return doctorCheck{Name: "Tabs", Status: "warn", Detail: resp.Error}
	}
	if resp.Data == nil || len(resp.Data.Tabs) == 0 {
		return doctorCheck{Name: "Tabs", Status: "warn", Detail: "no open tabs (open one with 'bb-browser open <url>')"}
	}
	return doctorCheck{Name: "Tabs", Status: "ok", Detail: fmt.Sprintf("%d open", len(resp.Data.Tabs))}
}

func emitDoctorText(checks []doctorCheck) {
	maxName := 0
	for _, c := range checks {
		if len(c.Name) > maxName {
			maxName = len(c.Name)
		}
	}
	failed := 0
	warned := 0
	for _, c := range checks {
		marker := "[OK]"
		switch c.Status {
		case "fail":
			marker = "[FAIL]"
			failed++
		case "warn":
			marker = "[WARN]"
			warned++
		}
		pad := strings.Repeat(" ", maxName-len(c.Name))
		if c.Detail != "" {
			fmt.Printf("  %s%s  %-6s %s\n", c.Name, pad, marker, c.Detail)
		} else {
			fmt.Printf("  %s%s  %s\n", c.Name, pad, marker)
		}
	}
	fmt.Println()
	switch {
	case failed > 0:
		fmt.Printf("%d failed, %d warning(s) — see above.\n", failed, warned)
	case warned > 0:
		fmt.Printf("All required checks passed (%d warning(s)).\n", warned)
	default:
		fmt.Println("All checks passed.")
	}
}

func emitDoctorJSON(checks []doctorCheck) {
	failed := false
	for _, c := range checks {
		if c.Status == "fail" {
			failed = true
			break
		}
	}
	out := struct {
		OK     bool          `json:"ok"`
		Checks []doctorCheck `json:"checks"`
	}{OK: !failed, Checks: checks}
	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(b))
}
