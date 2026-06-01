package migrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractBetween(t *testing.T) {
	cases := []struct {
		name, src, start, end, want string
		ok                          bool
	}{
		{"simple", "<a>hi</a>", "<a>", "</a>", "hi", true},
		{"start missing", "<b>hi</b>", "<a>", "</a>", "", false},
		{"end missing", "<a>hi", "<a>", "</a>", "", false},
		{"multiline", "<head>\n  <title>x</title>\n</head>", "<head>", "</head>", "\n  <title>x</title>\n", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := extractBetween(tc.src, tc.start, tc.end)
			if ok != tc.ok || got != tc.want {
				t.Errorf("got (%q, %v), want (%q, %v)", got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestReplaceBetween(t *testing.T) {
	got := replaceBetween("<head>OLD</head>", "<head>", "</head>", "NEW")
	if got != "<head>NEW</head>" {
		t.Errorf("got %q, want <head>NEW</head>", got)
	}
}

// TestMergeIndexHTML asserts the merge keeps the destination <body>
// (with @skip-protected scripts) but takes the source <head> meta.
func TestMergeIndexHTML(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.html")
	dst := filepath.Join(dir, "dst.html")

	mustWrite(t, src, `<html><head><title>SRC</title><meta name="x" /></head><body>SRC-BODY</body></html>`)
	mustWrite(t, dst, `<html><head><title>DST</title></head><body><!-- @skip-protected -->
<script src="https://goskip.dev/skip.js"></script></body></html>`)

	merged, err := mergeIndexHTML(src, dst)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(merged, "<title>SRC</title>") {
		t.Error("merged should contain SRC head")
	}
	if !strings.Contains(merged, "@skip-protected") {
		t.Error("merged should preserve dst's @skip-protected body")
	}
	if strings.Contains(merged, "SRC-BODY") {
		t.Error("merged should NOT contain src body")
	}
}

func mustExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected %s to exist: %v", path, err)
	}
}
