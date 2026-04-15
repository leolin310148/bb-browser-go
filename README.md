# bb-browser-go

**Your browser is the API.** A CLI tool that lets you control and observe any Chromium-based browser from the terminal via the Chrome DevTools Protocol (CDP).

`bb-browser-go` is a Go port of [bb-browser](https://github.com/nicepkg/bb-browser) (Node.js). It ships as a single static binary with zero runtime dependencies.

## Installation

### Download prebuilt binary

Grab the latest release for your platform from [GitHub Releases](https://github.com/leolin310148/bb-browser-go/releases):

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/leolin310148/bb-browser-go/releases/latest/download/bb-browser-darwin-arm64
chmod +x bb-browser-darwin-arm64
sudo mv bb-browser-darwin-arm64 /usr/local/bin/bb-browser

# macOS (Intel)
curl -LO https://github.com/leolin310148/bb-browser-go/releases/latest/download/bb-browser-darwin-amd64
chmod +x bb-browser-darwin-amd64
sudo mv bb-browser-darwin-amd64 /usr/local/bin/bb-browser

# Linux (x86_64)
curl -LO https://github.com/leolin310148/bb-browser-go/releases/latest/download/bb-browser-linux-amd64
chmod +x bb-browser-linux-amd64
sudo mv bb-browser-linux-amd64 /usr/local/bin/bb-browser

# Linux (ARM64)
curl -LO https://github.com/leolin310148/bb-browser-go/releases/latest/download/bb-browser-linux-arm64
chmod +x bb-browser-linux-arm64
sudo mv bb-browser-linux-arm64 /usr/local/bin/bb-browser
```

### Build from source

```bash
go install github.com/leolin310148/bb-browser-go@latest
```

Or clone and build:

```bash
git clone https://github.com/leolin310148/bb-browser-go.git
cd bb-browser-go
go build -o bb-browser .
```

## Prerequisites

You need a Chromium-based browser (Google Chrome, Microsoft Edge, Brave, Arc, etc.) installed on your machine.

`bb-browser` connects to the browser using CDP. It will automatically:

1. Detect a running browser with remote debugging enabled
2. Or launch a managed browser instance for you

If you prefer manual control, start Chrome with debugging enabled:

```bash
# macOS
"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --remote-debugging-port=19825

# Linux
google-chrome --remote-debugging-port=19825

# Windows
"C:\Program Files\Google\Chrome\Application\chrome.exe" --remote-debugging-port=19825
```

Or point to a remote browser via environment variable:

```bash
export BB_BROWSER_CDP_URL=http://127.0.0.1:19825
```

## How It Works

```
┌──────────┐       HTTP        ┌──────────┐     WebSocket/CDP     ┌──────────┐
│  bb-cli  │  ───────────────> │  daemon   │  ──────────────────>  │  Chrome  │
│ (client) │  <─────────────── │ (server)  │  <──────────────────  │ (browser)│
└──────────┘    JSON response  └──────────┘     DevTools Protocol  └──────────┘
```

When you run any command, `bb-browser`:

1. **Starts a daemon** (if not already running) that holds a persistent CDP WebSocket connection to Chrome
2. **Sends the command** as an HTTP request to the daemon
3. **The daemon translates** the command into CDP protocol calls and returns the result

The daemon runs in the background and auto-discovers your browser. You don't need to manage it manually.

## Quick Start

```bash
# Open a webpage
bb-browser open https://example.com

# Take a snapshot of the page (accessibility tree with element references)
bb-browser snapshot

# Click an element by its ref number from the snapshot
bb-browser click 5

# Fill a text input
bb-browser fill 3 "hello world"

# Get the page title
bb-browser get title

# Take a screenshot
bb-browser screenshot

# Execute JavaScript in the page
bb-browser eval "document.title"

# Get JSON output for scripting
bb-browser snapshot --json

# Filter with jq expressions
bb-browser snapshot --jq ".snapshotData.refs | keys | length"
```

## Commands

### Navigation

| Command | Description |
|---------|-------------|
| `open <url>` | Open a URL (creates a new tab, or navigates current tab with `--tab`) |
| `back` | Navigate back in history |
| `forward` | Navigate forward in history |
| `refresh` | Reload the current page |
| `close` | Close the current tab |

```bash
# Open a URL in a new tab
bb-browser open https://github.com

# Open in a specific existing tab
bb-browser open https://github.com --tab ab1c

# Navigate back
bb-browser back
```

### Observation

These commands let you **see** what's on the page.

#### `snapshot` - Get the accessibility tree

The most important command. It returns a structured text representation of the page with **ref numbers** you can use to interact with elements.

```bash
# Full accessibility tree
bb-browser snapshot

# Interactive elements only (buttons, links, inputs, etc.)
bb-browser snapshot -i

# Compact output (shorter names, no tag names)
bb-browser snapshot -c

# Limit tree depth
bb-browser snapshot -d 3

# Filter by selector/keyword
bb-browser snapshot -s "search"

# Combine flags
bb-browser snapshot -i -c
```

Example output:

```
- navigation <nav>
  - link [ref=0] "Home" <a>
  - link [ref=1] "About" <a>
  - link [ref=2] "Contact" <a>
- main <main>
  - heading "Welcome" <h1>
  - textbox [ref=3] "Search..." <input>
  - button [ref=4] "Submit" <button>
```

The `[ref=N]` numbers are what you use with interaction commands like `click`, `fill`, `type`, etc.

#### `screenshot`

```bash
# Capture screenshot (returned as base64 data URL in JSON)
bb-browser screenshot

# Save to file (use with --json and jq)
bb-browser screenshot --json --jq ".data.dataUrl"
```

#### `get` - Get element or page attributes

```bash
# Get the current page URL
bb-browser get url

# Get the page title
bb-browser get title

# Get the text content of an element
bb-browser get text 5

# Get an HTML attribute of an element
bb-browser get href 2
bb-browser get class 4
bb-browser get value 3
```

#### `network` - Monitor network requests

```bash
# List all captured network requests
bb-browser network

# Filter by URL pattern
bb-browser network requests --filter "api"

# Filter by HTTP method
bb-browser network requests --method POST

# Filter by status code
bb-browser network requests --status 404
bb-browser network requests --status 5xx

# Include response bodies
bb-browser network requests --with-body

# Show only requests since last action
bb-browser network requests --since last_action

# Clear captured requests
bb-browser network clear
```

#### `console` - Read console messages

```bash
# Get all console messages
bb-browser console

# Filter messages
bb-browser console --filter "error"

# Only messages since last action
bb-browser console --since last_action

# Clear console buffer
bb-browser console --clear
```

#### `errors` - Read JavaScript errors

```bash
# Get all JS errors
bb-browser errors

# Filter errors
bb-browser errors --filter "TypeError"

# Clear error buffer
bb-browser errors --clear
```

### Interaction

All interaction commands use **ref numbers** from the `snapshot` output.

#### `click` / `hover`

```bash
# Click a button (ref=4 from snapshot)
bb-browser click 4

# Hover over an element
bb-browser hover 2
```

#### `fill` / `type`

```bash
# Clear the input and fill with new text
bb-browser fill 3 "hello world"

# Append text to current value (like typing)
bb-browser type 3 " more text"
```

#### `check` / `uncheck`

```bash
# Check a checkbox
bb-browser check 7

# Uncheck it
bb-browser uncheck 7
```

#### `select`

```bash
# Select a dropdown option by value
bb-browser select 6 "option2"
```

#### `press`

```bash
# Press a key
bb-browser press Enter
bb-browser press Tab
bb-browser press ArrowDown
bb-browser press Escape
```

#### `scroll`

```bash
# Scroll down (default 300px)
bb-browser scroll down

# Scroll up 500px
bb-browser scroll up 500

# Scroll left/right
bb-browser scroll left 200
bb-browser scroll right 200
```

#### `eval` - Execute JavaScript

```bash
# Run arbitrary JavaScript in the page context
bb-browser eval "document.title"
bb-browser eval "document.querySelectorAll('a').length"
bb-browser eval "window.location.href"

# Multi-word scripts
bb-browser eval "document.querySelector('h1').textContent"

# Async/IIFE
bb-browser eval "(async () => { const r = await fetch('/api/data'); return r.json(); })()"
```

#### `wait`

```bash
# Wait 1 second (default)
bb-browser wait

# Wait 2 seconds
bb-browser wait 2000
```

### Tab Management

```bash
# List all open tabs
bb-browser tab

# Open a new tab
bb-browser tab new
bb-browser tab new https://google.com

# Switch to tab by index
bb-browser tab 0
bb-browser tab 2

# Switch to tab by short ID
bb-browser tab select ab1c

# Close a tab
bb-browser tab close 2
bb-browser tab close --id ab1c
```

Every response includes a short `tab` ID (e.g., `ab1c`) that you can use to target specific tabs:

```bash
# Run commands on a specific tab
bb-browser snapshot --tab ab1c
bb-browser click 3 --tab ab1c
bb-browser eval "document.title" --tab ab1c
```

### Frame (iframe) Navigation

```bash
# Switch to an iframe by CSS selector
bb-browser frame "#my-iframe"
bb-browser frame "iframe[name='content']"

# Switch back to the main frame
bb-browser frame main
```

### Dialog Handling

```bash
# Auto-accept future dialogs (alert, confirm, prompt)
bb-browser dialog accept

# Auto-dismiss future dialogs
bb-browser dialog dismiss

# Accept with prompt text
bb-browser dialog accept "my input"
```

### Authenticated Fetch

Make HTTP requests using the browser's cookies and session:

```bash
# GET request using browser's auth context
bb-browser fetch https://api.example.com/me

# POST request
bb-browser fetch https://api.example.com/data --method POST
```

This is useful for accessing authenticated APIs without extracting cookies manually.

### Trace (Record User Actions)

```bash
# Start recording
bb-browser trace start

# Check status
bb-browser trace status

# Stop and get recorded events
bb-browser trace stop
```

### Site Adapters

Site adapters are JavaScript plugins that automate interactions with specific websites.

```bash
# List available adapters
bb-browser site list

# Search for adapters
bb-browser site search twitter

# Get adapter details
bb-browser site info twitter/search

# Run an adapter
bb-browser site run twitter/search "AI news"
# or shorthand:
bb-browser twitter/search "AI news"

# Pull community adapters
bb-browser site update
```

### Daemon Management

```bash
# Start daemon in foreground (for debugging)
bb-browser daemon

# With custom CDP port
bb-browser daemon --cdp-port 9222

# Check daemon status
bb-browser daemon status
# or
bb-browser status

# Stop the daemon
bb-browser daemon shutdown
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--tab <id>` | Target a specific tab by short ID or index |
| `--json` | Output results as JSON |
| `--jq <expr>` | Apply a jq-like filter to the output |
| `--since <seq\|last_action>` | Only return events after a sequence number or the last action |

### JSON Output

Every command supports `--json` for machine-readable output:

```bash
bb-browser snapshot --json
bb-browser tab --json
bb-browser network requests --json
```

### jq Filtering

Built-in jq-compatible expression filtering (no external `jq` binary needed):

```bash
# Get just the snapshot text
bb-browser snapshot --jq ".data.snapshotData.snapshot"

# Count tabs
bb-browser tab --json --jq ".data.tabs | length"

# Get all request URLs
bb-browser network requests --jq ".data.networkRequests[].url"

# Filter network requests by status
bb-browser network requests --jq '.data.networkRequests[] | select(.status > 400) | {url: .url, status: .status}'
```

### Incremental Queries

Use `--since` to only get events that occurred after a specific point:

```bash
# Get events since a sequence number
bb-browser network requests --since 42

# Get events since the last user action
bb-browser console --since last_action
bb-browser errors --since last_action
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `BB_BROWSER_CDP_URL` | Override CDP endpoint (e.g., `http://127.0.0.1:9222`) |
| `BB_BROWSER_HOME` | Override config directory (default: `~/.bb-browser`) |

## Use Cases

### Web Scraping with Authentication

```bash
# Open the target site and log in manually (or use fill/click to automate)
bb-browser open https://app.example.com/login
bb-browser snapshot -i
bb-browser fill 0 "user@example.com"
bb-browser fill 1 "password123"
bb-browser click 2

# Now fetch authenticated API data
bb-browser fetch https://app.example.com/api/dashboard --json
```

### Automated Testing

```bash
# Navigate to the app
bb-browser open http://localhost:3000

# Fill in a form
bb-browser snapshot -i
bb-browser fill 0 "Test User"
bb-browser fill 1 "test@example.com"
bb-browser click 3

# Verify the result
bb-browser get text 5
bb-browser errors
```

### Monitoring & Debugging

```bash
# Watch network traffic
bb-browser open https://myapp.com
bb-browser network requests --filter "api" --json

# Check for JS errors
bb-browser errors

# Read console output
bb-browser console --filter "warning"
```

### AI Agent Integration

`bb-browser` is designed to work well with AI agents (like Claude, GPT, etc.) that need to interact with web pages. The snapshot command produces an accessibility tree that LLMs can understand:

```bash
# The agent runs snapshot to "see" the page
bb-browser snapshot -i -c

# The agent decides which element to interact with based on the ref numbers
bb-browser click 4
bb-browser fill 7 "search query"

# The agent checks the result
bb-browser snapshot -i -c
```

## Typical Workflow

```bash
# 1. Open a page
bb-browser open https://news.ycombinator.com

# 2. See what's on the page
bb-browser snapshot -i

# 3. Interact with elements using ref numbers
bb-browser click 5

# 4. See the result
bb-browser snapshot -i

# 5. Extract data
bb-browser eval "document.querySelector('.title').textContent"

# 6. Get structured JSON output
bb-browser snapshot --json --jq ".data.snapshotData.refs"
```

## License

MIT
