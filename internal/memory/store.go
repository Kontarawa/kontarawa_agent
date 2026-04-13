package memory

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Store struct {
	Root string
}

func New(root string) Store {
	root = strings.TrimSpace(root)
	if root == "" {
		root = "memory"
	}
	return Store{Root: root}
}

func (s Store) ProfilePath() string  { return filepath.Join(s.Root, "profile.md") }
func (s Store) KnowledgeDir() string { return filepath.Join(s.Root, "knowledge") }
func (s Store) LessonsDir() string   { return filepath.Join(s.Root, "lessons") }

func (s Store) EnsureDirs() {
	_ = os.MkdirAll(s.KnowledgeDir(), 0o755)
	_ = os.MkdirAll(s.LessonsDir(), 0o755)
	// Profile is optional; create a minimal template only if missing.
	if _, err := os.Stat(s.ProfilePath()); err != nil {
		_ = os.WriteFile(s.ProfilePath(), []byte("# Profile\n\n"), 0o644)
	}
}

func (s Store) ReadProfile() string {
	b, err := os.ReadFile(s.ProfilePath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

type Doc struct {
	DocID string
	Path  string
	Text  string
}

func (s Store) LoadDocs() []Doc {
	var out []Doc
	roots := []string{s.KnowledgeDir(), s.LessonsDir()}
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil || d == nil || d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
				return nil
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			txt := strings.TrimSpace(string(b))
			if txt == "" {
				return nil
			}
			out = append(out, Doc{
				DocID: filepath.ToSlash(path),
				Path:  path,
				Text:  txt,
			})
			return nil
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DocID < out[j].DocID })
	return out
}

func (s Store) SaveLesson(prompt, bad, good, why string, tags []string) (string, error) {
	ts := time.Now().UTC().Format("20060102-150405")
	path := filepath.Join(s.LessonsDir(), ts+".md")
	meta := map[string]any{
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"tags":       tags,
	}
	var b strings.Builder
	b.WriteString("# Lesson " + ts + "\n\n")
	b.WriteString("## Meta\n```json\n")
	j, _ := json.MarshalIndent(meta, "", "  ")
	b.Write(j)
	b.WriteString("\n```\n\n")
	b.WriteString("## Prompt\n" + strings.TrimSpace(prompt) + "\n\n")
	b.WriteString("## Bad answer\n" + strings.TrimSpace(bad) + "\n\n")
	b.WriteString("## Good answer\n" + strings.TrimSpace(good) + "\n\n")
	if strings.TrimSpace(why) != "" {
		b.WriteString("## Why\n" + strings.TrimSpace(why) + "\n\n")
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (s Store) SaveKnowledgeNote(title, text string, tags []string) (string, error) {
	ts := time.Now().UTC().Format("20060102-150405")
	slug := Slugify(title)
	name := ts + "-" + slug + ".md"
	path := filepath.Join(s.KnowledgeDir(), name)
	meta := map[string]any{
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"title":      title,
		"tags":       tags,
	}
	var b strings.Builder
	b.WriteString("# " + title + "\n\n")
	b.WriteString("## Meta\n```json\n")
	j, _ := json.MarshalIndent(meta, "", "  ")
	b.Write(j)
	b.WriteString("\n```\n\n")
	b.WriteString(strings.TrimSpace(text))
	b.WriteString("\n")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func SplitTags(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	return out
}

var nonWord = regexp.MustCompile(`[^\p{L}\p{N}]+`)

func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonWord.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		h := sha1.Sum([]byte(time.Now().UTC().String()))
		return hex.EncodeToString(h[:])[:10]
	}
	if len(s) > 60 {
		s = s[:60]
		s = strings.Trim(s, "-")
	}
	return s
}

