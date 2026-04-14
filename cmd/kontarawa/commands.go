package main

import (
	"fmt"
	"os"
	"strings"

	"kontarawa_agent/internal/memory"
	"kontarawa_agent/internal/ollama"
	"kontarawa_agent/internal/prompt"
	"kontarawa_agent/internal/retrieval"
)

const defaultModel = "qwen2.5-coder:7b-instruct"

func cmdDoctor(args []string) {
	fs := mustNewFlagSet("doctor")
	_ = fs.Parse(args)

	host := ollamaHost()
	fmt.Printf("ollama host: %s\n", host)

	ok := ollama.New(host).Ping()
	if ok {
		fmt.Println("ollama: OK")
	} else {
		fmt.Println("ollama: FAIL (is `ollama serve` running?)")
	}

	memory.New(memoryRoot()).EnsureDirs()
	fmt.Println("memory: OK")
}

func cmdAsk(args []string) {
	fs := mustNewFlagSet("ask")
	model := fs.String("model", defaultModel, "Ollama model name")
	k := fs.Int("k", 6, "How many memory docs to include")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "ask requires a question")
		os.Exit(2)
	}
	question := strings.Join(rest, " ")

	host := ollamaHost()
	store := memory.New(memoryRoot())
	profileText := store.ReadProfile()
	docs := retrieval.FromMemoryDocs(store.LoadDocs())
	retrieved := retrieval.Retrieve(question, docs, *k)

	system := prompt.BuildSystem(profileText, retrieved)
	answer, err := ollama.New(host).Chat(*model, system, question)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(retrieved) > 0 {
		fmt.Println("memory used:")
		for _, r := range retrieved {
			fmt.Printf("- %s (score=%.3f)\n", r.DocID, r.Score)
		}
		fmt.Println()
	}
	fmt.Println(answer)
}

func cmdLearn(args []string) {
	fs := mustNewFlagSet("learn")
	taskPrompt := fs.String("prompt", "", "Original prompt/task")
	bad := fs.String("bad", "", "Bad answer")
	good := fs.String("good", "", "Good answer")
	why := fs.String("why", "", "Why the good answer is better")
	tags := fs.String("tags", "", "Comma-separated tags")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if strings.TrimSpace(*taskPrompt) == "" || strings.TrimSpace(*bad) == "" || strings.TrimSpace(*good) == "" {
		fmt.Fprintln(os.Stderr, "learn requires --prompt, --bad, --good")
		os.Exit(2)
	}

	store := memory.New(memoryRoot())
	store.EnsureDirs()
	path, err := store.SaveLesson(*taskPrompt, *bad, *good, *why, memory.SplitTags(*tags))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("saved: %s\n", path)
}

func cmdAddNote(args []string) {
	fs := mustNewFlagSet("add-note")
	title := fs.String("title", "", "Note title")
	text := fs.String("text", "", "Note text (markdown)")
	tags := fs.String("tags", "", "Comma-separated tags")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if strings.TrimSpace(*title) == "" || strings.TrimSpace(*text) == "" {
		fmt.Fprintln(os.Stderr, "add-note requires --title and --text")
		os.Exit(2)
	}

	store := memory.New(memoryRoot())
	store.EnsureDirs()
	path, err := store.SaveKnowledgeNote(*title, *text, memory.SplitTags(*tags))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("saved: %s\n", path)
}

