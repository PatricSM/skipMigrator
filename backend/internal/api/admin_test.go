package api

import (
	"strings"
	"testing"
)

func TestGenerateTempPassword(t *testing.T) {
	const n = 16
	a := generateTempPassword(n)
	b := generateTempPassword(n)
	if len(a) != n || len(b) != n {
		t.Errorf("len mismatch: %d, %d", len(a), len(b))
	}
	if a == b {
		t.Errorf("passwords should not collide: %q == %q", a, b)
	}
	// URL-safe base64: only [A-Za-z0-9_-]
	for _, r := range a {
		ok := (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
		if !ok {
			t.Errorf("non-url-safe char %q in password %q", r, a)
		}
	}
}

func TestLooksLikeZip(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want bool
	}{
		{"empty", []byte{}, false},
		{"short", []byte("PK"), false},
		{"valid PKZIP", []byte{'P', 'K', 0x03, 0x04, 0x00}, true},
		{"wrong magic", []byte{'R', 'A', 0x03, 0x04}, false},
		{"text file", []byte("hello world"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := looksLikeZip(tc.data); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// Sanity check: random passwords from generateTempPassword always pass the
// "looksLikeZip" filter as `false` (just defensive — types differ but make
// sure no surprise overlap).
func TestPasswordIsNotMistakenForZip(t *testing.T) {
	for i := 0; i < 32; i++ {
		p := generateTempPassword(16)
		if strings.HasPrefix(p, "PK") {
			// Possible (rare) but would never be 4-byte PK\x03\x04 magic
			if looksLikeZip([]byte(p)) {
				t.Fatalf("password %q falsely looks like zip", p)
			}
		}
	}
}
