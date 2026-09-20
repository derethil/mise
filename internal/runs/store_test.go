package runs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type StoreSuite struct {
	suite.Suite

	dir string
}

func (s *StoreSuite) SetupTest() {
	s.dir = s.T().TempDir()
}

func TestStoreSuite(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}

func (s *StoreSuite) open(command string) *Store {
	store, err := Open(s.dir, command)
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = store.Close() })
	return store
}

func (s *StoreSuite) TestOpen_CreatesCacheDir() {
	nested := filepath.Join(s.dir, "nested", "cache")

	store, err := Open(nested, "normalize")
	s.Require().NoError(err)
	defer store.Close()

	_, err = os.Stat(nested)
	s.Require().NoError(err)
}

func (s *StoreSuite) TestHas_UnknownIDIsFalse() {
	store := s.open("normalize")

	s.False(store.Has(42))
}

func (s *StoreSuite) TestMark_MakesHasTrue() {
	store := s.open("normalize")

	s.Require().NoError(store.Mark(42, ResultOK))

	s.True(store.Has(42))
}

func (s *StoreSuite) TestMark_PersistsAcrossReopen() {
	store := s.open("normalize")
	s.Require().NoError(store.Mark(42, ResultOK))
	s.Require().NoError(store.Close())

	reopened, err := Open(s.dir, "normalize")
	s.Require().NoError(err)
	defer reopened.Close()

	s.True(reopened.Has(42))
}

func (s *StoreSuite) TestFailed_ReturnsOnlyFailedIDs() {
	store := s.open("keyword")

	s.Require().NoError(store.Mark(1, ResultFailure))
	s.Require().NoError(store.Mark(2, ResultOK))
	s.Require().NoError(store.Mark(3, ResultFailure))

	s.Equal([]int{1, 3}, store.Failed())
}

func (s *StoreSuite) TestMark_LaterResultOverridesEarlier() {
	store := s.open("keyword")

	s.Require().NoError(store.Mark(7, ResultFailure))
	s.Equal([]int{7}, store.Failed())

	s.Require().NoError(store.Mark(7, ResultOK))
	s.Empty(store.Failed())
	s.True(store.Has(7))
}

func (s *StoreSuite) TestReopen_ReplaysLatestResultPerID() {
	store := s.open("keyword")
	s.Require().NoError(store.Mark(9, ResultFailure))
	s.Require().NoError(store.Mark(9, ResultOK))
	s.Require().NoError(store.Close())

	reopened, err := Open(s.dir, "keyword")
	s.Require().NoError(err)
	defer reopened.Close()

	s.True(reopened.Has(9))
	s.Empty(reopened.Failed())
}

func (s *StoreSuite) TestSeparateCommandsHaveIndependentCaches() {
	normalize := s.open("normalize")
	keyword := s.open("keyword")

	s.Require().NoError(normalize.Mark(5, ResultOK))

	s.True(normalize.Has(5))
	s.False(keyword.Has(5))
}

func (s *StoreSuite) TestMarkOK_ClearsAPreviousFailure() {
	store := s.open("normalize")

	store.MarkFailure(context.Background(), 11)
	s.Equal([]int{11}, store.Failed())

	store.MarkOK(context.Background(), 11)
	s.Empty(store.Failed())
	s.True(store.Has(11))
}

func (s *StoreSuite) TestMarkFailure_ShowsUpInFailed() {
	store := s.open("normalize")

	store.MarkFailure(context.Background(), 22)

	s.Equal([]int{22}, store.Failed())
}
