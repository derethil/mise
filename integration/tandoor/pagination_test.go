//go:build integration

package tandoor_test

import (
	"strconv"

	"github.com/derethil/mise/internal/tandoor"
)

const paginationPageSize = 2

func (s *TandoorSuite) TestPagination() {
	s.Run("ReturnsEveryRecordOnce", func() {
		keywords := s.allKeywordsPaged()

		s.Greater(len(s.fixtures.paginated), paginationPageSize, "the fixtures did not produce enough keywords to force a second request")

		counts := make(map[int]int, len(keywords))
		for _, keyword := range keywords {
			counts[keyword.ID]++
		}

		for _, record := range s.fixtures.paginated {
			s.Equal(1, counts[record.ID], "keyword %q appeared %d times across pages", record.Name, counts[record.ID])
		}
	})
}

func (s *TandoorSuite) allKeywordsPaged() []tandoor.Keyword {
	keywords, err := s.client.Keywords.All(s.T().Context(), "page_size", strconv.Itoa(paginationPageSize))
	s.Require().NoError(err)

	return keywords
}
