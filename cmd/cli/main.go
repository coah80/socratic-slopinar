package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/coah80/socratic-slopinar/internal/config"
	"github.com/coah80/socratic-slopinar/internal/openrouter"
	"github.com/coah80/socratic-slopinar/internal/orchestrator"
)

const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"
)

var modelColors = []string{magenta, blue, green, yellow, cyan, red, white}

func colorFor(idx int) string {
	return modelColors[idx%len(modelColors)]
}

func main() {
	prompt := flag.String("prompt", "", "discussion prompt (required)")
	codebase := flag.String("codebase", ".", "codebase path")
	rounds := flag.Int("rounds", 3, "number of discussion rounds")
	flag.Parse()

	if *prompt == "" {
		fmt.Fprintf(os.Stderr, "usage: socratic-cli -prompt \"your topic\" [-codebase /path] [-rounds N]\n")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if cfg.APIKey == "" && len(cfg.ProviderKeys) == 0 {
		log.Fatal("no API keys configured. run the web UI first to set up keys.")
	}
	if len(cfg.Models) == 0 {
		log.Fatal("no models configured. run the web UI first to add models.")
	}

	providerKeys := config.BuildProviderKeys(cfg)
	client := openrouter.NewClient(providerKeys)

	nameMap := make(map[string]int)
	for i, m := range cfg.Models {
		nameMap[m] = i
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		fmt.Printf("\n%s%s--- stopped ---%s\n", bold, red, reset)
		cancel()
	}()

	mutes := orchestrator.NewMuteSet()
	pins := orchestrator.NewPinSet()
	injector := orchestrator.NewInjector()

	fmt.Printf("%s%s╔══════════════════════════════════════════╗%s\n", bold, magenta, reset)
	fmt.Printf("%s%s║         SOCRATIC SLOPINAR (CLI)          ║%s\n", bold, magenta, reset)
	fmt.Printf("%s%s╚══════════════════════════════════════════╝%s\n", bold, magenta, reset)
	fmt.Printf("\n%sPrompt:%s %s\n", dim, reset, *prompt)
	fmt.Printf("%sCodebase:%s %s\n", dim, reset, *codebase)
	fmt.Printf("%sModels:%s %d\n", dim, reset, len(cfg.Models))
	fmt.Printf("%sRounds:%s %d\n\n", dim, reset, *rounds)

	broadcast := func(e orchestrator.Event) {
		idx := 0
		if i, ok := nameMap[e.ModelID]; ok {
			idx = i
		}
		clr := colorFor(idx)
		dn := e.DisplayName
		if dn == "" {
			dn = e.ModelID
		}

		switch e.Type {
		case "message":
			content := fmt.Sprintf("%v", e.Content)
			if dn == "God" {
				fmt.Printf("\n%s%s⚡ God:%s %s\n", bold, yellow, reset, content)
			} else {
				fmt.Printf("\n%s%s● %s:%s %s\n", bold, clr, dn, reset, content)
			}

		case "token":
			fmt.Printf("%v", e.Content)

		case "tool_call":
			if m, ok := e.Content.(map[string]string); ok {
				fmt.Printf("  %s🔧 %s → %s(%s)%s\n", dim, dn, m["name"], truncate(m["arguments"], 60), reset)
			}

		case "tool_result":
			if m, ok := e.Content.(map[string]string); ok {
				fmt.Printf("  %s   ↳ %s%s\n", dim, truncate(m["result"], 80), reset)
			}

		case "notes_update":
			content := fmt.Sprintf("%v", e.Content)
			lines := strings.Count(content, "\n") + 1
			fmt.Printf("  %s📝 Notes updated by %s (%d lines)%s\n", dim, dn, lines, reset)

		case "pin":
			fmt.Printf("  %s%s📌 %s pinned: %v%s\n", bold, yellow, dn, e.Content, reset)

		case "error":
			fmt.Printf("  %s%s✗ %s: %v%s\n", bold, red, dn, e.Content, reset)

		case "status":
			s := fmt.Sprintf("%v", e.Content)
			if strings.Contains(s, "round") {
				fmt.Printf("\n%s%s═══ %s ═══%s\n", bold, magenta, s, reset)
			} else if s == "done" {
				fmt.Printf("\n%s%s═══ Discussion Complete ═══%s\n", bold, green, reset)
			} else if strings.Contains(s, "consensus") {
				fmt.Printf("\n%s%s★ %s%s\n", bold, yellow, s, reset)
			} else if strings.Contains(s, "thinking") && dn != "" {
				fmt.Printf("  %s%s is thinking...%s\n", dim, dn, reset)
			}

		case "execution_prompt":
			content := fmt.Sprintf("%v", e.Content)
			fmt.Printf("\n%s%s╔══════════════════════════════════════════╗%s\n", bold, green, reset)
			fmt.Printf("%s%s║          EXECUTION PROMPT                ║%s\n", bold, green, reset)
			fmt.Printf("%s%s╚══════════════════════════════════════════╝%s\n", bold, green, reset)
			fmt.Printf("%s\n", content)
		}
	}

	discID := fmt.Sprintf("cli_%d", os.Getpid())
	disc := orchestrator.NewDiscussion(discID, *prompt, *codebase, cfg.Models, *rounds)
	result := orchestrator.Run(ctx, disc, client, broadcast, mutes, pins, injector)

	record := orchestrator.BuildRecord(disc, result)
	if err := config.SaveDiscussion(record); err != nil {
		log.Printf("failed to save history: %v", err)
	}

	fmt.Printf("\n%sDiscussion saved to history.%s\n", dim, reset)
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
