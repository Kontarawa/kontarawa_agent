package memory

import "testing"

func TestSlugify_NotEmpty(t *testing.T) {
	if got := Slugify("Hello world"); got == "" {
		t.Fatalf("expected non-empty slug")
	}
}

func TestSplitTags_DedupTrim(t *testing.T) {
	got := SplitTags(" a, b,  a , ,c ")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("unexpected tags: %#v", got)
	}
}

