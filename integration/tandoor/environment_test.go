//go:build integration

package tandoor_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/derethil/mise/internal/tandoor"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	composeFile       = "compose.yaml"
	databaseContainer = "db_recipes"
	tandoorContainer  = "web_recipes"
	tandoorPort       = "80/tcp"
	tandoorHealth     = "/openapi/?format=json"
)

const (
	startupTimeout  = 3 * time.Minute
	teardownTimeout = time.Minute
)

const (
	postgresDB       = "tandoor_test"
	postgresUser     = "tandoor_test"
	postgresPassword = "integration-test-password"
	secretKey        = "integration-test-secret-key"
)

type TandoorSuite struct {
	suite.Suite

	identifier string
	version    string
	stack      compose.ComposeStack
	baseURL    string
	token      string
	client     *tandoor.Client
	fixtures   fixtures
}

func TestTandoorSuite(t *testing.T) {
	suite.Run(t, new(TandoorSuite))
}

func (s *TandoorSuite) SetupSuite() {
	s.version = os.Getenv("TANDOOR_VERSION")
	s.Require().NotEmpty(s.version, "TANDOOR_VERSION is not set")

	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	s.createStack()
	s.startStack(ctx)
	s.resolveBaseURL(ctx)
	s.bootstrapAuth(ctx)
	s.createFixtures(ctx)

	s.T().Logf("tandoor %s ready at %s (compose project %s)", s.version, s.baseURL, s.identifier)
}

func (s *TandoorSuite) createStack() {
	s.identifier = fmt.Sprintf("mise-tandoor-%d", time.Now().UnixNano())

	stack, err := compose.NewDockerComposeWith(
		compose.StackIdentifier(s.identifier),
		compose.WithStackFiles(composeFile),
	)
	s.Require().NoError(err, "create compose project %s", s.identifier)

	s.stack = stack
}

func (s *TandoorSuite) startStack(ctx context.Context) {
	err := s.stack.
		WithEnv(s.composeEnv()).
		WaitForService(databaseContainer, wait.ForHealthCheck().WithStartupTimeout(startupTimeout)).
		WaitForService(tandoorContainer, wait.ForHTTP(tandoorHealth).WithPort(tandoorPort).WithStartupTimeout(startupTimeout)).
		Up(ctx, compose.Wait(true))
	if err != nil {
		s.logContainers()
	}

	s.Require().NoError(err, "start vabene1111/recipes:%s as compose project %s", s.version, s.identifier)
}

func (s *TandoorSuite) composeEnv() map[string]string {
	return map[string]string{
		"TANDOOR_VERSION":   s.version,
		"POSTGRES_DB":       postgresDB,
		"POSTGRES_USER":     postgresUser,
		"POSTGRES_PASSWORD": postgresPassword,
		"SECRET_KEY":        secretKey,
	}
}

func (s *TandoorSuite) resolveBaseURL(ctx context.Context) {
	container, err := s.stack.ServiceContainer(ctx, tandoorContainer)
	s.Require().NoError(err, "resolve the %s container", tandoorContainer)

	s.baseURL, err = container.PortEndpoint(ctx, tandoorPort, "http")
	s.Require().NoError(err, "resolve the %s endpoint", tandoorContainer)
}

func (s *TandoorSuite) TearDownSuite() {
	if s.stack == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), teardownTimeout)
	defer cancel()

	err := s.stack.Down(
		ctx,
		compose.RemoveOrphans(true),
		compose.RemoveVolumes(true),
	)

	s.Assert().NoError(err, "tear down compose project %s", s.identifier)
}
