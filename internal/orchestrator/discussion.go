package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/coah80/socratic-slopinar/internal/config"
	"github.com/coah80/socratic-slopinar/internal/openrouter"
)

type DiscussionResult struct {
	Messages       []openrouter.ChatMessage
	Notes          string
	ExecutionPrompt string
	PinnedMessages []string
	NameMap        map[string]string
}

var tagPattern = regexp.MustCompile(`\[[\w\-./]+\]:\s*`)
var toolCallPattern = regexp.MustCompile(`(?s)<[｜|]?(?:tool_call|function_calls?|invoke|parameter|DSML|minimax)[｜|]?[\s>].*$`)
var markdownHeaderPattern = regexp.MustCompile(`(?m)^#{1,4}\s+.*$`)
var emojiHeaderPattern = regexp.MustCompile(`(?m)^[🏁🎯✅❌📋🔧💡🚀]\s*\*\*.*?\*\*`)
var progressPattern = regexp.MustCompile(`(?i)progress update|implementation summary|key evidence|quick summary`)

func stripTags(s string) string {
	cleaned := strings.TrimSpace(tagPattern.ReplaceAllString(s, ""))
	cleaned = strings.TrimSpace(toolCallPattern.ReplaceAllString(cleaned, ""))
	cleaned = markdownHeaderPattern.ReplaceAllStringFunc(cleaned, func(h string) string {
		content := strings.TrimLeft(h, "# ")
		return content
	})
	cleaned = emojiHeaderPattern.ReplaceAllStringFunc(cleaned, func(h string) string {
		h = strings.TrimSpace(h)
		if len(h) > 0 {
			rs := []rune(h)
			for i, r := range rs {
				if r < 128 {
					return strings.Trim(string(rs[i:]), "* ")
				}
			}
		}
		return h
	})
	cleaned = progressPattern.ReplaceAllString(cleaned, "")
	cleaned = strings.ReplaceAll(cleaned, "I'd be happy to", "")
	cleaned = strings.ReplaceAll(cleaned, "I'd be happy to", "")
	return strings.TrimSpace(cleaned)
}

type Event struct {
	Type        string      `json:"type"`
	ModelID     string      `json:"model_id,omitempty"`
	DisplayName string      `json:"display_name,omitempty"`
	Content     interface{} `json:"content"`
}

type Discussion struct {
	ID           string
	Prompt       string
	CodebasePath string
	Models       []string
	Messages     []openrouter.ChatMessage
	Notes        string
	Status       string
	MaxRounds    int
}

func NewDiscussion(id, prompt, codebasePath string, models []string, maxRounds int) Discussion {
	if maxRounds <= 0 {
		maxRounds = 5
	}
	return Discussion{
		ID:           id,
		Prompt:       prompt,
		CodebasePath: codebasePath,
		Models:       models,
		Messages:     []openrouter.ChatMessage{},
		Notes:        "",
		Status:       "running",
		MaxRounds:    maxRounds,
	}
}

func shortName(modelID string) string {
	parts := strings.Split(modelID, "/")
	name := parts[len(parts)-1]
	name = strings.Split(name, ":")[0]
	for _, strip := range []string{"-next", "-fast", "-free", "-preview", "-latest"} {
		name = strings.TrimSuffix(name, strip)
	}
	dashParts := strings.Split(name, "-")
	if len(dashParts) > 0 {
		known := map[string]string{
			"claude": "Claude", "gemini": "Gemini", "grok": "Grok",
			"qwen3": "Qwen", "qwen": "Qwen", "deepseek": "DeepSeek",
			"minimax": "MiniMax", "nemotron": "Nemotron", "llama": "Llama",
			"mistral": "Mistral", "mixtral": "Mixtral", "phi": "Phi",
			"command": "Command", "nova": "Nova", "codestral": "Codestral",
		}
		if nice, ok := known[strings.ToLower(dashParts[0])]; ok {
			return nice
		}
	}
	if len(name) > 12 {
		name = name[:12]
	}
	return name
}

