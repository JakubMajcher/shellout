// SPDX-License-Identifier: GPL-3.0-or-later

package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// completion returns an OpenAI-style chat completion body.
func completion(finish, content string) string {
	b, _ := json.Marshal(map[string]any{
		"choices": []any{map[string]any{
			"finish_reason": finish,
			"message":       map[string]any{"role": "assistant", "content": content},
		}},
	})
	return string(b)
}

const goodContent = `{"description":"Lists big files.","command":"find . -size +1G"}`

// server replies with the given responses in order (the last one repeats)
// and counts the calls.
func server(t *testing.T, responses ...func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, *int32) {
	t.Helper()
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(atomic.AddInt32(&calls, 1)) - 1
		if n >= len(responses) {
			n = len(responses) - 1
		}
		responses[n](w, r)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func reply(status int, body string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		io.WriteString(w, body)
	}
}

func client(url string) *Client {
	return &Client{BaseURL: url + "/", Model: "m", Profile: "test", Timeout: 2 * time.Second}
}

func TestSuggestValidResponse(t *testing.T) {
	srv, calls := server(t, reply(200, completion("stop", goodContent)))
	s, err := client(srv.URL).Suggest(context.Background(), "sys", "find big files")
	if err != nil {
		t.Fatal(err)
	}
	if s.Command != "find . -size +1G" || s.Description != "Lists big files." || *calls != 1 {
		t.Fatalf("got %+v after %d calls", s, *calls)
	}
}

func TestRequestBodyIsExact(t *testing.T) {
	var body map[string]any
	var auth string
	var path string
	srv, _ := server(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		auth = r.Header.Get("Authorization")
		json.NewDecoder(r.Body).Decode(&body)
		reply(200, completion("stop", goodContent))(w, r)
	})
	c := client(srv.URL)
	c.APIKey = "sk-1"
	c.Params = map[string]any{"reasoning_effort": "minimal", "max_completion_tokens": int64(1000),
		"provider": map[string]any{"require_parameters": true}}
	if _, err := c.Suggest(context.Background(), "sys", "req"); err != nil {
		t.Fatal(err)
	}
	if path != "/chat/completions" {
		t.Errorf("path = %s", path)
	}
	if auth != "Bearer sk-1" {
		t.Errorf("auth = %q", auth)
	}
	want := map[string]bool{"model": true, "messages": true, "response_format": true,
		"reasoning_effort": true, "max_completion_tokens": true, "provider": true}
	for k := range body {
		if !want[k] {
			t.Errorf("unexpected key %q in request body", k)
		}
	}
	if len(body) != len(want) {
		t.Errorf("body keys = %v", body)
	}
	if body["model"] != "m" || body["reasoning_effort"] != "minimal" || body["max_completion_tokens"] != float64(1000) {
		t.Errorf("body = %v", body)
	}
	rf := body["response_format"].(map[string]any)
	js := rf["json_schema"].(map[string]any)
	schema := js["schema"].(map[string]any)
	if rf["type"] != "json_schema" || js["name"] != "shell_command" || js["strict"] != true ||
		schema["additionalProperties"] != false {
		t.Errorf("response_format = %v", rf)
	}
	msgs := body["messages"].([]any)
	if len(msgs) != 2 || msgs[0].(map[string]any)["role"] != "system" || msgs[1].(map[string]any)["content"] != "req" {
		t.Errorf("messages = %v", msgs)
	}
}

func TestNoAuthorizationWithoutKey(t *testing.T) {
	var auth string
	srv, _ := server(t, func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		reply(200, completion("stop", goodContent))(w, r)
	})
	if _, err := client(srv.URL).Suggest(context.Background(), "s", "r"); err != nil {
		t.Fatal(err)
	}
	if auth != "" {
		t.Fatalf("Authorization sent: %q", auth)
	}
}

func TestRetriesOnceThenFails(t *testing.T) {
	cases := map[string]func(http.ResponseWriter, *http.Request){
		"malformed json": reply(200, completion("stop", `{"command": "ls"`)),
		"missing field":  reply(200, completion("stop", `{"command":"ls"}`)),
		"empty command":  reply(200, completion("stop", `{"description":"x","command":"   "}`)),
		"no choices":     reply(200, `{"choices":[]}`),
		"server error":   reply(500, `{"error":{"message":"boom"}}`),
	}
	for name, h := range cases {
		srv, calls := server(t, h)
		_, err := client(srv.URL).Suggest(context.Background(), "s", "r")
		if err == nil || *calls != 2 {
			t.Errorf("%s: err=%v calls=%d, want error after 2 calls", name, err, *calls)
		}
		if name == "server error" {
			if err.Error() != "500 Internal Server Error: boom" {
				t.Errorf("%s: message %q", name, err)
			}
		} else if err.Error() != "model did not return a complete command" {
			t.Errorf("%s: message %q", name, err)
		}
	}
}

