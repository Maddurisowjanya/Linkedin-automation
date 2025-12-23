package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"
	"github.com/sirupsen/logrus"
)

// Storage wraps a SQLite connection and exposes small helper methods for
// tracking actions. This is intentionally minimal and not a full ORM.
type Storage struct {
	db  *sql.DB
	log *logrus.Logger
}

func New(dsn string, log *logrus.Logger) (*Storage, error) {
	// Use the pure‑Go modernc.org/sqlite driver so this PoC works without CGO.
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	s := &Storage{db: db, log: log}
	if err := s.initSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) initSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS sent_requests (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	profile_url TEXT NOT NULL UNIQUE,
	sent_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	profile_url TEXT NOT NULL,
	message_type TEXT NOT NULL,
	sent_at TIMESTAMP NOT NULL
);
`
	_, err := s.db.Exec(schema)
	return err
}

// HasSentRequest returns true if a connection request has already been
// recorded for the given profile URL.
func (s *Storage) HasSentRequest(ctx context.Context, profileURL string) (bool, error) {
	row := s.db.QueryRowContext(ctx, `SELECT 1 FROM sent_requests WHERE profile_url = ?`, profileURL)
	var tmp int
	err := row.Scan(&tmp)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Storage) RecordRequest(ctx context.Context, profileURL string, when time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO sent_requests (profile_url, sent_at) VALUES (?, ?)`,
		profileURL, when.UTC(),
	)
	return err
}

// CountRequestsSince returns how many requests have been recorded since the
// given time. Used to enforce simple daily limits.
func (s *Storage) CountRequestsSince(ctx context.Context, since time.Time) (int, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM sent_requests WHERE sent_at >= ?`,
		since.UTC(),
	)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (s *Storage) RecordMessage(ctx context.Context, profileURL, msgType string, when time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO messages (profile_url, message_type, sent_at) VALUES (?, ?, ?)`,
		profileURL, msgType, when.UTC(),
	)
	return err
}

func (s *Storage) CountMessagesSince(ctx context.Context, msgType string, since time.Time) (int, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM messages WHERE message_type = ? AND sent_at >= ?`,
		msgType, since.UTC(),
	)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}