func buildNameMap(models []string) map[string]string {
	names := make(map[string]string)
	used := make(map[string]int)
	for _, m := range models {
		short := shortName(m)
		used[short]++
		names[m] = short
	}
	for m, short := range names {
		if used[short] > 1 {
			parts := strings.Split(m, "/")
			names[m] = parts[len(parts)-1]
		}
	}
	return names
}

func systemPrompt(prompt, codebasePath string, thisModel string, allModels []string, nameMap map[string]string) string {
	myName := nameMap[thisModel]
	var otherNames []string
	for _, m := range allModels {
		if m != thisModel {
			otherNames = append(otherNames, nameMap[m])
		}
	}
	otherList := strings.Join(otherNames, ", ")

	example1 := otherNames[0%len(otherNames)]
	example2 := otherNames[len(otherNames)-1]

	return fmt.Sprintf(
		"YOUR NAME: %s\n\n"+
			"You're in a dev group chat with: %s\n\n"+
			"Topic: %s\nCodebase: %s\n\n"+
			"You are a developer on a team. This is a slack channel. Talk like a real person.\n"+
			"You're talking TO THE OTHER DEVS, not to a user or client. A human is spectating but isn't part of the convo.\n\n"+
			"HOW TO TALK:\n"+
			"- \"yo %s that's actually solid, what if we also...\"\n"+
			"- \"%s nah that's overengineered, just use sqlite\"\n"+
			"- \"ok hear me out - what about doing X instead\"\n"+
			"- \"wait actually i just checked the code and there's already a...\"\n"+
			"- \"disagree, the latency would be terrible\"\n"+
			"- \"lmao yeah that'd break everything\"\n"+
			"- \"honestly i think we're overcomplicating this\"\n\n"+
			"BANNED PHRASES (using these = instant cringe):\n"+
			"\"I'd be happy to\", \"Great question\", \"Shall I proceed\", \"Let me break this down\",\n"+
			"\"Absolutely\", \"That's a fantastic point\", \"Based on my exploration\",\n"+
			"\"Here's what I found\", \"Let me provide\", \"Progress Update\",\n"+
			"\"Implementation Summary\", \"Key Evidence\", \"After analyzing\",\n"+
			"\"Let me start by exploring\", \"I can confirm\", \"I want to understand\"\n\n"+
			"BANNED FORMATTING:\n"+
			"- NO markdown headers (# or ##)\n"+
			"- NO emoji bullet headers (🏁 **thing**)\n"+
			"- NO numbered lists with bold headers\n"+
			"- NO \"## Summary\" or \"## Overview\" sections\n"+
			"- Just write plain sentences like a normal person in a chat\n\n"+
			"SHARED NOTES:\n"+
			"The group has a shared markdown doc. Use update_notes to record plans and decisions.\n"+
			"Markdown formatting is fine IN THE NOTES (that's a doc, not chat).\n"+
			"If your context gets compacted, the notes have everything discussed so far.\n\n"+
			"RULES:\n"+
			"- You are %s. Don't confuse yourself with anyone else.\n"+
			"- 2-4 sentences max. SHORT. This is chat, not an essay.\n"+
			"- If someone already said something, don't repeat it. Say \"agreed\" and move on.\n"+
			"- Use tools only when you need to check something specific. 1-2 max per turn.\n"+
			"- Use update_notes to record decisions as you go.\n"+
			"- Disagree when something's wrong. Don't just agree with everyone.\n\n"+
			"HARD NOs:\n"+
			"- Don't offer to implement or write code. Planning only.\n"+
			"- Don't ask permission. Don't ask the human anything.\n"+
			"- Don't summarize what you're about to do. Just do it.\n"+
			"- Don't write analysis reports. This is a chat, not a document.\n"+
			"- Don't repeat the same point another model already made.\n"+
			"- Don't say what model you are or announce yourself.",
		myName, otherList, prompt, codebasePath,
		example1, example2, myName,
	)
}

const maxToolCallsPerTurn = 5

