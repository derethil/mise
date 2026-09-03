package tandoor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
}

func (s *ClientSuite) SetupTest() {
	s.lastMethod = ""
	s.lastPath = ""
	s.lastAuth = ""
	s.status = http.StatusOK
	s.response = map[string]any{"id": 42, "name": "Tacos"}

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	s.server.Close()
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}
