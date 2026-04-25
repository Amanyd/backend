package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/eulerbutcooler/wingman-backend/internal/port"
)

type ragClient struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewRAGClient(baseURL, token string) port.RagClient {
	return &ragClient{
		baseURL: baseURL,
		token:   token,
		http: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *ragClient) ChatStream(ctx context.Context, req port.ChatRequest) (io.ReadCloser, error) {
	req.Stream = true
	return c.doRequest(ctx, req)
}

func (c *ragClient) Chat(ctx context.Context, req port.ChatRequest) (*port.ChatResponse, error) {
	req.Stream = false
	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var resp port.ChatResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("rag decode response: %w", err)
	}
	return &resp, nil
}

func (c *ragClient) doRequest(ctx context.Context, req port.ChatRequest) (io.ReadCloser, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("rag marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/api/v1/chat",
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("rag new request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Internal-Token", c.token)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("rag http do: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("rag unexpected status: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
