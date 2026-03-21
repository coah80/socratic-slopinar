package orchestrator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/coah80/socratic-slopinar/internal/config"
	"github.com/coah80/socratic-slopinar/internal/openrouter"
)

type Tool interface {
	Name() string
	Definition() openrouter.ToolDefinition
	Execute(codebasePath string, argsJSON string) (string, error)
}

func AllTools() []Tool {
	return []Tool{
		ReadFileTool{},
		ListFilesTool{},
		SearchCodeTool{},
		UpdateNotesTool{},
		WebSearchTool{},
		PinMessageTool{},
	}
}

func AllToolDefinitions() []openrouter.ToolDefinition {
	tools := AllTools()
	defs := make([]openrouter.ToolDefinition, len(tools))
	for i, t := range tools {
		defs[i] = t.Definition()
	}
	return defs
}

func ExecuteTool(name string, codebasePath string, argsJSON string, notes *string, pins *PinSet) (string, error) {
	for _, t := range AllTools() {
		if t.Name() == name {
			if name == "update_notes" {
				return executeUpdateNotes(argsJSON, notes)
			}
			if name == "pin_message" {
				return executePinMessage(argsJSON, pins)
			}
			return t.Execute(codebasePath, argsJSON)
		}
	}
	return "", fmt.Errorf("unknown tool: %s", name)
}

func validatePath(codebasePath, requestedPath string) (string, error) {
	abs, err := filepath.Abs(filepath.Join(codebasePath, requestedPath))
	if err != nil {
		return "", fmt.Errorf("invalid path")
	}
	cbAbs, err := filepath.Abs(codebasePath)
	if err != nil {
		return "", fmt.Errorf("invalid codebase path")
	}
	if !strings.HasPrefix(abs, cbAbs) {
		return "", fmt.Errorf("path escapes codebase root")
	}
	return abs, nil
}

// ------- read_file -------

type ReadFileTool struct{}

func (ReadFileTool) Name() string { return "read_file" }

func (ReadFileTool) Definition() openrouter.ToolDefinition {
	return openrouter.ToolDefinition{
		Type: "function",
		Function: openrouter.FunctionDef{
			Name:        "read_file",
			Description: "Read contents of a file in the codebase",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "File path relative to codebase root",
					},
				},
				"required": []string{"path"},
			},
		},
	}
}

func (ReadFileTool) Execute(codebasePath string, argsJSON string) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	absPath, err := validatePath(codebasePath, args.Path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}
	if len(data) > 50000 {
		return string(data[:50000]) + "\n... (truncated)", nil
	}
	return string(data), nil
}

// ------- list_files -------

type ListFilesTool struct{}

func (ListFilesTool) Name() string { return "list_files" }

func (ListFilesTool) Definition() openrouter.ToolDefinition {
	return openrouter.ToolDefinition{
		Type: "function",
		Function: openrouter.FunctionDef{
			Name:        "list_files",
			Description: "List files in a directory, optionally matching a glob pattern",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Directory path relative to codebase root",
					},
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "Glob pattern like *.go or **/*.ts",
					},
				},
				"required": []string{"path"},
			},
		},
	}
}

func (ListFilesTool) Execute(codebasePath string, argsJSON string) (string, error) {
	var args struct {
		Path    string `json:"path"`
		Pattern string `json:"pattern"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	absPath, err := validatePath(codebasePath, args.Path)
	if err != nil {
		return "", err
	}

	pattern := "*"
	if args.Pattern != "" {
		pattern = args.Pattern
	}

	matches, err := filepath.Glob(filepath.Join(absPath, pattern))
	if err != nil {
		return "", err
	}

	lines := make([]string, len(matches))
	for i, m := range matches {
		rel, _ := filepath.Rel(codebasePath, m)
		lines[i] = rel
	}
	return strings.Join(lines, "\n"), nil
}

// ------- search_code -------

type SearchCodeTool struct{}

func (SearchCodeTool) Name() string { return "search_code" }

func (SearchCodeTool) Definition() openrouter.ToolDefinition {
	return openrouter.ToolDefinition{
		Type: "function",
		Function: openrouter.FunctionDef{
			Name:        "search_code",
			Description: "Search for a pattern in the codebase using grep",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search pattern (regex supported)",
					},
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Directory to search in, relative to codebase root. Defaults to root.",
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

func (SearchCodeTool) Execute(codebasePath string, argsJSON string) (string, error) {
	var args struct {
		Query string `json:"query"`
		Path  string `json:"path"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}

	searchDir := codebasePath
	if args.Path != "" {
		p, err := validatePath(codebasePath, args.Path)
		if err != nil {
			return "", err
		}
		searchDir = p
	}

	cmd := exec.Command("grep", "-rn", "--include=*", "-m", "50", args.Query, searchDir)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "no matches found", nil
		}
		return "", err
	}
	return string(out), nil
}

