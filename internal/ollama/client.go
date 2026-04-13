package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	Host   string
	Client *http.Client
}

func New(host string) *Client {
	host = strings.TrimSpace(host)
	if host == "" {
		host = "http://localhost:11434"
	}
	return &Client{
		Host: host,
		Client: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

func (c *Client) Ping() bool {
	hc := c.Client
	if hc == nil {
		hc = &http.Client{Timeout: 2 * time.Second}
	}
	req, _ := http.NewRequest(http.MethodGet, strings.TrimRight(c.Host, "/")+"/api/tags", nil)
	resp, err := hc.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatReq struct {
	Model    string    `json:"model"`
	Stream   bool      `json:"stream"`
	Messages []Message `json:"messages"`
}

type chatResp struct {
	Message Message `json:"message"`
}

func (c *Client) Chat(model, system, user string) (string, error) {
	url := strings.TrimRight(c.Host, "/") + "/api/chat"
	reqBody := chatReq{
		Model:  model,
		Stream: false,
		Messages: []Message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	}
	b, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	hc := c.Client
	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Minute}
	}
	resp, err := hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ollama error %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var out chatResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("ollama bad json: %w", err)
	}
	return strings.TrimSpace(out.Message.Content), nil
}

