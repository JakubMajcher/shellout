// SPDX-License-Identifier: GPL-3.0-or-later

package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Suggestion is one validated answer from the model.
type Suggestion struct {
	Description string
	Command     string
}

// Client talks to one OpenAI-compatible endpoint.
type Client struct {
	BaseURL string
	Model   string
	APIKey  string // empty: no Authorization header
	Profile string // used only in error messages
	Params  map[string]any
	Timeout time.Duration
	HTTP    *http.Client // nil: http.DefaultClient
}

const maxAttempts = 2

var errIncomplete = errors.New("model did not return a complete command")

// retryable marks an attempt failure that is worth one more try.
type retryable struct{ err error }

func (r *retryable) Error() string { return r.err.Error() }
func (r *retryable) Unwrap() error { return r.err }

var responseFormat = map[string]any{
	"type": "json_schema",
	"json_schema": map[string]any{
		"name":   "shell_command",
		"strict": true,
		"schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"description": map[string]any{"type": "string"},
				"command":     map[string]any{"type": "string"},
			},
			"required":             []string{"description", "command"},
			"additionalProperties": false,
		},
	},
}

// Suggest asks the model for one command. It makes at most two HTTP calls.
func (c *Client) Suggest(ctx context.Context, system, request string) (Suggestion, error) {
	body := make(map[string]any, len(c.Params)+3)
	for k, v := range c.Params {
		body[k] = v
	}
	body["model"] = c.Model
	body["messages"] = []map[string]string{
		{"role": "system", "content": system},
		{"role": "user", "content": request},
	}
	body["response_format"] = responseFormat
	payload, err := json.Marshal(body)
	if err != nil {
		return Suggestion{}, fmt.Errorf("cannot encode request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		s, err := c.attempt(ctx, payload)
		if err == nil {
			return s, nil
		}
		var r *retryable
		if !errors.As(err, &r) {
			return Suggestion{}, err
		}
		lastErr = r.err
	}
	return Suggestion{}, lastErr
}

func (c *Client) attempt(ctx context.Context, payload []byte) (Suggestion, error) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	url := strings.TrimRight(c.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return Suggestion{}, fmt.Errorf("request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	hc := c.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(req)
	if err != nil {
		return Suggestion{}, &retryable{fmt.Errorf("request failed: %w", err)}
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Suggestion{}, &retryable{fmt.Errorf("request failed: %w", err)}
	}

	if resp.StatusCode != http.StatusOK {
		msg := resp.Status + ": " + apiErrorMessage(raw)
		if resp.StatusCode == http.StatusBadRequest &&
			(bytes.Contains(raw, []byte("response_format")) || bytes.Contains(raw, []byte("json_schema"))) {
			msg += "\nthis model or endpoint does not support structured outputs"
		}
		if resp.StatusCode >= 500 {
			return Suggestion{}, &retryable{errors.New(msg)}
		}
		return Suggestion{}, errors.New(msg)
	}

	var cr struct {
		Choices []struct {
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Content *string `json:"content"`
				Refusal *string `json:"refusal"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &cr); err != nil || len(cr.Choices) == 0 {
		return Suggestion{}, &retryable{errIncomplete}
	}
	ch := cr.Choices[0]
	if ch.Message.Refusal != nil && *ch.Message.Refusal != "" {
		return Suggestion{}, fmt.Errorf("model refused: %s", *ch.Message.Refusal)
	}
	if ch.FinishReason == "length" {
		return Suggestion{}, fmt.Errorf("model ran out of tokens; raise the token limit in params (profile: %s)", c.Profile)
	}
	if ch.FinishReason != "stop" || ch.Message.Content == nil {
		return Suggestion{}, &retryable{errIncomplete}
	}

	var out struct {
		Description *string `json:"description"`
		Command     *string `json:"command"`
	}
	if err := json.Unmarshal([]byte(*ch.Message.Content), &out); err != nil ||
		out.Description == nil || out.Command == nil {
		return Suggestion{}, &retryable{errIncomplete}
	}
	cmd := strings.TrimSpace(*out.Command)
	if cmd == "" {
		return Suggestion{}, &retryable{errIncomplete}
	}
	return Suggestion{Description: strings.TrimSpace(*out.Description), Command: cmd}, nil
}

// apiErrorMessage extracts a readable message from an error response body.
func apiErrorMessage(raw []byte) string {
	var body struct {
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(raw, &body) == nil && len(body.Error) > 0 {
		var asString string
		if json.Unmarshal(body.Error, &asString) == nil && asString != "" {
			return asString
		}
		var asObject struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body.Error, &asObject) == nil && asObject.Message != "" {
			return asObject.Message
		}
	}
	s := strings.TrimSpace(string(raw))
	if len(s) > 300 {
		s = s[:300]
	}
	return s
}
