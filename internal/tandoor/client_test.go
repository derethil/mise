package tandoor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ClientSuite struct {
	suite.Suite

	server *httptest.Server
	client *Client

	lastMethod string
	lastPath   string
	lastQuery  string
	lastAuth   string
	status     int
	response   map[string]any
	block      chan struct{}
}

func (s *ClientSuite) SetupTest() {
	s.lastMethod = ""
	s.lastPath = ""
	s.lastQuery = ""
	s.lastAuth = ""
	s.status = http.StatusOK
	s.response = map[string]any{"id": 42, "name": "Tacos"}

	s.block = nil

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.block != nil {
			select {
			case <-s.block:
			case <-r.Context().Done():
				return
			}
		}

		s.lastMethod = r.Method
		s.lastPath = r.URL.Path
		s.lastQuery = r.URL.RawQuery
		s.lastAuth = r.Header.Get("Authorization")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(s.status)
		_ = json.NewEncoder(w).Encode(s.response)
	}))

	s.client = NewClient(s.server.URL+"/api", "test-token")
}

func (s *ClientSuite) TearDownTest() {
	if s.block != nil {
		close(s.block)
		s.block = nil
	}
	s.server.Close()
}

func (s *ClientSuite) TestRequestTimesOut() {
	s.block = make(chan struct{})
	s.client.timeout = 20 * time.Millisecond

	_, err := s.client.Request(s.T().Context(), http.MethodGet, "recipe/42/", nil)

	s.Require().Error(err)
	s.ErrorIs(err, context.DeadlineExceeded)
}

func (s *ClientSuite) TestRequestHonoursCallerCancellation() {
	s.block = make(chan struct{})

	ctx, cancel := context.WithCancel(s.T().Context())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, err := s.client.Request(ctx, http.MethodGet, "recipe/42/", nil)

	s.Require().Error(err)
	s.ErrorIs(err, context.Canceled)
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

type paginatedItem struct {
	N int `json:"n"`
}

func (s *ClientSuite) TestRequestAllPagesFollowsNextLink() {
	s.server.Close()

	var page2URL string
	requests := 0

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(`{"results":[{"n":3}],"next":null}`))
			return
		}
		_, _ = fmt.Fprintf(w, `{"results":[{"n":1},{"n":2}],"next":%q}`, page2URL)
	}))
	page2URL = s.server.URL + "/api/items/?page=2"
	s.client = NewClient(s.server.URL+"/api", "test-token")

	items, err := RequestAllPages[paginatedItem](s.T().Context(), s.client, "items/")

	s.Require().NoError(err)
	s.Equal([]paginatedItem{{N: 1}, {N: 2}, {N: 3}}, items)
	s.Equal(2, requests)
}

func (s *ClientSuite) TestRequestAllPagesStopsWhenNextIsNil() {
	s.response = map[string]any{"results": []map[string]any{{"n": 1}}, "next": nil}

	items, err := RequestAllPages[paginatedItem](s.T().Context(), s.client, "items/")

	s.Require().NoError(err)
	s.Equal([]paginatedItem{{N: 1}}, items)
}

func (s *ClientSuite) TestRequestAllPages_HTTPError() {
	s.status = http.StatusForbidden

	_, err := RequestAllPages[paginatedItem](s.T().Context(), s.client, "items/")

	s.Error(err)
}

func (s *ClientSuite) TestRequestAllPages_InvalidJSON() {
	s.server.Close()
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not json`))
	}))
	s.client = NewClient(s.server.URL+"/api", "test-token")

	_, err := RequestAllPages[paginatedItem](s.T().Context(), s.client, "items/")

	s.Error(err)
}

func (s *ClientSuite) TestRequestRejectsNonJSONResponse() {
	s.server.Close()
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><body>Log in</body></html>"))
	}))
	s.client = NewClient(s.server.URL, "test-token")

	_, err := s.client.Request(s.T().Context(), http.MethodGet, "recipe/31/", nil)

	s.Require().Error(err)
	s.ErrorIs(err, ErrTandoorRequestFailed)
	s.Contains(err.Error(), "instead of JSON")
}
