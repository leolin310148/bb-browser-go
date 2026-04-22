package mcp

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/leolin310148/bb-browser-go/internal/protocol"
	"github.com/mark3labs/mcp-go/mcp"
)

// stubSend swaps sendCommand for the duration of a test.
func stubSend(t *testing.T, fn func(*protocol.Request) (*protocol.Response, error)) {
	t.Helper()
	orig := sendCommand
	sendCommand = fn
	t.Cleanup(func() { sendCommand = orig })
}

// capturingSend records the request and returns a preset response.
type capture struct {
	req  *protocol.Request
	resp *protocol.Response
	err  error
}

func capturingSend(t *testing.T, resp *protocol.Response) *capture {
	t.Helper()
	c := &capture{resp: resp}
	stubSend(t, func(r *protocol.Request) (*protocol.Response, error) {
		c.req = r
		return c.resp, c.err
	})
	return c
}

func mkReq(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}}
}

func ok() *protocol.Response { return &protocol.Response{Success: true} }

// --- helpers ---

func TestNormalizeRef(t *testing.T) {
	if normalizeRef("@5") != "5" {
		t.Error("strip @ prefix")
	}
	if normalizeRef("5") != "5" {
		t.Error("no-op without @")
	}
}

func TestIntPtr(t *testing.T) {
	if p := intPtr(7); p == nil || *p != 7 {
		t.Errorf("intPtr(7) = %v", p)
	}
}

func TestNewID_HexLength(t *testing.T) {
	a := newID()
	if len(a) != 16 {
		t.Errorf("len = %d, want 16 hex chars", len(a))
	}
	// likely unique across consecutive calls
	if a == newID() {
		t.Error("two consecutive newIDs are equal — RNG broken?")
	}
}

func TestSetTab(t *testing.T) {
	req := &protocol.Request{}
	setTab(req, mkReq(map[string]any{"tab": "t1"}))
	if req.TabID != "t1" {
		t.Errorf("TabID = %v", req.TabID)
	}

	req = &protocol.Request{}
	setTab(req, mkReq(nil))
	if req.TabID != nil {
		t.Errorf("TabID should stay nil, got %v", req.TabID)
	}
}

// --- navigation handlers ---

func TestHandleNavigate_MissingURL(t *testing.T) {
	res, _ := handleNavigate(context.Background(), mkReq(nil))
	if !res.IsError {
		t.Fatalf("expected error result")
	}
}

func TestHandleNavigate_Success(t *testing.T) {
	cap := capturingSend(t, ok())
	res, _ := handleNavigate(context.Background(), mkReq(map[string]any{"url": "https://example.com", "new": true, "tab": "t1"}))
	if res.IsError {
		t.Fatalf("unexpected error: %v", res)
	}
	if cap.req.Action != protocol.ActionOpen || cap.req.URL != "https://example.com" || !cap.req.New || cap.req.TabID != "t1" {
		t.Errorf("req = %+v", cap.req)
	}
}

func TestHandleNavigate_SendError(t *testing.T) {
	stubSend(t, func(*protocol.Request) (*protocol.Response, error) {
		return nil, errors.New("down")
	})
	res, _ := handleNavigate(context.Background(), mkReq(map[string]any{"url": "x"}))
	if !res.IsError {
		t.Errorf("expected error, got %v", res)
	}
}

func TestHandleNavigate_CommandFailure(t *testing.T) {
	capturingSend(t, &protocol.Response{Success: false, Error: "boom"})
	res, _ := handleNavigate(context.Background(), mkReq(map[string]any{"url": "x"}))
	if !res.IsError {
		t.Error("expected error result")
	}
}

func TestHandleBackForwardRefreshClose(t *testing.T) {
	cases := []struct {
		name   string
		fn     func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
		action protocol.ActionType
	}{
		{"back", handleBack, protocol.ActionBack},
		{"forward", handleForward, protocol.ActionForward},
		{"refresh", handleRefresh, protocol.ActionRefresh},
		{"close", handleClose, protocol.ActionClose},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cap := capturingSend(t, ok())
			res, _ := c.fn(context.Background(), mkReq(map[string]any{"tab": "T"}))
			if res.IsError {
				t.Fatalf("unexpected error: %v", res)
			}
			if cap.req.Action != c.action {
				t.Errorf("action = %v, want %v", cap.req.Action, c.action)
			}
			if cap.req.TabID != "T" {
				t.Errorf("tab = %v", cap.req.TabID)
			}
		})
	}
}

