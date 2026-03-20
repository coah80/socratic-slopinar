package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/coah80/socratic-slopinar/internal/config"
	"github.com/coah80/socratic-slopinar/internal/openrouter"
	"github.com/coah80/socratic-slopinar/internal/orchestrator"
)

func NewRouter(frontend http.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/config", handleGetConfig)
	mux.HandleFunc("POST /api/config", handleSetConfig)
	mux.HandleFunc("POST /api/config/models", handleAddModel)
	mux.HandleFunc("DELETE /api/config/models/{id...}", handleRemoveModel)
	mux.HandleFunc("GET /api/discuss", handleDiscuss)

	if frontend != nil {
		mux.Handle("/", frontend)
	}

	return mux
}

func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"api_key":        cfg.APIKey,
		"models":         cfg.Models,
		"tavily_api_key": cfg.TavilyKey,
	})
}

func handleSetConfig(w http.ResponseWriter, r *http.Request) {
	var body struct {
		APIKey    *string  `json:"api_key"`
		Models   []string `json:"models"`
		TavilyKey *string `json:"tavily_api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cfg, err := config.Load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated := cfg
	if body.APIKey != nil {
		updated = config.Config{APIKey: *body.APIKey, Models: updated.Models, TavilyKey: updated.TavilyKey}
	}
	if body.Models != nil {
		updated = config.Config{APIKey: updated.APIKey, Models: body.Models, TavilyKey: updated.TavilyKey}
	}
	if body.TavilyKey != nil {
		updated = config.Config{APIKey: updated.APIKey, Models: updated.Models, TavilyKey: *body.TavilyKey}
	}

	if err := config.Save(updated); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"api_key":        updated.APIKey,
		"models":         updated.Models,
		"tavily_api_key": updated.TavilyKey,
	})
}

func handleAddModel(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Model string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Model == "" {
		writeError(w, http.StatusBadRequest, "model is required")
		return
	}

	cfg, err := config.Load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated := config.AddModel(cfg, body.Model)
	if err := config.Save(updated); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"models": updated.Models})
}

func handleRemoveModel(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("id")
	if modelID == "" {
		writeError(w, http.StatusBadRequest, "model id is required")
		return
	}

	cfg, err := config.Load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated := config.RemoveModel(cfg, modelID)
	if err := config.Save(updated); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"models": updated.Models})
}

type discussRequest struct {
	Prompt       string `json:"prompt"`
	CodebasePath string `json:"codebase_path"`
	Rounds       int    `json:"rounds"`
}

type clientMessage struct {
	Action string `json:"action"`
}

func handleDiscuss(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		http.Error(w, "websocket accept failed", http.StatusInternalServerError)
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(1024 * 1024)

	ctx := r.Context()

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				conn.Ping(ctx)
			}
		}
	}()

	var req discussRequest
	if err := wsjson.Read(ctx, conn, &req); err != nil {
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		wsjson.Write(ctx, conn, orchestrator.Event{Type: "error", Content: "prompt is required"})
		return
	}
	if strings.TrimSpace(req.CodebasePath) == "" {
		wsjson.Write(ctx, conn, orchestrator.Event{Type: "error", Content: "codebase_path is required"})
		return
	}

	cfg, err := config.Load()
	if err != nil {
		wsjson.Write(ctx, conn, orchestrator.Event{Type: "error", Content: "failed to load config: " + err.Error()})
		return
	}
	if cfg.APIKey == "" {
		wsjson.Write(ctx, conn, orchestrator.Event{Type: "error", Content: "API key not configured"})
		return
	}
	if len(cfg.Models) == 0 {
		wsjson.Write(ctx, conn, orchestrator.Event{Type: "error", Content: "no models configured"})
		return
	}

	client := openrouter.NewClient(cfg.APIKey)
	discCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var writeMu sync.Mutex
	broadcast := func(event orchestrator.Event) {
		writeMu.Lock()
		defer writeMu.Unlock()
		_ = wsjson.Write(discCtx, conn, event)
	}

	go func() {
		for {
			var msg clientMessage
			if err := wsjson.Read(ctx, conn, &msg); err != nil {
				return
			}
			if msg.Action == "stop" {
				cancel()
				return
			}
		}
	}()

	discID := fmt.Sprintf("disc_%d", time.Now().UnixMilli())
	disc := orchestrator.NewDiscussion(discID, req.Prompt, req.CodebasePath, cfg.Models, req.Rounds)
	orchestrator.Run(discCtx, disc, client, broadcast)

	conn.Close(websocket.StatusNormalClosure, "discussion complete")
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
