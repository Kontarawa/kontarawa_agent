package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	initEnv()

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "doctor":
		cmdDoctor(os.Args[2:])
	case "ask":
		cmdAsk(os.Args[2:])
	case "learn":
		cmdLearn(os.Args[2:])
	case "add-note":
		cmdAddNote(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `kontarawa - local agent with memory

Usage:
  kontarawa doctor
  kontarawa ask "question" [--model ...] [--k 6]
  kontarawa learn --prompt "..." --bad "..." --good "..." [--why "..."] [--tags a,b]
  kontarawa add-note --title "..." --text "..." [--tags a,b]

Env:
  OLLAMA_HOST   default: http://localhost:11434
  MEMORY_DIR    default: memory

`)
}

func mustNewFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	return fs
}

func ollamaHost() string {
	if v := strings.TrimSpace(os.Getenv("OLLAMA_HOST")); v != "" {
		return v
	}
	return "http://localhost:11434"
}

func memoryRoot() string {
	if v := strings.TrimSpace(os.Getenv("MEMORY_DIR")); v != "" {
		return v
	}
	return "memory"
}

