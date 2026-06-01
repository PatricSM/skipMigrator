package migrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPkgToAlias(t *testing.T) {
	cases := []struct{ in, want string }{
		{"tailwindcss-animate", "tailwindcssAnimatePlugin"},
		{"foo", "fooPlugin"},
		{"@scope/pkg", "scopePkgPlugin"},
		{"foo-bar-baz", "fooBarBazPlugin"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := pkgToAlias(tc.in)
			if got != tc.want {
				t.Errorf("pkgToAlias(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestConvertTailwindRequireToImport replaces CommonJS require() with ESM
// imports at the top of the file, leaving plugin references using the alias.
func TestConvertTailwindRequireToImport(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tailwind.config.ts")
	original := `import type { Config } from 'tailwindcss'
export default {
  plugins: [require("tailwindcss-animate")],
} satisfies Config
`
	mustWrite(t, path, original)

	if err := convertTailwindRequireToImport(path); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	out := string(got)

	if strings.Contains(out, `require("tailwindcss-animate")`) {
		t.Error("require() not replaced")
	}
	if !strings.Contains(out, `import tailwindcssAnimatePlugin from "tailwindcss-animate"`) {
		t.Error("ESM import not inserted")
	}
	if !strings.Contains(out, `plugins: [tailwindcssAnimatePlugin]`) {
		t.Error("plugin call not aliased")
	}
}

// TestRelaxOxlint appends the soft rules right after the strict marker.
func TestRelaxOxlint(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".oxlintrc.json")
	mustWrite(t, path, `{
  "rules": {
    "react-hooks/rules-of-hooks": "error",
    "no-debugger": "error"
  }
}`)
	if err := relaxOxlint(path); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	out := string(got)
	if !strings.Contains(out, `"no-empty": "warn"`) {
		t.Error("no-empty soften not injected")
	}
	if !strings.Contains(out, `"react-hooks/rules-of-hooks": "error"`) {
		t.Error("strict marker line was removed")
	}
}
