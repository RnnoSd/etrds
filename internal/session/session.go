package session

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"

	"github.com/RnnoSd/etrds/internal/cache"
)

var (
	mu       sync.Mutex
	sessions = map[string]*cache.CachedSession{}
)

type aliveSession struct {
	ID           string
	WMode        string
	WPath        string             // This is the local DuckDB connection for storage
	RConnections map[string]*sql.DB // These are local DuckDB connection for transformations, there could be multiple since we can have multiple reading connection
	Fetching     map[string]cache.FetchingStruct
	Consults     map[string]*Consult
}

type Consult struct {
	Name    string
	Content []byte
	State   cache.State
}

func SetUpSessionCached(cs *cache.CachedSession) (*aliveSession, error) {
	consults := make(map[string]*Consult)
	rconnections := make(map[string]*sql.DB)

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

	for name, path := range cs.ReadPaths {
		dbRef, err := sql.Open("duckdb", path+"?access_mode=read_only")
		if err != nil {
			return nil, err
		}
		rconnections[name] = dbRef
	}

	return &aliveSession{
		ID:           cs.ID,
		WMode:        "READING",
		WPath:        cs.WritingPath,
		RConnections: rconnections,
		Fetching:     cs.Fetching,
		Consults:     consults,
	}, nil
}

func (cons *Consult) Queued() {
	cons.State = cache.Queued
}
