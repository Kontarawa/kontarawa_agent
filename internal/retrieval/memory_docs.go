package retrieval

import "kontarawa_agent/internal/memory"

// FromMemoryDocs maps memory store docs to retrieval.Doc (same DocID + text).
func FromMemoryDocs(docs []memory.Doc) []Doc {
	out := make([]Doc, 0, len(docs))
	for _, d := range docs {
		out = append(out, Doc{DocID: d.DocID, Text: d.Text})
	}
	return out
}
