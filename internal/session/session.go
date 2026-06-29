package session

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/RnnoSd/etrds/internal/cache"
	duckdb "github.com/marcboeker/go-duckdb/v2"
)

var (
	mu       sync.Mutex
	sessions = map[string]*cache.CachedSession{}
)

type AliveSession struct {
	ID       string
	WMode    string
	WPath    string  // Path of the local DuckDB store (the read_write database)
	DB       *sql.DB // Single read_write connection to WPath; ReadPaths and Fetching are ATTACHed READ_ONLY on top of it
	Fetching map[string]cache.FetchingStruct
	Consults map[string]*Consult
}

type Consult struct {
	Name    string
	Content []byte
	State   cache.State
}

func SetUpSessionCached(cs *cache.CachedSession) (*AliveSession, error) {
	consults := make(map[string]*Consult)

	root := cache.ProjectRoot()

	for name, consult := range cs.Consults {
		content, err := os.ReadFile(filepath.Join(root, consult.Path))
		if err != nil {
			return nil, err
		}
		consults[name] = &Consult{
			Name:    name,
			Content: content,
			State:   consult.State,
		}
	}

	// One read_write connection to the local store. ReadPaths (other DuckDB
	// files) and Fetching (external engines) are ATTACHed READ_ONLY in the
	// connector init, so every pooled connection shares the same attachments.
	connector, err := duckdb.NewConnector(cs.WritingPath, attachReadOnly(cs))
	if err != nil {
		return nil, err
	}

	return &AliveSession{
		ID:       cs.ID,
		WMode:    "READWRITE",
		WPath:    cs.WritingPath,
		DB:       sql.OpenDB(connector),
		Fetching: cs.Fetching,
		Consults: consults,
	}, nil
}

// attachReadOnly builds the connection bootstrap that ATTACHes every read
// source READ_ONLY. It runs on each new pooled connection; ATTACH IF NOT EXISTS
// keeps that idempotent and LOAD is re-issued per connection as DuckDB requires.
func attachReadOnly(cs *cache.CachedSession) func(driver.ExecerContext) error {
	return func(execer driver.ExecerContext) error {
		ctx := context.Background()
		exec := func(query string) error {
			_, err := execer.ExecContext(ctx, query, nil)
			return err
		}

		// Other DuckDB files (default TYPE is duckdb).
		for name, path := range cs.ReadPaths {
			if err := exec(fmt.Sprintf("ATTACH IF NOT EXISTS '%s' AS %s (READ_ONLY)", path, name)); err != nil {
				return fmt.Errorf("attaching read path %q: %w", name, err)
			}
		}

		// External engines (mysql, postgres, ...), each needing its extension.
		loaded := make(map[string]bool)
		for name, f := range cs.Fetching {
			ext := strings.ToLower(f.Type)
			if !loaded[ext] {
				if err := exec("INSTALL " + ext); err != nil {
					return fmt.Errorf("installing %q extension: %w", ext, err)
				}
				if err := exec("LOAD " + ext); err != nil {
					return fmt.Errorf("loading %q extension: %w", ext, err)
				}
				loaded[ext] = true
			}
			if err := exec(fmt.Sprintf("ATTACH IF NOT EXISTS '%s' AS %s (TYPE %s, READ_ONLY)", f.DSN, name, ext)); err != nil {
				return fmt.Errorf("attaching source %q: %w", name, err)
			}
		}
		return nil
	}
}

// Close releases the session's database connection.
func (s *AliveSession) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}

func (cons *Consult) Queued() {
	cons.State = cache.Queued
}