// --- interaction handlers ---

func TestRefRequiredHandlers(t *testing.T) {
	handlers := map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error){
		"click":   handleClick,
		"hover":   handleHover,
		"check":   handleCheck,
		"uncheck": handleUncheck,
	}
	for name, h := range handlers {
		t.Run(name, func(t *testing.T) {
			res, _ := h(context.Background(), mkReq(nil))
			if !res.IsError {
				t.Errorf("%s without ref should error", name)
			}
		})
	}
}

func TestHandleClick_NormalizesRefAndSends(t *testing.T) {
	cap := capturingSend(t, ok())
	_, _ = handleClick(context.Background(), mkReq(map[string]any{"ref": "@7"}))
	if cap.req.Ref != "7" || cap.req.Action != protocol.ActionClick {
		t.Errorf("req = %+v", cap.req)
	}
}

func TestSimpleRefHandlers_Success(t *testing.T) {
	cases := []struct {
		name   string
		fn     func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
		action protocol.ActionType
	}{
		{"hover", handleHover, protocol.ActionHover},
		{"check", handleCheck, protocol.ActionCheck},
		{"uncheck", handleUncheck, protocol.ActionUncheck},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cap := capturingSend(t, ok())
			res, _ := c.fn(context.Background(), mkReq(map[string]any{"ref": "@4", "tab": "t"}))
			if res.IsError {
				t.Fatalf("unexpected error: %v", res)
			}
			if cap.req.Action != c.action || cap.req.Ref != "4" || cap.req.TabID != "t" {
				t.Errorf("req = %+v", cap.req)
			}
		})
	}
}

func TestHandleFill(t *testing.T) {
	// missing ref
	res, _ := handleFill(context.Background(), mkReq(nil))
	if !res.IsError {
		t.Error("missing ref should error")
	}
	// missing text
	res, _ = handleFill(context.Background(), mkReq(map[string]any{"ref": "1"}))
	if !res.IsError {
		t.Error("missing text should error")
	}
	// success
	cap := capturingSend(t, ok())
	_, _ = handleFill(context.Background(), mkReq(map[string]any{"ref": "@9", "text": "hi"}))
	if cap.req.Ref != "9" || cap.req.Text != "hi" || cap.req.Action != protocol.ActionFill {
		t.Errorf("req = %+v", cap.req)
	}
}

func TestHandleType(t *testing.T) {
	res, _ := handleType(context.Background(), mkReq(map[string]any{"ref": "1"}))
	if !res.IsError {
		t.Error("missing text should error")
	}
	cap := capturingSend(t, ok())
	_, _ = handleType(context.Background(), mkReq(map[string]any{"ref": "2", "text": "x"}))
	if cap.req.Action != protocol.ActionType_ || cap.req.Text != "x" {
		t.Errorf("req = %+v", cap.req)
	}
}

func TestHandleSelect(t *testing.T) {
	res, _ := handleSelect(context.Background(), mkReq(nil))
	if !res.IsError {
		t.Error("missing ref")
	}
	res, _ = handleSelect(context.Background(), mkReq(map[string]any{"ref": "1"}))
	if !res.IsError {
		t.Error("missing value")
	}
	cap := capturingSend(t, ok())
	_, _ = handleSelect(context.Background(), mkReq(map[string]any{"ref": "1", "value": "opt"}))
	if cap.req.Value != "opt" {
		t.Errorf("value = %q", cap.req.Value)
	}
}

func TestHandlePress(t *testing.T) {
	res, _ := handlePress(context.Background(), mkReq(nil))
	if !res.IsError {
		t.Error("missing key")
	}
	cap := capturingSend(t, ok())
	_, _ = handlePress(context.Background(), mkReq(map[string]any{"key": "Enter"}))
	if cap.req.Key != "Enter" {
		t.Errorf("key = %q", cap.req.Key)
	}
}

