package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Minimal MCP server over stdio (JSON-RPC 2.0).
// Spec: https://modelcontextprotocol.io/specification/latest/

type requestID any

type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      requestID       `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      requestID     `json:"id,omitempty"`
	Result  any           `json:"result,omitempty"`
	Error   *jsonrpcError `json:"error,omitempty"`
}

type implementation struct {
	Name    string `json:"name"`
	Title   string `json:"title,omitempty"`
	Version string `json:"version"`
}

type initializeParams struct {
	ProtocolVersion string `json:"protocolVersion"`
	// capabilities, clientInfo exist but we don't need them for minimal operation.
}

type initializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ServerInfo      implementation `json:"serverInfo"`
	Instructions    string         `json:"instructions,omitempty"`
}

type tool struct {
	Name        string         `json:"name"`
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema"`
}

type toolsListResult struct {
	Tools []tool `json:"tools"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type toolResultContentText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolCallResult struct {
	Content []toolResultContentText `json:"content"`
	IsError bool                    `json:"isError,omitempty"`
}

func main() {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)

	dec := json.NewDecoder(os.Stdin)

	var initialized bool
	var protocolVersion string

	for {
		var req jsonrpcRequest
		if err := dec.Decode(&req); err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			// Can't respond reliably without an id; best-effort stderr.
			fmt.Fprintln(os.Stderr, "decode error:", err)
			return
		}

		// Notifications have no id; ignore unless we need them.
		if req.ID == nil {
			// Accept the client's "notifications/initialized".
			continue
		}

		switch req.Method {
		case "initialize":
			var p initializeParams
			_ = json.Unmarshal(req.Params, &p)
			protocolVersion = p.ProtocolVersion
			if protocolVersion == "" {
				// Fallback to a known good version string; clients usually accept.
				protocolVersion = "2025-11-25"
			}

			res := initializeResult{
				ProtocolVersion: protocolVersion,
				Capabilities: map[string]any{
					"tools": map[string]any{
						"listChanged": false,
					},
				},
				ServerInfo: implementation{
					Name:    "kontarawa-mcp",
					Title:   "Kontarawa MCP",
					Version: "0.1.0",
				},
				Instructions: strings.TrimSpace(`
Use these tools to interact with the local kontarawa agent.

- kontarawa_ask: ask a question (uses kontarawa memory + retrieval)
- kontarawa_doctor: quick health check (ollama connectivity, memory dir)
- kontarawa_learn: store a lesson (bad/good/why) to improve future answers
`),
			}
			initialized = true
			writeResult(enc, req.ID, res)

		case "ping":
			writeResult(enc, req.ID, map[string]any{})

		case "tools/list":
			if !initialized {
				writeError(enc, req.ID, -32002, "Server not initialized", nil)
				continue
			}
			writeResult(enc, req.ID, toolsListResult{Tools: []tool{
				{
					Name:        "kontarawa_ask",
					Title:       "Kontarawa Ask",
					Description: "Ask the kontarawa agent a question.",
					InputSchema: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"prompt": map[string]any{
								"type":        "string",
								"description": "User question/instruction.",
							},
							"kontarawa_path": map[string]any{
								"type":        "string",
								"description": "Optional path to kontarawa binary. Defaults to ./kontarawa in the current working directory.",
							},
						},
						"required": []string{"prompt"},
					},
				},
				{
					Name:        "kontarawa_doctor",
					Title:       "Kontarawa Doctor",
					Description: "Run kontarawa doctor to check setup.",
					InputSchema: map[string]any{
						"type":       "object",
						"properties": map[string]any{},
					},
				},
				{
					Name:        "kontarawa_learn",
					Title:       "Kontarawa Learn",
					Description: "Save a lesson (bad/good/why) into kontarawa memory.",
					InputSchema: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"prompt": map[string]any{
								"type":        "string",
								"description": "Original prompt/question.",
							},
							"bad": map[string]any{
								"type":        "string",
								"description": "Bad answer (what went wrong).",
							},
							"good": map[string]any{
								"type":        "string",
								"description": "Good answer (desired output).",
							},
							"why": map[string]any{
								"type":        "string",
								"description": "Explanation of why the good answer is better.",
							},
							"kontarawa_path": map[string]any{
								"type":        "string",
								"description": "Optional path to kontarawa binary. Defaults to ./kontarawa in the current working directory.",
							},
						},
						"required": []string{"prompt", "bad", "good", "why"},
					},
				},
			}})

		case "tools/call":
			if !initialized {
				writeError(enc, req.ID, -32002, "Server not initialized", nil)
				continue
			}
			var p toolCallParams
			if err := json.Unmarshal(req.Params, &p); err != nil {
				writeError(enc, req.ID, -32602, "Invalid params", err.Error())
				continue
			}

			out, isErr, err := handleToolCall(p)
			if err != nil {
				writeError(enc, req.ID, -32000, "Tool execution failed", err.Error())
				continue
			}
			writeResult(enc, req.ID, toolCallResult{
				Content: []toolResultContentText{{Type: "text", Text: out}},
				IsError: isErr,
			})

		default:
			writeError(enc, req.ID, -32601, "Method not found", req.Method)
		}
	}
}

func handleToolCall(p toolCallParams) (string, bool, error) {
	switch p.Name {
	case "kontarawa_ask":
		prompt, _ := p.Arguments["prompt"].(string)
		if strings.TrimSpace(prompt) == "" {
			return "missing required argument: prompt", true, nil
		}
		kPath := kontarawaPathFromArgs(p.Arguments)
		return runKontarawa([]string{"ask", prompt}, kPath)

	case "kontarawa_doctor":
		kPath := kontarawaPathFromArgs(p.Arguments)
		return runKontarawa([]string{"doctor"}, kPath)

	case "kontarawa_learn":
		get := func(k string) string {
			v, _ := p.Arguments[k].(string)
			return v
		}
		prompt := get("prompt")
		bad := get("bad")
		good := get("good")
		why := get("why")
		if strings.TrimSpace(prompt) == "" || strings.TrimSpace(bad) == "" || strings.TrimSpace(good) == "" || strings.TrimSpace(why) == "" {
			return "missing required arguments: prompt, bad, good, why", true, nil
		}
		kPath := kontarawaPathFromArgs(p.Arguments)
		args := []string{
			"learn",
			"--prompt", prompt,
			"--bad", bad,
			"--good", good,
			"--why", why,
		}
		return runKontarawa(args, kPath)
	default:
		return fmt.Sprintf("unknown tool: %s", p.Name), true, nil
	}
}

func kontarawaPathFromArgs(args map[string]any) string {
	if args == nil {
		return ""
	}
	if v, ok := args["kontarawa_path"]; ok {
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func runKontarawa(args []string, kontarawaPath string) (string, bool, error) {
	kPath := kontarawaPath
	if kPath == "" {
		kPath = "./kontarawa"
	}
	if !filepath.IsAbs(kPath) {
		abs, err := filepath.Abs(kPath)
		if err == nil {
			kPath = abs
		}
	}

	// Keep calls bounded; clients can retry.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, kPath, args...)
	cmd.Env = os.Environ()

	// Capture both streams (kontarawa may log to stderr).
	outBytes, err := cmd.CombinedOutput()
	out := strings.TrimSpace(string(outBytes))
	if err == nil {
		return out, false, nil
	}

	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if out == "" {
			out = ee.Error()
		}
		// Non-zero exit: treat as tool error but still return text to the model.
		return out, true, nil
	}
	return out, true, err
}

func writeResult(enc *json.Encoder, id requestID, result any) {
	_ = enc.Encode(jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func writeError(enc *json.Encoder, id requestID, code int, message string, data any) {
	_ = enc.Encode(jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &jsonrpcError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	})
}
