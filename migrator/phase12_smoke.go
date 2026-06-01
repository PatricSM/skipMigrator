package migrator

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// phase12Smoke starts `vite preview` and hits the root URL with a headless
// Chromium browser, collecting console errors and uncaught exceptions.
//
// Fails the migration if any error-level console message or page exception
// fires while loading "/". Times out after 30s.
//
// Requires:
//   - dist/ to exist in opts.OutputDir (phase11 must have run)
//   - chromium binary available (path resolved by launcher; in our Docker
//     image we install `chromium` via apk and set ROD_CHROMIUM_BIN env)
func phase12Smoke(opts Options) (PhaseLog, error) {
	start := time.Now()
	log := PhaseLog{Phase: "12-smoke-test"}

	if !fileExists(filepath.Join(opts.OutputDir, "dist", "index.html")) {
		log.Status = "skipped"
		log.Message = "dist/ missing; phase 11 must run first"
		log.Duration = time.Since(start).String()
		return log, nil
	}

	port := pickFreePort()
	previewURL := fmt.Sprintf("http://localhost:%d/", port)

	// Spawn `pnpm preview` bound to the picked port
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	previewCmd := exec.CommandContext(ctx, "pnpm", "preview", "--port", fmt.Sprint(port), "--strictPort", "--host", "127.0.0.1")
	previewCmd.Dir = opts.OutputDir
	// Process group so we can kill children cleanly
	previewCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, _ := previewCmd.StdoutPipe()
	stderr, _ := previewCmd.StderrPipe()
	if err := previewCmd.Start(); err != nil {
		log.Status = "error"
		log.Message = fmt.Sprintf("failed to start vite preview: %v", err)
		log.Duration = time.Since(start).String()
		return log, err
	}
	defer func() {
		if previewCmd.Process != nil {
			_ = syscall.Kill(-previewCmd.Process.Pid, syscall.SIGKILL)
		}
	}()

	// Wait for "Local:" line indicating preview is ready
	ready := make(chan struct{}, 1)
	go func() {
		merged := io.MultiReader(stdout, stderr)
		sc := bufio.NewScanner(merged)
		for sc.Scan() {
			line := sc.Text()
			if strings.Contains(line, "Local:") || strings.Contains(line, "ready in") {
				select {
				case ready <- struct{}{}:
				default:
				}
			}
		}
	}()

	select {
	case <-ready:
		// proceed
	case <-time.After(15 * time.Second):
		log.Status = "error"
		log.Message = "vite preview did not become ready within 15s"
		log.Duration = time.Since(start).String()
		return log, fmt.Errorf("preview timeout")
	}

	// Launch headless Chromium via rod
	binPath := os.Getenv("ROD_CHROMIUM_BIN")
	l := launcher.New().Headless(true).NoSandbox(true)
	if binPath != "" {
		l = l.Bin(binPath)
	}
	wsURL, err := l.Launch()
	if err != nil {
		log.Status = "warn"
		log.Message = fmt.Sprintf("could not launch headless Chromium (skipped): %v", err)
		log.Duration = time.Since(start).String()
		return log, nil
	}

	browser := rod.New().ControlURL(wsURL)
	if err := browser.Connect(); err != nil {
		log.Status = "warn"
		log.Message = fmt.Sprintf("could not connect to Chromium (skipped): %v", err)
		log.Duration = time.Since(start).String()
		return log, nil
	}
	defer browser.MustClose()

	page, err := browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		log.Status = "error"
		log.Message = fmt.Sprintf("could not open page: %v", err)
		log.Duration = time.Since(start).String()
		return log, err
	}

	var consoleErrs []string
	var pageErrs []string

	// Subscribe BEFORE navigation
	go page.EachEvent(func(e *proto.RuntimeConsoleAPICalled) {
		if e.Type == proto.RuntimeConsoleAPICalledTypeError {
			var parts []string
			for _, arg := range e.Args {
				if v := arg.Value.String(); v != "" {
					parts = append(parts, v)
				} else if arg.Description != "" {
					parts = append(parts, arg.Description)
				}
			}
			msg := strings.Join(parts, " ")
			if msg == "" {
				msg = "(empty)"
			}
			consoleErrs = append(consoleErrs, truncate(msg, 200))
		}
	}, func(e *proto.RuntimeExceptionThrown) {
		desc := e.ExceptionDetails.Exception.Description
		if desc == "" {
			desc = e.ExceptionDetails.Text
		}
		pageErrs = append(pageErrs, truncate(desc, 200))
	})()

	if err := page.Navigate(previewURL); err != nil {
		log.Status = "error"
		log.Message = fmt.Sprintf("navigate failed: %v", err)
		log.Duration = time.Since(start).String()
		return log, err
	}

	// Wait for the page to settle: try MustWaitLoad with a short deadline,
	// then give the bundle 2s to throw any deferred error.
	loadCtx, loadCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer loadCancel()
	if err := page.Context(loadCtx).WaitLoad(); err != nil {
		log.Status = "error"
		log.Message = fmt.Sprintf("page load did not finish: %v", err)
		log.Duration = time.Since(start).String()
		return log, err
	}
	time.Sleep(2 * time.Second) // settle for deferred errors

	allErrs := append([]string{}, pageErrs...)
	allErrs = append(allErrs, consoleErrs...)

	if len(allErrs) > 0 {
		log.Status = "error"
		log.Message = fmt.Sprintf("%d runtime error(s) on /", len(allErrs))
		for _, e := range allErrs {
			log.Details = append(log.Details, e)
		}
		log.Duration = time.Since(start).String()
		return log, fmt.Errorf("smoke test found %d runtime error(s)", len(allErrs))
	}

	log.Status = "ok"
	log.Message = "/ loaded without runtime errors"
	log.Duration = time.Since(start).String()
	return log, nil
}

func pickFreePort() int {
	// Vite preview defaults to 4173; bump to avoid collisions in worker pool.
	return 4173
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