type modelResult struct {
	modelID string
	resp    openrouter.ChatResponse
	err     error
}

func parallelRound(
	ctx context.Context,
	client *openrouter.Client,
	disc Discussion,
	messages []openrouter.ChatMessage,
	notes string,
	toolDefs []openrouter.ToolDefinition,
	nameMap map[string]string,
	broadcast func(Event),
	mutes *MuteSet,
	pins *PinSet,
) ([]openrouter.ChatMessage, string) {
	activeModels := mutes.ActiveModels(disc.Models)
	if len(activeModels) == 0 {
		return messages, notes
	}

	log.Printf("[DISC %s] Firing %d models in parallel", disc.ID, len(activeModels))
	for _, m := range activeModels {
		broadcast(Event{Type: "status", ModelID: m, Content: "thinking..."})
	}

	results := make(chan modelResult, len(activeModels))
	for _, modelID := range activeModels {
		go func(mid string) {
			sysmsg := openrouter.ChatMessage{
				Role:    "system",
				Content: systemPrompt(disc.Prompt, disc.CodebasePath, mid, disc.Models, nameMap),
			}
			msgs := append([]openrouter.ChatMessage{sysmsg}, withNotesContext(messages, notes)...)
			resp, err := chatWithRetry(ctx, client, mid, msgs, toolDefs, nameMap[mid], disc.ID)
			results <- modelResult{modelID: mid, resp: resp, err: err}
		}(modelID)
	}

	updatedMessages := cloneMessages(messages)
	updatedNotes := notes
	for i := 0; i < len(activeModels); i++ {
		r := <-results
		if r.err != nil {
			log.Printf("[DISC %s] [%s] ERROR: %s", disc.ID, nameMap[r.modelID], r.err.Error())
			broadcast(Event{Type: "error", ModelID: r.modelID, Content: r.err.Error()})
			continue
		}
		if len(r.resp.Choices) == 0 {
			broadcast(Event{Type: "error", ModelID: r.modelID, Content: "no response"})
			continue
		}
		msg := r.resp.Choices[0].Message
		log.Printf("[DISC %s] [%s] got: %d chars, %d tools", disc.ID, nameMap[r.modelID], len(msg.Content), len(msg.ToolCalls))
		updatedMessages, updatedNotes = handleModelResponse(ctx, client, r.modelID, msg, updatedMessages, updatedNotes, toolDefs, disc.CodebasePath, broadcast, nameMap, 0, pins)
	}
	return updatedMessages, updatedNotes
}

func sequentialRound(
	ctx context.Context,
	client *openrouter.Client,
	disc Discussion,
	messages []openrouter.ChatMessage,
	notes string,
	toolDefs []openrouter.ToolDefinition,
	nameMap map[string]string,
	broadcast func(Event),
	mutes *MuteSet,
	pins *PinSet,
) ([]openrouter.ChatMessage, string) {
	for _, modelID := range disc.Models {
		if ctx.Err() != nil {
			return messages, notes
		}
		if mutes.IsMuted(modelID) {
			continue
		}

		log.Printf("[DISC %s] [%s] requesting (%d msgs)", disc.ID, nameMap[modelID], len(messages))
		broadcast(Event{Type: "status", ModelID: modelID, Content: "thinking..."})

		sysmsg := openrouter.ChatMessage{
			Role:    "system",
			Content: systemPrompt(disc.Prompt, disc.CodebasePath, modelID, disc.Models, nameMap),
		}
		currentMessages := append([]openrouter.ChatMessage{sysmsg}, withNotesContext(messages, notes)...)

		resp, err := chatWithRetry(ctx, client, modelID, currentMessages, toolDefs, nameMap[modelID], disc.ID)
		if err != nil {
			log.Printf("[DISC %s] [%s] ERROR: %s", disc.ID, nameMap[modelID], err.Error())
			broadcast(Event{Type: "error", ModelID: modelID, Content: err.Error()})
			continue
		}
		if len(resp.Choices) == 0 {
			broadcast(Event{Type: "error", ModelID: modelID, Content: "no response"})
			continue
		}

		msg := resp.Choices[0].Message
		log.Printf("[DISC %s] [%s] got: %d chars, %d tools", disc.ID, nameMap[modelID], len(msg.Content), len(msg.ToolCalls))
		messages, notes = handleModelResponse(ctx, client, modelID, msg, messages, notes, toolDefs, disc.CodebasePath, broadcast, nameMap, 0, pins)
	}
	return messages, notes
}

