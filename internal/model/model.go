package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type Provider interface {
	Chat(context.Context, []Message) (string, error)
}
type OpenAI struct {
	BaseURL, Key, Name string
	Client             *http.Client
}

func New(base, key, name string, development bool) (Provider, error) {
	if base == "" && key == "" && name == "" {
		return nil, nil
	}
	u, err := url.Parse(base)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || name == "" || key == "" {
		return nil, errors.New("incomplete model configuration")
	}
	if u.Scheme != "https" && !(development && u.Scheme == "http" && (u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost")) {
		return nil, errors.New("model endpoint requires HTTPS")
	}
	return &OpenAI{strings.TrimRight(base, "/"), key, name, &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("redirect denied") }}}, nil
}
func (p *OpenAI) Chat(ctx context.Context, messages []Message) (string, error) {
	body, _ := json.Marshal(map[string]any{"model": p.Name, "messages": messages, "max_tokens": 512, "stream": false})
	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", errors.New("model_request_failed")
	}
	req.Header.Set("Authorization", "Bearer "+p.Key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.Client.Do(req)
	if err != nil {
		return "", errors.New("model_unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New("model_unavailable")
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 32769))
	if err != nil || len(raw) > 32768 {
		return "", errors.New("model_output_limit")
	}
	var out struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}
	if json.Unmarshal(raw, &out) != nil || len(out.Choices) != 1 || out.Choices[0].Message.Content == "" {
		return "", errors.New("invalid_model_response")
	}
	return out.Choices[0].Message.Content, nil
}
