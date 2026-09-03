// Package backup stores timestamped copies of JSON documents on disk.
package backup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var ErrNoBackups = errors.New("no backups found")

const timeLayout = "20060102T150405Z"

type Entry struct {
	Path string
	Time time.Time
}

type Store struct {
	dir string
	now func() time.Time
}

func NewStore(dir string) *Store {
	return &Store{dir: dir, now: time.Now}
}

func (s *Store) Save(id int, data []byte) (Entry, error) {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "    "); err != nil {
		return Entry{}, fmt.Errorf("invalid JSON: %w", err)
	}

	dir := s.subdir(id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Entry{}, err
	}

	at := s.now().UTC()
	path := filepath.Join(dir, at.Format(timeLayout)+".json")
	if err := os.WriteFile(path, pretty.Bytes(), 0o644); err != nil {
		return Entry{}, err
	}

	return Entry{Path: path, Time: at}, nil
}

func (s *Store) Load(id int) ([]byte, error) {
	entries, err := s.List(id)
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, ErrNoBackups
	}

	return os.ReadFile(entries[0].Path)
}

func (s *Store) List(id int) ([]Entry, error) {
	dirEntries, err := os.ReadDir(s.subdir(id))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, e := range dirEntries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}

		at, err := time.Parse(timeLayout, strings.TrimSuffix(e.Name(), ".json"))
		if err != nil {
			continue
		}

		entries = append(entries, Entry{
			Path: filepath.Join(s.subdir(id), e.Name()),
			Time: at,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time.After(entries[j].Time)
	})

	return entries, nil
}

func (s *Store) subdir(id int) string {
	return filepath.Join(s.dir, strconv.Itoa(id))
}
