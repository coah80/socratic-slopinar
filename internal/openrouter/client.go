package openrouter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	providerKeys map[string]string
	httpClient   *http.Client
}

func NewClient(providerKeys map[string]string) *Client {
	keys := make(map[string]string, len(providerKeys))
	for k, v := range providerKeys {
		keys[k] = v
	}
	return &Client{
		providerKeys: keys,
		httpClient:   &http.Client{},
	}
}

func (c *Client) resolveRequest(modelID string) (url string, resolvedModel string, provider Provider, apiKey string, err error) {
	providerID := DetectProvider(modelID)
	provider = GetProvider(providerID)

	if providerID == "anthropic" {
		log.Printf("[PROVIDER] Anthropic detected for %s, routing through OpenRouter (incompatible API format)", modelID)
		providerID = "openrouter"
		provider = GetProvider("openrouter")
	}

	apiKey, hasKey := c.providerKeys[providerID]
	if !hasKey || apiKey == "" {
		apiKey, hasKey = c.providerKeys["openrouter"]
		if !hasKey || apiKey == "" {
			return "", "", Provider{}, "", fmt.Errorf("no API key for provider %q and no OpenRouter fallback", providerID)
		}
		provider = GetProvider("openrouter")
		resolvedModel = modelID
		log.Printf("[PROVIDER] No key for %s, falling back to OpenRouter for %s", providerID, modelID)
	} else {
		resolvedModel = modelID
		if provider.StripPrefix {
			resolvedModel = stripModelPrefix(modelID)
		}
		if providerID != "openrouter" {
			log.Printf("[PROVIDER] Routing %s directly to %s (model: %s)", modelID, provider.Name, resolvedModel)
		}
	}

	return provider.BaseURL, resolvedModel, provider, apiKey, nil
}

func (c *Client) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	url, resolvedModel, provider, apiKey, err := c.resolveRequest(req.Model)
	if err != nil {
		return ChatResponse{}, err
	}

	r := ChatRequest{
		Model:      resolvedModel,
		Messages:   req.Messages,
		Tools:      req.Tools,
		Stream:     false,
		ToolChoice: req.ToolChoice,
	}

	body, err := json.Marshal(r)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	callCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(callCtx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return ChatResponse{}, err
	}
	setProviderHeaders(httpReq, provider, apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ChatResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return ChatResponse{}, fmt.Errorf("%s %d: %s", provider.Name, resp.StatusCode, string(b))
	}

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ChatResponse{}, err
	}
	return result, nil
}

func (c *Client) StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	url, resolvedModel, provider, apiKey, err := c.resolveRequest(req.Model)
	if err != nil {
		return nil, err
	}

	r := ChatRequest{
		Model:      resolvedModel,
		Messages:   req.Messages,
		Tools:      req.Tools,
		Stream:     true,
		ToolChoice: req.ToolChoice,
	}

	body, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	setProviderHeaders(httpReq, provider, apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s %d: %s", provider.Name, resp.StatusCode, string(b))
	}

	ch := make(chan StreamChunk, 64)
	go func() {
		defer resp.Body.Close()
		defer close(ch)
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}
			var chunk StreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}
			select {
			case ch <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

func setProviderHeaders(req *http.Request, provider Provider, apiKey string) {
	req.Header.Set(provider.AuthHeader, provider.AuthPrefix+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "http://localhost:8080")
}
