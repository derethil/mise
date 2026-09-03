package tandoor

import (
	"context"
	"encoding/json"
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
	lastAuth   string
	status     int
	response   map[string]any
	block      chan struct{}
}

func (s *ClientSuite) SetupTest() {
	s.lastMethod = ""
	s.lastPath = ""
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
		s.lastAuth = r.Header.Get("Authorization")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(s.status)
		_ = json.NewEncoder(w).Encode(s.response)
	}))

	s.client = NewClient(s.server.URL, "test-token")
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