func countAgreements(messages []openrouter.ChatMessage, lookback int) int {
	agreePhrases := []string{
		"agreed", "i agree", "sounds good", "that works", "let's go with",
		"spot on", "solid plan", "nailed it", "makes sense", "on board",
		"same page", "consensus", "let's do it", "ship it",
	}
	count := 0
	start := len(messages) - lookback
	if start < 0 {
		start = 0
	}
	for _, m := range messages[start:] {
		if m.Role != "assistant" {
			continue
		}
		lower := strings.ToLower(m.Content)
		for _, phrase := range agreePhrases {
			if strings.Contains(lower, phrase) {
				count++
				break
			}
		}
	}
	return count
}

func chatWithRetry(ctx context.Context, client *openrouter.Client, modelID string, messages []openrouter.ChatMessage, tools []openrouter.ToolDefinition, displayName string, discID string) (openrouter.ChatResponse, error) {
	resp, err := client.Chat(ctx, openrouter.ChatRequest{
		Model:    modelID,
		Messages: messages,
		Tools:    tools,
	})
	if err == nil {
		return resp, nil
	}

	log.Printf("[DISC %s] [%s] error: %s, retrying with trimmed context + no tools", discID, displayName, err.Error())
	if len(messages) <= 2 {
		return resp, err
	}

	sys := messages[0]
	rest := messages[1:]
	keep := 6
	if keep > len(rest) {
		keep = len(rest)
	}
	kept := sanitizeMessages(rest[len(rest)-keep:])
	trimmed := append([]openrouter.ChatMessage{sys}, kept...)
	log.Printf("[DISC %s] [%s] retry: %d msgs (was %d), no tools", discID, displayName, len(trimmed), len(messages))

	resp2, err2 := client.Chat(ctx, openrouter.ChatRequest{
		Model:    modelID,
		Messages: trimmed,
	})
	if err2 != nil {
		return resp, fmt.Errorf("retry also failed: %w", err2)
	}
	return resp2, nil
}

