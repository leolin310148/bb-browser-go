package mcp

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

// Run starts the MCP server over stdio.
func Run(version string) {
	s := server.NewMCPServer(
		"bb-browser",
		version,
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	// Navigation
	s.AddTool(navigateTool, handleNavigate)
	s.AddTool(backTool, handleBack)
	s.AddTool(forwardTool, handleForward)
	s.AddTool(refreshTool, handleRefresh)
	s.AddTool(closeTool, handleClose)

	// Interaction
	s.AddTool(clickTool, handleClick)
	s.AddTool(hoverTool, handleHover)
	s.AddTool(fillTool, handleFill)
	s.AddTool(typeTool, handleType)
	s.AddTool(checkTool, handleCheck)
	s.AddTool(uncheckTool, handleUncheck)
	s.AddTool(selectTool, handleSelect)
	s.AddTool(pressTool, handlePress)
	s.AddTool(scrollTool, handleScroll)

	// Observation
	s.AddTool(snapshotTool, handleSnapshot)
	s.AddTool(screenshotTool, handleScreenshot)
	s.AddTool(getTool, handleGet)
	s.AddTool(evalTool, handleEval)
	s.AddTool(waitTool, handleWait)

	// Tab Management
	s.AddTool(tabListTool, handleTabList)
	s.AddTool(tabNewTool, handleTabNew)
	s.AddTool(tabSelectTool, handleTabSelect)
	s.AddTool(tabCloseTool, handleTabClose)

	// Diagnostics
	s.AddTool(networkTool, handleNetwork)
	s.AddTool(consoleTool, handleConsole)
	s.AddTool(errorsTool, handleErrors)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}