func TestHandleScroll_Defaults(t *testing.T) {
	cap := capturingSend(t, ok())
	_, _ = handleScroll(context.Background(), mkReq(nil))
	if cap.req.Direction != "down" {
		t.Errorf("default direction = %q", cap.req.Direction)
	}
	if cap.req.Pixels == nil || *cap.req.Pixels != 300 {
		t.Errorf("default pixels = %v", cap.req.Pixels)
	}
}

func TestHandleScroll_Custom(t *testing.T) {
	cap := capturingSend(t, ok())
	_, _ = handleScroll(context.Background(), mkReq(map[string]any{"direction": "up", "pixels": float64(100)}))
	if cap.req.Direction != "up" || *cap.req.Pixels != 100 {
		t.Errorf("req = %+v", cap.req)
	}
}

// --- observation handlers ---

func TestHandleSnapshot(t *testing.T) {
	cap := capturingSend(t, &protocol.Response{Success: true, Data: &protocol.ResponseData{
		SnapshotData: &protocol.SnapshotData{Snapshot: "tree"},
	}})
	_, _ = handleSnapshot(context.Background(), mkReq(map[string]any{
		"interactive": true, "compact": true, "maxDepth": float64(3), "selector": "body",
	}))
	if !cap.req.Interactive || !cap.req.Compact || cap.req.Selector != "body" {
		t.Errorf("req = %+v", cap.req)
	}
	if cap.req.MaxDepth == nil || *cap.req.MaxDepth != 3 {
		t.Errorf("maxDepth = %v", cap.req.MaxDepth)
	}
}

func TestHandleSnapshot_ZeroDepthOmitted(t *testing.T) {
	cap := capturingSend(t, ok())
	_, _ = handleSnapshot(context.Background(), mkReq(nil))
	if cap.req.MaxDepth != nil {
		t.Errorf("zero depth should be omitted: %v", cap.req.MaxDepth)
	}
}

func TestHandleScreenshot(t *testing.T) {
	capturingSend(t, &protocol.Response{Success: true, Data: &protocol.ResponseData{DataURL: "data:image/png;base64,AAA"}})
	res, _ := handleScreenshot(context.Background(), mkReq(nil))
	if res.IsError {
		t.Errorf("unexpected err: %v", res)
	}
}

func TestHandleGet(t *testing.T) {
	res, _ := handleGet(context.Background(), mkReq(nil))
	if !res.IsError {
		t.Error("missing attribute")
	}
	cap := capturingSend(t, &protocol.Response{Success: true, Data: &protocol.ResponseData{Value: "ok"}})
	_, _ = handleGet(context.Background(), mkReq(map[string]any{"attribute": "text", "ref": "@3"}))
	if cap.req.Attribute != "text" || cap.req.Ref != "3" {
		t.Errorf("req = %+v", cap.req)
	}
}

func TestHandleEval(t *testing.T) {
	res, _ := handleEval(context.Background(), mkReq(nil))
	if !res.IsError {
		t.Error("missing script")
	}
	cap := capturingSend(t, &protocol.Response{Success: true, Data: &protocol.ResponseData{Result: "hi"}})
	_, _ = handleEval(context.Background(), mkReq(map[string]any{"script": "1+1"}))
	if cap.req.Script != "1+1" {
		t.Errorf("script = %q", cap.req.Script)
	}
}

func TestHandleWait(t *testing.T) {
	cap := capturingSend(t, ok())
	_, _ = handleWait(context.Background(), mkReq(map[string]any{"ms": float64(250)}))
	if cap.req.Ms == nil || *cap.req.Ms != 250 {
		t.Errorf("ms = %v", cap.req.Ms)
	}

	cap = capturingSend(t, ok())
	_, _ = handleWait(context.Background(), mkReq(nil))
	if cap.req.Ms == nil || *cap.req.Ms != 1000 {
		t.Errorf("default ms = %v", cap.req.Ms)
	}
}

// --- tab handlers ---

func TestHandleTabList(t *testing.T) {
	cap := capturingSend(t, &protocol.Response{Success: true, Data: &protocol.ResponseData{Tabs: []protocol.TabInfo{{Index: 0, URL: "u"}}}})
	res, _ := handleTabList(context.Background(), mkReq(nil))
	if res.IsError {
		t.Errorf("unexpected error")
	}
	if cap.req.Action != protocol.ActionTabList {
		t.Errorf("action = %v", cap.req.Action)
	}
}

