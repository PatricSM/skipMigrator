// Command migrate-cli runs the Lovable → Skip migration end-to-end.
//
// Usage:
//
//	migrate-cli -src <source-zip-or-dir> -out <output-zip-or-dir> [-validate] [-pixel-perfect]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PatricSM/skip-migrator/migrator"
)

func main() {
	var (
		src          = flag.String("src", "", "source: path to a Lovable project ZIP file or extracted directory")
		out          = flag.String("out", "", "output: path where the migrated Skip project ZIP or dir will be written")
		validate     = flag.Bool("validate", false, "run pnpm install + tsc + build to verify the output")
		pixelPerfect = flag.Bool("pixel-perfect", false, "override components/ui with source versions (adapts calendar v8→v9)")
		supaURL      = flag.String("supabase-url", "", "override Supabase URL (otherwise extracted from source)")
		supaKey      = flag.String("supabase-anon-key", "", "override Supabase anon key (otherwise extracted from source)")
	)
	flag.Parse()

	if *src == "" || *out == "" {
		flag.Usage()
		os.Exit(2)
	}

	// Resolve source: if a ZIP file, extract to a temp dir
	srcDir := *src
	if isZip(*src) {
		tmp, err := os.MkdirTemp("", "skip-migrator-src-*")
		if err != nil {
			die("creating temp dir for source: %v", err)
		}
		defer os.RemoveAll(tmp)
		if _, err := migrator.UnzipToDir(*src, tmp); err != nil {
			die("extracting source zip: %v", err)
		}
		srcDir = tmp
		fmt.Fprintf(os.Stderr, "Source extracted to %s\n", srcDir)
	}

	// Resolve output: if ZIP path, run in temp dir then zip up
	outIsZip := isZip(*out)
	outDir := *out
	if outIsZip {
		tmp, err := os.MkdirTemp("", "skip-migrator-out-*")
		if err != nil {
			die("creating temp dir for output: %v", err)
		}
		defer os.RemoveAll(tmp)
		outDir = tmp
	}

	opts := migrator.Options{
		SourceDir:       srcDir,
		OutputDir:       outDir,
		SupabaseURL:     *supaURL,
		SupabaseAnonKey: *supaKey,
		PixelPerfect:    *pixelPerfect,
		Validate:        *validate,
	}

	result, err := migrator.Run(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nMIGRATION FAILED: %v\n", err)
		if result != nil && result.BuildLog != "" {
			fmt.Fprintln(os.Stderr, "\n--- Build log ---")
			fmt.Fprintln(os.Stderr, result.BuildLog)
		}
		os.Exit(1)
	}

	if outIsZip {
		if err := migrator.ZipDir(outDir, *out); err != nil {
			die("zipping output: %v", err)
		}
		abs, _ := filepath.Abs(*out)
		fmt.Fprintf(os.Stderr, "\n✓ Output ZIP written to %s\n", abs)
	} else {
		fmt.Fprintf(os.Stderr, "\n✓ Output written to %s\n", outDir)
	}

	fmt.Fprintln(os.Stderr, "\n--- Phase summary ---")
	for _, p := range result.PhaseLogs {
		fmt.Fprintf(os.Stderr, "  [%s] %-10s %s (%s)\n", p.Phase, p.Status, p.Message, p.Duration)
	}
}

func isZip(p string) bool {
	if len(p) < 4 {
		return false
	}
	return p[len(p)-4:] == ".zip"
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