// ------- update_notes -------

type UpdateNotesTool struct{}

func (UpdateNotesTool) Name() string { return "update_notes" }

func (UpdateNotesTool) Definition() openrouter.ToolDefinition {
	return openrouter.ToolDefinition{
		Type: "function",
		Function: openrouter.FunctionDef{
			Name:        "update_notes",
			Description: "Update the shared notes document. Use this to record plans, findings, and decisions.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"append", "replace_section", "clear"},
						"description": "append adds to end, replace_section replaces the full notes, clear empties them",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content to append or replace with",
					},
				},
				"required": []string{"action", "content"},
			},
		},
	}
}

func (UpdateNotesTool) Execute(_ string, _ string) (string, error) {
	return "", fmt.Errorf("update_notes must be called through ExecuteTool")
}

func executeUpdateNotes(argsJSON string, notes *string) (string, error) {
	var args struct {
		Action  string `json:"action"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}

	switch args.Action {
	case "append":
		*notes = *notes + "\n" + args.Content
	case "replace_section":
		*notes = args.Content
	case "clear":
		*notes = ""
	default:
		return "", fmt.Errorf("unknown action: %s", args.Action)
	}
	return "notes updated", nil
}

// ------- web_search -------

type WebSearchTool struct{}

func (WebSearchTool) Name() string { return "web_search" }

func (WebSearchTool) Definition() openrouter.ToolDefinition {
	return openrouter.ToolDefinition{
		Type: "function",
		Function: openrouter.FunctionDef{
			Name:        "web_search",
			Description: "Search the web for information",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

type PinMessageTool struct{}

func (PinMessageTool) Name() string { return "pin_message" }

func (PinMessageTool) Definition() openrouter.ToolDefinition {
	return openrouter.ToolDefinition{
		Type: "function",
		Function: openrouter.FunctionDef{
			Name:        "pin_message",
			Description: "Pin an important message so it stays visible to all participants",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "The message to pin",
					},
				},
				"required": []string{"message"},
			},
		},
	}
}

func (PinMessageTool) Execute(_ string, _ string) (string, error) {
	return "", fmt.Errorf("pin_message must be called through ExecuteTool")
}

func executePinMessage(argsJSON string, pins *PinSet) (string, error) {
	var args struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	if args.Message == "" {
		return "", fmt.Errorf("message is required")
	}
	pins.Add(args.Message)
	return "message pinned", nil
}

func (WebSearchTool) Execute(_ string, argsJSON string) (string, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}

	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	if cfg.TavilyKey == "" {
		return "web search not configured - add tavily_api_key in settings", nil
	}

	reqBody, err := json.Marshal(map[string]interface{}{
		"query":          args.Query,
		"max_results":    5,
		"include_answer": true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.tavily.com/search", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.TavilyKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("tavily request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("tavily API error (%d): %s", resp.StatusCode, string(body)), nil
	}

	var tavilyResp struct {
		Answer  string `json:"answer"`
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &tavilyResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	var sb strings.Builder
	if tavilyResp.Answer != "" {
		sb.WriteString("Answer: ")
		sb.WriteString(tavilyResp.Answer)
		sb.WriteString("\n\n---\n\n")
	}
	for _, r := range tavilyResp.Results {
		sb.WriteString("Title: ")
		sb.WriteString(r.Title)
		sb.WriteString("\nURL: ")
		sb.WriteString(r.URL)
		sb.WriteString("\nSnippet: ")
		sb.WriteString(r.Content)
		sb.WriteString("\n\n")
	}

	result := sb.String()
	if len(result) > 5000 {
		result = result[:5000] + "\n... (truncated)"
	}
	return result, nil
}
