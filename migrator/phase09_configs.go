package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// phase09BuildConfigs replaces Skip baseline's main.css with source's index.css,
// adopts source's tailwind.config (converting CommonJS require to ESM imports),
// and loosens Oxlint rules for imported code.
func phase09BuildConfigs(opts Options) (PhaseLog, error) {
	start := time.Now()
	log := PhaseLog{Phase: "09-build-configs"}

	// 1. CSS: prefer source's index.css over baseline's main.css
	srcCSS := filepath.Join(opts.SourceDir, "src", "index.css")
	dstCSS := filepath.Join(opts.OutputDir, "src", "main.css")
	if fileExists(srcCSS) {
		if err := copyFile(srcCSS, dstCSS); err != nil {
			return log, fmt.Errorf("copying CSS: %w", err)
		}
	}

	// 2. Tailwind: copy source's then convert require() → ESM import for TS strict mode compat
	srcTW := filepath.Join(opts.SourceDir, "tailwind.config.ts")
	dstTW := filepath.Join(opts.OutputDir, "tailwind.config.ts")
	if fileExists(srcTW) {
		if err := copyFile(srcTW, dstTW); err != nil {
			return log, fmt.Errorf("copying tailwind.config: %w", err)
		}
		if err := convertTailwindRequireToImport(dstTW); err != nil {
			return log, fmt.Errorf("patching tailwind.config: %w", err)
		}
	}

	// 3. Loosen Oxlint a bit for imported code
	oxPath := filepath.Join(opts.OutputDir, ".oxlintrc.json")
	if fileExists(oxPath) {
		if err := relaxOxlint(oxPath); err != nil {
			return log, fmt.Errorf("relaxing oxlint: %w", err)
		}
	}

	log.Status = "ok"
	log.Message = "CSS swapped, tailwind copied+patched, oxlint relaxed"
	log.Duration = time.Since(start).String()
	return log, nil
}

var tailwindRequireRe = regexp.MustCompile(`require\(["']([^"']+)["']\)`)

// convertTailwindRequireToImport rewrites CommonJS require() calls in tailwind.config.ts
// to ESM-style top-level imports.
//
// Example:
//
//	plugins: [require("tailwindcss-animate")]
//
// becomes:
//
//	import animatePlugin from "tailwindcss-animate"
//	...
//	plugins: [animatePlugin]
func convertTailwindRequireToImport(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	src := string(data)
	matches := tailwindRequireRe.FindAllStringSubmatch(src, -1)
	if len(matches) == 0 {
		return nil
	}

	importStmts := []string{}
	for _, m := range matches {
		pkg := m[1]
		alias := pkgToAlias(pkg)
		importStmts = append(importStmts, fmt.Sprintf("import %s from %q", alias, pkg))
		src = strings.ReplaceAll(src, m[0], alias)
	}

	// Insert imports after the first import line (or at top if none)
	firstImportEnd := strings.Index(src, "\n")
	if strings.HasPrefix(src, "import ") && firstImportEnd > 0 {
		insertion := "\n" + strings.Join(importStmts, "\n")
		src = src[:firstImportEnd] + insertion + src[firstImportEnd:]
	} else {
		src = strings.Join(importStmts, "\n") + "\n" + src
	}

	return os.WriteFile(path, []byte(src), 0o644)
}

// pkgToAlias converts a package name to a valid JS identifier alias.
// Example: "tailwindcss-animate" -> "tailwindcssAnimatePlugin"
func pkgToAlias(pkg string) string {
	parts := strings.FieldsFunc(pkg, func(r rune) bool {
		return r == '-' || r == '/' || r == '@'
	})
	var sb strings.Builder
	for i, p := range parts {
		if i == 0 {
			sb.WriteString(p)
		} else if p != "" {
			sb.WriteString(strings.ToUpper(p[:1]) + p[1:])
		}
	}
	sb.WriteString("Plugin")
	return sb.String()
}

// relaxOxlint adds permissive rule overrides commonly triggered by Lovable code.
func relaxOxlint(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	marker := `"react-hooks/rules-of-hooks": "error",`
	if i := indexOf(content, marker); i >= 0 {
		insert := `"react-hooks/rules-of-hooks": "error",
        "no-empty": "warn",
        "no-useless-escape": "warn",
        "no-fallthrough": "warn",
        "no-case-declarations": "warn",
        "no-useless-catch": "warn",
        "no-constant-condition": "warn",
        "no-prototype-builtins": "warn",
        "no-unused-expressions": "warn",
        "@typescript-eslint/no-empty-object-type": "off",
        "@typescript-eslint/no-require-imports": "warn",`
		content = content[:i] + insert + content[i+len(marker):]
		return os.WriteFile(path, []byte(content), 0o644)
	}
	return nil
}
