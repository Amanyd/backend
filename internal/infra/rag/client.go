package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Amanyd/backend/internal/port"
)

type ragClient struct {
	baseURL    string
	token      string
	http       *http.Client // for non-streaming (short timeout)
	httpStream *http.Client // for SSE streaming (no timeout; context drives cancellation)
}

func NewRAGClient(baseURL, token string) port.RagClient {
	return &ragClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http: &http.Client{
			Timeout: 60 * time.Second,
		},
		// No Timeout on the streaming client: http.Client.Timeout covers the
		// entire response including body reads, which would cut off long LLM
		// generations. Cancellation is handled via the request context instead.
		httpStream: &http.Client{},
	}
}

func (c *ragClient) ChatStream(ctx context.Context, req port.ChatRequest) (io.ReadCloser, error) {
	req.Stream = true
	return c.doRequest(ctx, req, true)
}

func (c *ragClient) Chat(ctx context.Context, req port.ChatRequest) (*port.ChatResponse, error) {
	req.Stream = false
	body, err := c.doRequest(ctx, req, false)
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

func (c *ragClient) GradeAnswer(ctx context.Context, req port.GradeRequest) (*port.GradeResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("rag marshal grade request: %w", err)
	}

	endpoint := strings.TrimRight(c.baseURL, "/") + "/api/v1/quiz/grade"
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("rag new grade request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Internal-Token", c.token)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("rag grade http do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rag grade unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var gradeResp port.GradeResponse
	if err := json.NewDecoder(resp.Body).Decode(&gradeResp); err != nil {
		return nil, fmt.Errorf("rag decode grade response: %w", err)
	}
	return &gradeResp, nil
}

func (c *ragClient) DeleteCourse(ctx context.Context, courseID string) error {
	endpoint := strings.TrimRight(c.baseURL, "/") + "/api/v1/course/" + courseID
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		endpoint,
		nil,
	)
	if err != nil {
		return fmt.Errorf("rag new delete request: %w", err)
	}

	httpReq.Header.Set("X-Internal-Token", c.token)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("rag delete http do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("rag delete unexpected status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (c *ragClient) doRequest(ctx context.Context, req port.ChatRequest, stream bool) (io.ReadCloser, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("rag marshal request: %w", err)
	}

	endpoint := strings.TrimRight(c.baseURL, "/") + "/api/v1/chat/"
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("rag new request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Internal-Token", c.token)

	client := c.http
	if stream {
		client = c.httpStream
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("rag http do: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("rag unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	return resp.Body, nil
}