func Run(ctx context.Context, disc Discussion, client *openrouter.Client, rawBroadcast func(Event), mutes *MuteSet, pins *PinSet, injector *Injector) DiscussionResult {
	nameMap := buildNameMap(disc.Models)
	broadcast := func(e Event) {
		if e.ModelID != "" {
			e.DisplayName = nameMap[e.ModelID]
		}
		rawBroadcast(e)
	}

	log.Printf("[DISC %s] Starting: %d models, %d rounds", disc.ID, len(disc.Models), disc.MaxRounds)
	log.Printf("[DISC %s] Prompt: %s", disc.ID, disc.Prompt)
	for _, m := range disc.Models {
		log.Printf("[DISC %s]   %s → %s", disc.ID, m, nameMap[m])
	}
	broadcast(Event{Type: "status", Content: "starting discussion"})

	toolDefs := AllToolDefinitions()
	notes := disc.Notes
	messages := append([]openrouter.ChatMessage{}, disc.Messages...)

	agreementCount := 0

	for round := 0; round < disc.MaxRounds; round++ {
		if ctx.Err() != nil {
			broadcast(Event{Type: "status", Content: "stopped"})
			return DiscussionResult{Messages: messages, Notes: notes, PinnedMessages: pins.All(), NameMap: nameMap}
		}

		injected := injector.Drain()
		for _, msg := range injected {
			messages = append(cloneMessages(messages), msg)
			broadcast(Event{Type: "message", ModelID: "god", DisplayName: "God", Content: strings.TrimPrefix(msg.Content, "[God]: ")})
		}

		log.Printf("[DISC %s] === Round %d/%d ===", disc.ID, round+1, disc.MaxRounds)
		broadcast(Event{Type: "status", Content: fmt.Sprintf("round %d/%d", round+1, disc.MaxRounds)})

		if round == 0 {
			messages, notes = parallelRound(ctx, client, disc, messages, notes, toolDefs, nameMap, broadcast, mutes, pins)
		} else {
			messages, notes = sequentialRound(ctx, client, disc, messages, notes, toolDefs, nameMap, broadcast, mutes, pins)
		}

		agreementCount = countAgreements(messages, 6)
		if agreementCount >= 3 && round < disc.MaxRounds-1 {
			log.Printf("[DISC %s] Convergence detected (%d agreements), wrapping up", disc.ID, agreementCount)
			broadcast(Event{Type: "status", Content: "consensus reached, generating plan"})
			break
		}
	}

	log.Printf("[DISC %s] Generating execution prompt", disc.ID)
	execPrompt := generateExecutionPrompt(ctx, client, disc.Models, messages, notes)
	broadcast(Event{Type: "execution_prompt", Content: execPrompt})
	broadcast(Event{Type: "status", Content: "done"})
	log.Printf("[DISC %s] Done", disc.ID)

	return DiscussionResult{
		Messages:        messages,
		Notes:           notes,
		ExecutionPrompt: execPrompt,
		PinnedMessages:  pins.All(),
		NameMap:         nameMap,
	}
}

func BuildRecord(disc Discussion, result DiscussionResult) config.DiscussionRecord {
	records := make([]config.MessageRecord, 0, len(result.Messages))
	for _, m := range result.Messages {
		if m.Role != "assistant" {
			continue
		}
		modelID := ""
		displayName := ""
		for id, name := range result.NameMap {
			prefix := "[" + name + "]: "
			if strings.HasPrefix(m.Content, prefix) {
				modelID = id
				displayName = name
				break
			}
		}
		content := m.Content
		for _, name := range result.NameMap {
			prefix := "[" + name + "]: "
			content = strings.TrimPrefix(content, prefix)
		}
		records = append(records, config.MessageRecord{
			ModelID:     modelID,
			DisplayName: displayName,
			Content:     content,
			Timestamp:   time.Now().Unix(),
		})
	}
	return config.DiscussionRecord{
		ID:              disc.ID,
		Prompt:          disc.Prompt,
		CodebasePath:    disc.CodebasePath,
		Models:          disc.Models,
		Messages:        records,
		SharedNotes:     result.Notes,
		ExecutionPrompt: result.ExecutionPrompt,
		PinnedMessages:  result.PinnedMessages,
		CreatedAt:       time.Now().Unix(),
	}
}