func TestRetrySucceedsOnSecondAttempt(t *testing.T) {
	srv, calls := server(t, reply(500, `{}`), reply(200, completion("stop", goodContent)))
	s, err := client(srv.URL).Suggest(context.Background(), "s", "r")
	if err != nil || s.Command != "find . -size +1G" || *calls != 2 {
		t.Fatalf("s=%+v err=%v calls=%d", s, err, *calls)
	}
}

func TestLengthFailsWithoutRetry(t *testing.T) {
	srv, calls := server(t, reply(200, completion("length", `{"description":"x","comm`)))
	_, err := client(srv.URL).Suggest(context.Background(), "s", "r")
	want := "model ran out of tokens; raise the token limit in params (profile: test)"
	if err == nil || err.Error() != want || *calls != 1 {
		t.Fatalf("err=%v calls=%d", err, *calls)
	}
}

func TestClientErrorsNeverRetry(t *testing.T) {
	for _, status := range []int{400, 401} {
		srv, calls := server(t, reply(status, `{"error":{"message":"nope"}}`))
		_, err := client(srv.URL).Suggest(context.Background(), "s", "r")
		if err == nil || *calls != 1 || !strings.HasSuffix(err.Error(), ": nope") {
			t.Errorf("%d: err=%v calls=%d", status, err, *calls)
		}
	}
}

func TestStructuredOutputHint(t *testing.T) {
	srv, _ := server(t, reply(400, `{"error":{"message":"Invalid parameter: 'response_format' of type 'json_schema' is not supported with this model."}}`))
	_, err := client(srv.URL).Suggest(context.Background(), "s", "r")
	if err == nil || !strings.HasSuffix(err.Error(), "\nthis model or endpoint does not support structured outputs") {
		t.Fatalf("err = %v", err)
	}
}

func TestErrorBodyStringForm(t *testing.T) {
	srv, _ := server(t, reply(404, `{"error":"model 'qwen9' not found"}`))
	_, err := client(srv.URL).Suggest(context.Background(), "s", "r")
	if err == nil || err.Error() != "404 Not Found: model 'qwen9' not found" {
		t.Fatalf("err = %v", err)
	}
}

func TestErrorBodyRawIsTruncated(t *testing.T) {
	srv, _ := server(t, reply(403, strings.Repeat("x", 1000)))
	_, err := client(srv.URL).Suggest(context.Background(), "s", "r")
	if err == nil || err.Error() != "403 Forbidden: "+strings.Repeat("x", 300) {
		t.Fatalf("err = %v", err)
	}
}

func TestRefusalDoesNotRetry(t *testing.T) {
	body := `{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":null,"refusal":"I can't help with that."}}]}`
	srv, calls := server(t, reply(200, body))
	_, err := client(srv.URL).Suggest(context.Background(), "s", "r")
	if err == nil || err.Error() != "model refused: I can't help with that." || *calls != 1 {
		t.Fatalf("err=%v calls=%d", err, *calls)
	}
}

func TestCommandIsTrimmed(t *testing.T) {
	content := `{"description":"Two steps.","command":"\n  cd /tmp &&\n  ls -la  \n"}`
	srv, _ := server(t, reply(200, completion("stop", content)))
	s, err := client(srv.URL).Suggest(context.Background(), "s", "r")
	if err != nil || s.Command != "cd /tmp &&\n  ls -la" {
		t.Fatalf("s=%q err=%v", s.Command, err)
	}
}

func TestTimeoutIsHonoredAndRetried(t *testing.T) {
	srv, calls := server(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	})
	c := client(srv.URL)
	c.Timeout = 50 * time.Millisecond
	start := time.Now()
	_, err := c.Suggest(context.Background(), "s", "r")
	if err == nil || !strings.HasPrefix(err.Error(), "request failed: ") || *calls != 2 {
		t.Fatalf("err=%v calls=%d", err, *calls)
	}
	if time.Since(start) > time.Second {
		t.Fatalf("timeout not honored: took %v", time.Since(start))
	}
}
