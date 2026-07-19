package openrouter

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

	"github.com/rs/zerolog"
)

const (
	endpoint = "https://openrouter.ai/api/v1/chat/completions"

	retryBackoff = 2 * time.Second

	maxResponseBytes = 1 << 20
)

type Config struct {
	APIKey        string
	Model         string
	FallbackModel string
	MaxRetries    int
	Referer       string
}

type Client struct {
	http *http.Client
	cfg  Config
}

func NewClient(cfg Config, httpClient *http.Client) *Client {
	return &Client{http: httpClient, cfg: cfg}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Schema struct {
	Name       string
	Definition map[string]any
}

var errUnauthorized = errors.New("openrouter: unauthorized")

func (c *Client) CompleteJSON(ctx context.Context, messages []Message, schema Schema, out any) error {
	log := zerolog.Ctx(ctx)

	models := []string{c.cfg.Model}
	if c.cfg.FallbackModel != "" && c.cfg.FallbackModel != c.cfg.Model {
		models = append(models, c.cfg.FallbackModel)
	}

	var lastErr error

	for _, model := range models {
		for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
			if attempt > 0 {
				if err := sleep(ctx, retryBackoff<<(attempt-1)); err != nil {
					return err
				}
			}

			err := c.attempt(ctx, model, messages, schema, out)
			if err == nil {
				if model != c.cfg.Model {
					log.Warn().Str("model", model).Msg("openrouter: served by the fallback model")
				}

				return nil
			}

			if errors.Is(err, errUnauthorized) || ctx.Err() != nil {
				return err
			}

			lastErr = err

			log.Warn().Err(err).Str("model", model).Int("attempt", attempt+1).Msg("openrouter: attempt failed")
		}
	}

	return fmt.Errorf("openrouter: all models failed, last error: %w", lastErr)
}

func (c *Client) attempt(ctx context.Context, model string, messages []Message, schema Schema, out any) error {
	content, err := c.call(ctx, model, messages, schema)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(unfence(content)), out); err != nil {
		return fmt.Errorf("decode model output: %w", err)
	}

	return nil
}

type request struct {
	Model          string          `json:"model"`
	Messages       []Message       `json:"messages"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
	Temperature    float64         `json:"temperature"`
}

type responseFormat struct {
	Type       string     `json:"type"`
	JSONSchema jsonSchema `json:"json_schema"`
}

type jsonSchema struct {
	Name   string         `json:"name"`
	Strict bool           `json:"strict"`
	Schema map[string]any `json:"schema"`
}

type response struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

func (c *Client) call(ctx context.Context, model string, messages []Message, schema Schema) (string, error) {
	body, err := json.Marshal(request{
		Model:       model,
		Messages:    messages,
		Temperature: 0.2,
		ResponseFormat: &responseFormat{
			Type: "json_schema",
			JSONSchema: jsonSchema{
				Name:   schema.Name,
				Strict: true,
				Schema: schema.Definition,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	if c.cfg.Referer != "" {
		req.Header.Set("HTTP-Referer", c.cfg.Referer)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("call openrouter: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("%w: %s", errUnauthorized, strings.TrimSpace(string(raw)))
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openrouter %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}

	var decoded response
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if decoded.Error != nil {
		return "", fmt.Errorf("openrouter: %s (code %d)", decoded.Error.Message, decoded.Error.Code)
	}

	if len(decoded.Choices) == 0 || decoded.Choices[0].Message.Content == "" {
		return "", errors.New("openrouter: empty completion")
	}

	return decoded.Choices[0].Message.Content, nil
}

func unfence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}

	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[i+1:]
	}

	return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "```"))
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