func handleModelResponse(
	ctx context.Context,
	client *openrouter.Client,
	modelID string,
	msg openrouter.ChatMessage,
	messages []openrouter.ChatMessage,
	notes string,
	toolDefs []openrouter.ToolDefinition,
	codebasePath string,
	broadcast func(Event),
	nameMap map[string]string,
	totalToolCalls int,
	pins *PinSet,
) ([]openrouter.ChatMessage, string) {

	cleanContent := stripTags(msg.Content)
	if cleanContent != "" {
		broadcast(Event{Type: "message", ModelID: modelID, Content: cleanContent})
	}

	taggedContent := fmt.Sprintf("[%s]: %s", nameMap[modelID], cleanContent)
	assistantMsg := openrouter.ChatMessage{
		Role:      "assistant",
		Content:   taggedContent,
		ToolCalls: msg.ToolCalls,
	}
	updatedMessages := append(cloneMessages(messages), assistantMsg)

	if len(msg.ToolCalls) == 0 {
		return updatedMessages, notes
	}

	newTotal := totalToolCalls + len(msg.ToolCalls)
	if newTotal > maxToolCallsPerTurn {
		log.Printf("[DISC] [%s] hit tool cap (%d), forcing text response", nameMap[modelID], newTotal)
		return forceTextResponse(ctx, client, modelID, updatedMessages, notes, broadcast, nameMap)
	}

	updatedNotes := notes
	for _, tc := range msg.ToolCalls {
		log.Printf("[TOOL] [%s] %s(%s)", nameMap[modelID], tc.Function.Name, tc.Function.Arguments)
		broadcast(Event{
			Type:    "tool_call",
			ModelID: modelID,
			Content: map[string]string{"name": tc.Function.Name, "arguments": tc.Function.Arguments},
		})

		result, err := ExecuteTool(tc.Function.Name, codebasePath, tc.Function.Arguments, &updatedNotes, pins)
		if err != nil {
			log.Printf("[TOOL] [%s] error: %s", nameMap[modelID], err.Error())
			result = "error: " + err.Error()
		} else {
			log.Printf("[TOOL] [%s] %s → %d chars", nameMap[modelID], tc.Function.Name, len(result))
		}

		if tc.Function.Name == "update_notes" {
			broadcast(Event{Type: "notes_update", ModelID: modelID, Content: updatedNotes})
		}
		if tc.Function.Name == "pin_message" && err == nil {
			var pinArgs struct {
				Message string `json:"message"`
			}
			_ = json.Unmarshal([]byte(tc.Function.Arguments), &pinArgs)
			broadcast(Event{Type: "pin", ModelID: modelID, Content: pinArgs.Message})
		}

		displayResult := result
		if len(displayResult) > 3000 {
			displayResult = displayResult[:3000] + "\n... (truncated)"
		}
		broadcast(Event{
			Type:    "tool_result",
			ModelID: modelID,
			Content: map[string]string{"name": tc.Function.Name, "result": displayResult},
		})

		contextResult := result
		if len(contextResult) > 800 {
			contextResult = contextResult[:800] + "\n... (truncated, " + fmt.Sprintf("%d", len(result)) + " chars total)"
		}
		updatedMessages = append(updatedMessages, openrouter.ChatMessage{
			Role:       "tool",
			Content:    contextResult,
			ToolCallID: tc.ID,
		})
	}

	followup, err := client.Chat(ctx, openrouter.ChatRequest{
		Model:    modelID,
		Messages: updatedMessages,
		Tools:    toolDefs,
	})
	if err != nil {
		broadcast(Event{Type: "error", ModelID: modelID, Content: err.Error()})
		return updatedMessages, updatedNotes
	}

	if len(followup.Choices) > 0 {
		fmsg := followup.Choices[0].Message
		if len(fmsg.ToolCalls) > 0 {
			return handleModelResponse(ctx, client, modelID, fmsg, updatedMessages, updatedNotes, toolDefs, codebasePath, broadcast, nameMap, newTotal, pins)
		}
		fClean := stripTags(fmsg.Content)
		if fClean != "" {
			broadcast(Event{Type: "message", ModelID: modelID, Content: fClean})
			updatedMessages = append(updatedMessages, openrouter.ChatMessage{
				Role:    "assistant",
				Content: fmt.Sprintf("[%s]: %s", nameMap[modelID], fClean),
			})
		}
	}

	return updatedMessages, updatedNotes
}

func forceTextResponse(
	ctx context.Context,
	client *openrouter.Client,
	modelID string,
	messages []openrouter.ChatMessage,
	notes string,
	broadcast func(Event),
	nameMap map[string]string,
) ([]openrouter.ChatMessage, string) {
	broadcast(Event{Type: "status", ModelID: modelID, Content: "wrapping up tools..."})
	resp, err := client.Chat(ctx, openrouter.ChatRequest{
		Model:    modelID,
		Messages: messages,
	})
	if err != nil {
		broadcast(Event{Type: "error", ModelID: modelID, Content: err.Error()})
		return messages, notes
	}
	if len(resp.Choices) > 0 && resp.Choices[0].Message.Content != "" {
		fClean := stripTags(resp.Choices[0].Message.Content)
		if fClean != "" {
			broadcast(Event{Type: "message", ModelID: modelID, Content: fClean})
			messages = append(cloneMessages(messages), openrouter.ChatMessage{
				Role:    "assistant",
				Content: fmt.Sprintf("[%s]: %s", nameMap[modelID], fClean),
			})
		}
	}
	return messages, notes
}