func TestHandleTabNew(t *testing.T) {
	cap := capturingSend(t, ok())
	_, _ = handleTabNew(context.Background(), mkReq(map[string]any{"url": "https://x"}))
	if cap.req.URL != "https://x" {
		t.Errorf("url = %q", cap.req.URL)
	}

	cap = capturingSend(t, ok())
	res, _ := handleTabNew(context.Background(), mkReq(nil))
	if cap.req.URL != "" {
		t.Errorf("URL should be empty")
	}
	if res.IsError {
		t.Errorf("unexpected error")
	}
}

func TestHandleTabSelect(t *testing.T) {
	cap := capturingSend(t, ok())
	_, _ = handleTabSelect(context.Background(), mkReq(map[string]any{"tab": "t1", "index": float64(2)}))
	if cap.req.TabID != "t1" || cap.req.Index == nil || *cap.req.Index != 2 {
		t.Errorf("req = %+v", cap.req)
	}

	cap = capturingSend(t, ok())
	_, _ = handleTabSelect(context.Background(), mkReq(nil))
	if cap.req.TabID != nil || cap.req.Index != nil {
		t.Errorf("empty tab-select should have no fields: %+v", cap.req)
	}
}

func TestHandleTabClose(t *testing.T) {
	cap := capturingSend(t, ok())
	_, _ = handleTabClose(context.Background(), mkReq(map[string]any{"tab": "t2", "index": float64(1)}))
	if cap.req.TabID != "t2" || cap.req.Index == nil || *cap.req.Index != 1 {
		t.Errorf("req = %+v", cap.req)
	}
}

// --- diagnostics handlers ---

func TestHandleNetwork_List(t *testing.T) {
	cap := capturingSend(t, &protocol.Response{Success: true, Data: &protocol.ResponseData{
		NetworkRequests: []protocol.NetworkRequestInfo{{URL: "u", Method: "GET", Type: "xhr"}},
	}})
	res, _ := handleNetwork(context.Background(), mkReq(map[string]any{
		"command": "requests", "filter": "f", "withBody": true, "method": "POST", "status": "200",
	}))
	if res.IsError {
		t.Error("unexpected error")
	}
	if cap.req.NetworkCommand != "requests" || !cap.req.WithBody || cap.req.Method != "POST" {
		t.Errorf("req = %+v", cap.req)
	}
}

func TestHandleNetwork_Clear(t *testing.T) {
	capturingSend(t, ok())
	res, _ := handleNetwork(context.Background(), mkReq(map[string]any{"command": "clear"}))
	if !strings.Contains(firstText(t, res), "cleared") {
		t.Errorf("got %q", firstText(t, res))
	}
}

func TestHandleConsole(t *testing.T) {
	// get mode (default)
	cap := capturingSend(t, &protocol.Response{Success: true, Data: &protocol.ResponseData{
		ConsoleMessages: []protocol.ConsoleMessageInfo{{Type: "log", Text: "x"}},
	}})
	res, _ := handleConsole(context.Background(), mkReq(map[string]any{"filter": "f"}))
	if res.IsError || cap.req.ConsoleCommand != "get" {
		t.Errorf("get-mode failed: %v / %+v", res, cap.req)
	}

	// clear mode
	capturingSend(t, ok())
	res, _ = handleConsole(context.Background(), mkReq(map[string]any{"clear": true}))
	if !strings.Contains(firstText(t, res), "cleared") {
		t.Errorf("got %q", firstText(t, res))
	}
}

func TestHandleErrors(t *testing.T) {
	cap := capturingSend(t, &protocol.Response{Success: true, Data: &protocol.ResponseData{
		JSErrors: []protocol.JSErrorInfo{{Message: "oops"}},
	}})
	res, _ := handleErrors(context.Background(), mkReq(nil))
	if res.IsError || cap.req.ErrorsCommand != "get" {
		t.Errorf("get-mode failed: %v", res)
	}

	capturingSend(t, ok())
	res, _ = handleErrors(context.Background(), mkReq(map[string]any{"clear": true}))
	if !strings.Contains(firstText(t, res), "cleared") {
		t.Errorf("got %q", firstText(t, res))
	}
}
