package worker

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// notifyListener wraps a dedicated pgx connection for LISTEN/NOTIFY.
type notifyListener struct {
	conn *pgxpool.Conn
}

func (p *Pool) newListener(ctx context.Context) *notifyListener {
	conn, err := p.db.Acquire(ctx)
	if err != nil {
		log.Printf("[listener] acquire failed: %v", err)
		return &notifyListener{}
	}
	_, err = conn.Exec(ctx, "LISTEN "+notifyChannel)
	if err != nil {
		log.Printf("[listener] LISTEN failed: %v", err)
		conn.Release()
		return &notifyListener{}
	}
	return &notifyListener{conn: conn}
}

// waitForNotification blocks until a NOTIFY arrives or the timeout elapses.
// Returns silently on cancel / error — caller loops regardless.
func (l *notifyListener) waitForNotification(ctx context.Context, timeout time.Duration) {
	if l.conn == nil {
		select {
		case <-ctx.Done():
		case <-time.After(timeout):
		}
		return
	}
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	// We ignore the error: caller loops back to FetchAndLockNext regardless.
	_, _ = l.conn.Conn().WaitForNotification(waitCtx)
}

func (l *notifyListener) close() {
	if l.conn != nil {
		_, _ = l.conn.Exec(context.Background(), "UNLISTEN "+notifyChannel)
		l.conn.Release()
	}
}
