package migrator

import (
	"path/filepath"
	"testing"
)

func TestExtractSupabaseCreds(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "src", "integrations", "supabase", "client.ts"), `
const SUPABASE_URL = "https://abcd1234efgh5678ijkl.supabase.co";
const SUPABASE_PUBLISHABLE_KEY = "eyJhbGciOiJIUzI1NiI.payload.sig";
`)

	url, key, err := extractSupabaseCreds(dir)
	if err != nil {
		t.Fatalf("expected creds, got err: %v", err)
	}
	if url != "https://abcd1234efgh5678ijkl.supabase.co" {
		t.Errorf("url = %q", url)
	}
	if key != "eyJhbGciOiJIUzI1NiI.payload.sig" {
		t.Errorf("key = %q", key)
	}
}

func TestExtractSupabaseCreds_missing(t *testing.T) {
	_, _, err := extractSupabaseCreds(t.TempDir())
	if err == nil {
		t.Error("expected error when no client.ts present")
	}
}

func TestMaskKey(t *testing.T) {
	if got := maskKey(""); got != "<empty>" {
		t.Errorf("empty key not masked: %q", got)
	}
	if got := maskKey("abcdef"); got != "***" {
		t.Errorf("short key should be ***, got %q", got)
	}
	if got := maskKey("aaaaaa.payload.bbbbbb"); !contains(got, "...") {
		t.Errorf("long key should preserve start/end with ..., got %q", got)
	}
}

func contains(s, sub string) bool {
	return indexOf(s, sub) >= 0
}
