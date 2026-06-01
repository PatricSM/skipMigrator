// Package worker runs migration jobs picked up from Postgres LISTEN/NOTIFY.
package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PatricSM/skip-migrator/backend/internal/db"
	"github.com/PatricSM/skip-migrator/backend/internal/storage"
	"github.com/PatricSM/skip-migrator/migrator"
)

const (
	notifyChannel  = "migrations_queue"
	sourceBucket   = "source-zips"
	outputBucket   = "output-zips"
	signedURLTTLs  = 60 * 60 * 24 * 7 // 7 days
)

// Pool wraps the worker concurrency model.
type Pool struct {
	db      *pgxpool.Pool
	store   *storage.SupabaseStorage
	workers int
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

func New(pool *pgxpool.Pool, store *storage.SupabaseStorage, workers int) *Pool {
	if workers < 1 {
		workers = 1
	}
	return &Pool{db: pool, store: store, workers: workers, stopCh: make(chan struct{})}
}

// Start launches N goroutines listening on the NOTIFY channel.
// Also a fallback ticker picks up any orphan queued jobs every 30s.
func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.runWorker(ctx, i)
	}
}

// Stop signals all workers to drain and waits.
func (p *Pool) Stop() {
	close(p.stopCh)
	p.wg.Wait()
}

func (p *Pool) runWorker(ctx context.Context, id int) {
	defer p.wg.Done()
	log.Printf("[worker %d] started", id)

	listener := p.newListener(ctx)
	defer listener.close()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		default:
		}

		// Try to pick up a job; if none, wait for a notification or tick.
		m, ok, err := db.FetchAndLockNext(ctx, p.db)
		if err != nil {
			log.Printf("[worker %d] fetch error: %v", id, err)
			time.Sleep(5 * time.Second)
			continue
		}
		if !ok {
			listener.waitForNotification(ctx, 30*time.Second)
			continue
		}
		log.Printf("[worker %d] running migration %s for user %s", id, m.ID, m.UserID)
		p.runJob(ctx, m)
	}
}

func (p *Pool) runJob(ctx context.Context, m db.Migration) {
	// 1. Download source ZIP
	srcBytes, err := p.store.DownloadObject(sourceBucket, m.SourceZipPath)
	if err != nil {
		p.fail(ctx, m.ID, "downloading source", err, "")
		return
	}

	// 2. Persist source ZIP to /tmp, extract, prepare output dir
	work, err := os.MkdirTemp("", "migration-"+m.ID+"-*")
	if err != nil {
		p.fail(ctx, m.ID, "creating workdir", err, "")
		return
	}
	defer os.RemoveAll(work)

	srcZipPath := fmt.Sprintf("%s/source.zip", work)
	if err := os.WriteFile(srcZipPath, srcBytes, 0o600); err != nil {
		p.fail(ctx, m.ID, "writing source zip", err, "")
		return
	}
	srcDir := fmt.Sprintf("%s/source", work)
	if _, err := migrator.UnzipToDir(srcZipPath, srcDir); err != nil {
		p.fail(ctx, m.ID, "unzipping source", err, "")
		return
	}
	outDir := fmt.Sprintf("%s/output", work)

	// 3. Run migration
	opts := migrator.Options{
		SourceDir:    srcDir,
		OutputDir:    outDir,
		PixelPerfect: m.PixelPerfect,
		Validate:     m.Validate,
		LogFn: func(format string, args ...any) {
			log.Printf("[migration %s] "+format, append([]any{m.ID}, args...)...)
		},
	}
	result, err := migrator.Run(opts)
	buildLog := ""
	if result != nil {
		buildLog = result.BuildLog
	}
	if err != nil {
		p.fail(ctx, m.ID, "migration", err, buildLog)
		return
	}

	// 4. Zip output
	outZip := fmt.Sprintf("%s/output.zip", work)
	if err := migrator.ZipDir(outDir, outZip); err != nil {
		p.fail(ctx, m.ID, "zipping output", err, buildLog)
		return
	}
	outBytes, err := os.ReadFile(outZip)
	if err != nil {
		p.fail(ctx, m.ID, "reading output zip", err, buildLog)
		return
	}

	// 5. Upload output
	outPath := fmt.Sprintf("%s/%s/output.zip", m.UserID, m.ID)
	if _, err := p.store.UploadObject(outputBucket, outPath, outBytes, "application/zip"); err != nil {
		p.fail(ctx, m.ID, "uploading output", err, buildLog)
		return
	}

	// 6. Mark success
	if err := db.MarkSuccess(ctx, p.db, m.ID, outPath, buildLog); err != nil {
		log.Printf("[worker] mark success failed: %v", err)
	}
}

func (p *Pool) fail(ctx context.Context, id, stage string, err error, buildLog string) {
	msg := fmt.Sprintf("%s: %v", stage, err)
	log.Printf("[migration %s] FAILED %s", id, msg)
	if dbErr := db.MarkFailed(ctx, p.db, id, msg, buildLog); dbErr != nil {
		log.Printf("[worker] mark failed errored: %v", dbErr)
	}
}
