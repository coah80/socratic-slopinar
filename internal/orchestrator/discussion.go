package orchestrator

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/coah80/socratic-slopinar/internal/openrouter"
)

var tagPattern = regexp.MustCompile(`\[[\w\-./]+\]:\s*`)
var toolCallPattern = regexp.MustCompile(`(?s)<(tool_call|function_calls?|invoke|parameter)[\s>].*$`)

func stripTags(s string) string {
	cleaned := strings.TrimSpace(tagPattern.ReplaceAllString(s, ""))
	cleaned = strings.TrimSpace(toolCallPattern.ReplaceAllString(cleaned, ""))
	return cleaned
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
			"This is a casual roundtable. You're all devs hashing out a plan together.\n"+
			"Talk to the OTHER MODELS, not a human. The human is just watching.\n\n"+
			"TONE - talk like these examples:\n"+
			"- \"yo %s that's actually solid, what if we also...\"\n"+
			"- \"%s nah that's overengineered, just use sqlite\"\n"+
			"- \"ok hear me out - what about doing X instead\"\n"+
			"- \"wait actually i just checked the code and there's already a...\"\n"+
			"- \"disagree, the latency would be terrible\"\n\n"+
			"NEVER say: \"I'd be happy to\", \"Great question\", \"Shall I proceed\", "+
			"\"Let me break this down\", \"Absolutely\", \"That's a fantastic point\", "+
			"\"Based on my exploration\", \"Here's what I found\", \"Let me provide\"\n"+
			"NEVER use markdown headers (##) or bullet point analysis. Just talk normally.\n\n"+
			"SHARED NOTES:\n"+
			"The group has a shared notes document (markdown). This is your team's living doc.\n"+
			"Use update_notes to add decisions, plans, and findings. Keep it organized.\n"+
			"If your context gets compacted, the notes will tell you everything discussed so far.\n"+
			"Treat the notes like a collaborative google doc - add to it, restructure it, keep it clean.\n\n"+
			"RULES:\n"+
			"- You are %s. Don't confuse yourself with anyone else.\n"+
			"- 2-4 sentences max. Be concise.\n"+
			"- Use tools SPARINGLY - max 1-2 per turn. Don't explore the whole codebase.\n"+
			"- Only read a file if you need specific info to make your point.\n"+
			"- ACTIVELY use update_notes to record decisions and plans as you discuss.\n"+
			"- Push back on bad ideas. Hype good ones. Be real.\n\n"+
			"DO NOT:\n"+
			"- Offer to implement anything. You're planning, not coding.\n"+
			"- Ask for permission or confirmation.\n"+
			"- Start your message with your name or anyone else's name in brackets.\n"+
			"- Talk to the human. They're not here.\n"+
			"- Be formal or robotic.",
		myName, otherList, prompt, codebasePath,
		example1, example2, myName,
	)
}

const maxToolCallsPerTurn = 5

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
	trimmed := append([]openrouter.ChatMessage{sys}, rest[len(rest)-keep:]...)
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

func Run(ctx context.Context, disc Discussion, client *openrouter.Client, rawBroadcast func(Event)) {
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

	for round := 0; round < disc.MaxRounds; round++ {
		if ctx.Err() != nil {
			broadcast(Event{Type: "status", Content: "stopped"})
			return
		}

		log.Printf("[DISC %s] === Round %d/%d ===", disc.ID, round+1, disc.MaxRounds)
		broadcast(Event{Type: "status", Content: fmt.Sprintf("round %d/%d", round+1, disc.MaxRounds)})

		for _, modelID := range disc.Models {
			if ctx.Err() != nil {
				broadcast(Event{Type: "status", Content: "stopped"})
				return
			}

			log.Printf("[DISC %s] [%s] requesting (%d msgs in ctx)", disc.ID, nameMap[modelID], len(messages))
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
				log.Printf("[DISC %s] [%s] no response choices", disc.ID, nameMap[modelID])
				broadcast(Event{Type: "error", ModelID: modelID, Content: "no response from model"})
				continue
			}

			msg := resp.Choices[0].Message
			log.Printf("[DISC %s] [%s] got: %d chars, %d tool calls", disc.ID, nameMap[modelID], len(msg.Content), len(msg.ToolCalls))
			messages, notes = handleModelResponse(ctx, client, modelID, msg, messages, notes, toolDefs, disc.CodebasePath, broadcast, nameMap, 0)
		}
	}

	log.Printf("[DISC %s] Generating execution prompt", disc.ID)
	execPrompt := generateExecutionPrompt(ctx, client, disc.Models, messages, notes)
	broadcast(Event{Type: "execution_prompt", Content: execPrompt})
	broadcast(Event{Type: "status", Content: "done"})
	log.Printf("[DISC %s] Done", disc.ID)
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

		result, err := ExecuteTool(tc.Function.Name, codebasePath, tc.Function.Arguments, &updatedNotes)
		if err != nil {
			log.Printf("[TOOL] [%s] error: %s", nameMap[modelID], err.Error())
			result = "error: " + err.Error()
		} else {
			log.Printf("[TOOL] [%s] %s → %d chars", nameMap[modelID], tc.Function.Name, len(result))
		}

		if tc.Function.Name == "update_notes" {
			broadcast(Event{Type: "notes_update", ModelID: modelID, Content: updatedNotes})
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
			return handleModelResponse(ctx, client, modelID, fmsg, updatedMessages, updatedNotes, toolDefs, codebasePath, broadcast, nameMap, newTotal)
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

func withNotesContext(messages []openrouter.ChatMessage, notes string) []openrouter.ChatMessage {
	trimmed := messages
	wasCompacted := false
	if len(trimmed) > maxContextMessages {
		trimmed = trimmed[len(trimmed)-maxContextMessages:]
		wasCompacted = true
	}
	cloned := cloneMessages(trimmed)

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
