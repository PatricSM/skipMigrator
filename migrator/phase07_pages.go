package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// phase07Pages replaces baseline pages/App.tsx with source's, and copies docs/ if present
// (for pages that import markdown via ?raw).
func phase07Pages(opts Options) (PhaseLog, error) {
	start := time.Now()
	log := PhaseLog{Phase: "07-pages"}

	// Remove baseline pages and replace with source
	pagesDst := filepath.Join(opts.OutputDir, "src", "pages")
	if err := os.RemoveAll(pagesDst); err != nil {
		return log, err
	}
	pagesSrc := filepath.Join(opts.SourceDir, "src", "pages")
	if !fileExists(pagesSrc) {
		return log, fmt.Errorf("source has no src/pages dir")
	}
	if err := os.MkdirAll(pagesDst, 0o755); err != nil {
		return log, err
	}
	if err := copyDir(pagesSrc, pagesDst, nil); err != nil {
		return log, fmt.Errorf("copying pages: %w", err)
	}

	// Replace App.tsx
	srcApp := filepath.Join(opts.SourceDir, "src", "App.tsx")
	if fileExists(srcApp) {
		if err := copyFile(srcApp, filepath.Join(opts.OutputDir, "src", "App.tsx")); err != nil {
			return log, fmt.Errorf("copying App.tsx: %w", err)
		}
	}

	// Keep baseline's main.tsx (it has @skip-protected marker)

	// Copy docs/ if present (DownloadDocs.tsx and similar do ?raw imports)
	srcDocs := filepath.Join(opts.SourceDir, "docs")
	if fileExists(srcDocs) {
		dst := filepath.Join(opts.OutputDir, "docs")
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return log, err
		}
		if err := copyDir(srcDocs, dst, nil); err != nil {
			return log, fmt.Errorf("copying docs: %w", err)
		}
	}

	// Copy index.html with source's SEO meta (preserve @skip-protected scripts)
	srcIndex := filepath.Join(opts.SourceDir, "index.html")
	dstIndex := filepath.Join(opts.OutputDir, "index.html")
	if fileExists(srcIndex) && fileExists(dstIndex) {
		if merged, err := mergeIndexHTML(srcIndex, dstIndex); err == nil {
			_ = os.WriteFile(dstIndex, []byte(merged), 0o644)
		}
	}

	log.Status = "ok"
	log.Message = "pages + App.tsx replaced; docs/ copied if present; index.html merged"
	log.Duration = time.Since(start).String()
	return log, nil
}

// mergeIndexHTML keeps Skip's @skip-protected scripts and replaces the head meta
// with the source's. Returns the merged HTML.
func mergeIndexHTML(srcPath, dstPath string) (string, error) {
	srcBytes, err := os.ReadFile(srcPath)
	if err != nil {
		return "", err
	}
	dstBytes, err := os.ReadFile(dstPath)
	if err != nil {
		return "", err
	}
	src := string(srcBytes)
	dst := string(dstBytes)
	srcHead, srcOK := extractBetween(src, "<head>", "</head>")
	if !srcOK {
		return dst, nil
	}
	srcBody, srcBodyOK := extractBetween(src, "<body>", "</body>")
	dstBody, dstBodyOK := extractBetween(dst, "<body>", "</body>")
	if !srcBodyOK || !dstBodyOK {
		return dst, nil
	}
	// Use Skip's body (has @skip-protected scripts)
	merged := replaceBetween(dst, "<head>", "</head>", srcHead)
	_ = srcBody
	merged = replaceBetween(merged, "<body>", "</body>", dstBody)
	return merged, nil
}

func extractBetween(s, start, end string) (string, bool) {
	i := indexOf(s, start)
	if i < 0 {
		return "", false
	}
	j := indexOf(s[i+len(start):], end)
	if j < 0 {
		return "", false
	}
	return s[i+len(start) : i+len(start)+j], true
}

func replaceBetween(s, start, end, newContent string) string {
	i := indexOf(s, start)
	if i < 0 {
		return s
	}
	j := indexOf(s[i+len(start):], end)
	if j < 0 {
		return s
	}
	return s[:i+len(start)] + newContent + s[i+len(start)+j:]
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
