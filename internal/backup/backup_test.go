package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type StoreSuite struct {
	suite.Suite

	dir   string
	store *Store
	clock time.Time
}

func (s *StoreSuite) SetupTest() {
	s.dir = s.T().TempDir()
	s.clock = time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)

	s.store = NewStore(s.dir)
	s.store.now = func() time.Time { return s.clock }
}

func (s *StoreSuite) advance(d time.Duration) {
	s.clock = s.clock.Add(d)
}

func TestStoreSuite(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}

func (s *StoreSuite) TestSaveWritesPrettyPrintedJSON() {
	entry, err := s.store.Save(42, []byte(`{"a":1,"b":2}`))
	s.Require().NoError(err)

	s.Equal(filepath.Join(s.dir, "42", "20260902T100000Z.json"), entry.Path)
	s.Equal(s.clock, entry.Time)

	data, err := os.ReadFile(entry.Path)
	s.Require().NoError(err)
	s.Equal("{\n    \"a\": 1,\n    \"b\": 2\n}", string(data))
}

func (s *StoreSuite) TestSaveRejectsInvalidJSON() {
	_, err := s.store.Save(42, []byte(`not json`))

	s.Error(err)
}

func (s *StoreSuite) TestSaveKeepsPreviousBackups() {
	first, err := s.store.Save(42, []byte(`{"v":"original"}`))
	s.Require().NoError(err)

	s.advance(time.Hour)
	second, err := s.store.Save(42, []byte(`{"v":"mangled"}`))
	s.Require().NoError(err)

	s.NotEqual(first.Path, second.Path)
	s.FileExists(first.Path)
	s.FileExists(second.Path)
}

func (s *StoreSuite) TestLoadReturnsMostRecent() {
	_, err := s.store.Save(42, []byte(`{"v":"original"}`))
	s.Require().NoError(err)

	s.advance(time.Hour)
	_, err = s.store.Save(42, []byte(`{"v":"newest"}`))
	s.Require().NoError(err)

	data, err := s.store.Load(42)
	s.Require().NoError(err)
	s.JSONEq(`{"v":"newest"}`, string(data))
}

func (s *StoreSuite) TestLoadWithNoBackups() {
	_, err := s.store.Load(999)

	s.ErrorIs(err, ErrNoBackups)
}

func (s *StoreSuite) TestListIsNewestFirst() {
	for _, v := range []string{"first", "second", "third"} {
		_, err := s.store.Save(42, []byte(`{"v":"`+v+`"}`))
		s.Require().NoError(err)
		s.advance(time.Hour)
	}

	entries, err := s.store.List(42)
	s.Require().NoError(err)
	s.Require().Len(entries, 3)

	for i := 1; i < len(entries); i++ {
		s.True(entries[i-1].Time.After(entries[i].Time), "entries should be newest first")
	}
}

func (s *StoreSuite) TestListWithNoBackups() {
	entries, err := s.store.List(999)

	s.NoError(err)
	s.Empty(entries)
}

func (s *StoreSuite) TestListIgnoresForeignFiles() {
	_, err := s.store.Save(42, []byte(`{"a":1}`))
	s.Require().NoError(err)

	for _, name := range []string{"notes.txt", "42.json", "README"} {
		err := os.WriteFile(filepath.Join(s.dir, "42", name), []byte("x"), 0o644)
		s.Require().NoError(err)
	}

	entries, err := s.store.List(42)
	s.Require().NoError(err)
	s.Len(entries, 1)
}

func (s *StoreSuite) TestBackupsAreIsolatedPerID() {
	_, err := s.store.Save(42, []byte(`{"v":"forty-two"}`))
	s.Require().NoError(err)
	_, err = s.store.Save(7, []byte(`{"v":"seven"}`))
	s.Require().NoError(err)

	data, err := s.store.Load(7)
	s.Require().NoError(err)
	s.JSONEq(`{"v":"seven"}`, string(data))
}