func generateExecutionPrompt(
	ctx context.Context,
	client *openrouter.Client,
	models []string,
	messages []openrouter.ChatMessage,
	notes string,
) string {
	if len(models) == 0 {
		return notes
	}
	if strings.TrimSpace(notes) == "" {
		notes = "(no notes were recorded during the discussion)"
	}

	last10 := messages
	if len(last10) > 10 {
		last10 = last10[len(last10)-10:]
	}

	summaryMessages := []openrouter.ChatMessage{
		{Role: "system", Content: "You are a technical writer. Synthesize discussion notes into a clear execution prompt."},
	}
	summaryMessages = append(summaryMessages, last10...)
	summaryMessages = append(summaryMessages, openrouter.ChatMessage{
		Role: "user",
		Content: "Based on the discussion and shared notes below, generate a clear, actionable execution prompt " +
			"that another AI could use to implement the plan. Include specific file paths, code changes, " +
			"and steps. Output only the prompt, no preamble.\n\nShared notes:\n" + notes,
	})

	for _, modelID := range models {
		resp, err := client.Chat(ctx, openrouter.ChatRequest{
			Model:    modelID,
			Messages: summaryMessages,
		})
		if err != nil {
			log.Printf("[EXEC] %s failed: %s, trying next model", modelID, err.Error())
			continue
		}
		if len(resp.Choices) > 0 && resp.Choices[0].Message.Content != "" {
			return resp.Choices[0].Message.Content
		}
	}
	return notes
}

const maxContextMessages = 25

func sanitizeMessages(msgs []openrouter.ChatMessage) []openrouter.ChatMessage {
	var out []openrouter.ChatMessage
	toolCallIDs := make(map[string]bool)

	for _, m := range msgs {
		if m.Role == "assistant" && len(m.ToolCalls) > 0 {
			for _, tc := range m.ToolCalls {
				toolCallIDs[tc.ID] = true
			}
		}
	}

	for _, m := range msgs {
		if m.Role == "tool" && m.ToolCallID != "" {
			if !toolCallIDs[m.ToolCallID] {
				continue
			}
		}
		out = append(out, m)
	}

	if len(out) > 0 && out[len(out)-1].Role == "assistant" {
		out = append(out, openrouter.ChatMessage{
			Role:    "user",
			Content: "[Continue the discussion]",
		})
	}

	return out
}

func withNotesContext(messages []openrouter.ChatMessage, notes string) []openrouter.ChatMessage {
	trimmed := messages
	wasCompacted := false
	if len(trimmed) > maxContextMessages {
		trimmed = trimmed[len(trimmed)-maxContextMessages:]
		wasCompacted = true
	}

	cloned := sanitizeMessages(cloneMessages(trimmed))

	if wasCompacted && strings.TrimSpace(notes) != "" {
		catchup := openrouter.ChatMessage{
			Role: "user",
			Content: "[Context was compacted - earlier messages were removed to save space. " +
				"The shared notes below contain everything discussed so far. " +
				"Read them and continue the discussion from where it left off.]\n\n" +
				"[Shared Notes]\n" + notes,
		}
		cloned = append([]openrouter.ChatMessage{catchup}, cloned...)
	} else if strings.TrimSpace(notes) != "" {
		cloned = append(cloned, openrouter.ChatMessage{
			Role:    "user",
			Content: "[Shared Notes]\n" + notes,
		})
	}
	return cloned
}

func cloneMessages(msgs []openrouter.ChatMessage) []openrouter.ChatMessage {
	out := make([]openrouter.ChatMessage, len(msgs))
	copy(out, msgs)
	return out
}
