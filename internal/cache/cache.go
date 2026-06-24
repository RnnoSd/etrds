package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const etrdsDirName = ".etrds"

var (
	ErrNoEtrdsDir      = errors.New("etrds project found")
	ErrSessionNotFound = errors.New("session not found in cache")
)

// CachedSession is the inert, serializable form of a session. It holds only the id, the consults, and the specs needed to REOPEN the
// connections when the session is promoted to an aliveSession.
type Cache struct {
	CachedSessions []*CachedSession `yaml:"cached_sessions"`
}

type CachedConsult struct {
	Name  string `yaml:"name"`
	Path  string `yaml:"path"`
	State State  `yaml:"state"`
}

type FetchingStruct struct {
	Type string `yaml:"type"`
	DSN  string `yaml:"dsn"`
}

type CachedSession struct {
	ID          string                    `yaml:"id"`
	WritingPath string                    `yaml:"writing_path"`
	ReadPaths   map[string]string         `yaml:"read_paths"`
	Fetching    map[string]FetchingStruct `yaml:"fetching"`
	Consults    map[string]CachedConsult  `yaml:"consult_paths"`
}

type State int

const (
	Queued State = iota + 1
	Running
	Fetched
	Stored
	Done
)

func findEtrdsDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, etrdsDirName)
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return candidate, nil
		}
		parent, _ := filepath.Abs(filepath.Dir(dir))
		if parent == dir {
			return "", ErrNoEtrdsDir
		}
		dir = parent
	}
}

func ProjectRoot() string {
	projectroot, _ := findEtrdsDir()

	return filepath.Dir(projectroot)
}

func ReadCachedSession(id string) (*CachedSession, error) {
	dir, err := findEtrdsDir()
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(dir + `/sessions/` + id + `.yaml`)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%q: %w", id, ErrSessionNotFound)
	}
	if err != nil {
		return nil, err
	}

	var cs CachedSession
	if err := json.Unmarshal(raw, &cs); err != nil {
		return nil, fmt.Errorf("corrupt cache for session %q: %w", id, err)
	}
	return &cs, nil
}

func (cs *CachedSession) save() error {
	dir, err := findEtrdsDir()
	if err != nil {
		return err
	}
	raw, err := json.MarshalIndent(cs, "", "  ")
	if err != nil {
		return err
	}
	writingPath := dir + `/sessions/` + cs.ID + `.yaml`
	return os.WriteFile(writingPath, raw, 0o644)
}
