//go:build integration

package tandoor_test

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/derethil/mise/internal/tandoor"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

const (
	tandoorUser     = "mise"
	tandoorEmail    = "mise@example.invalid"
	tandoorPassword = "integration-test-user-password"
	tandoorSpace    = "mise"
)

const venvPython = "/opt/recipes/venv/bin/python"

const provisionSpaceScript = `
from django.contrib.auth.models import User
from cookbook.helper.permission_helper import create_space_for_user

create_space_for_user(User.objects.get(username=%q), %q)
`

func (s *TandoorSuite) bootstrapAuth(ctx context.Context) {
	s.provisionUser(ctx)
	s.provisionSpace(ctx)
	s.token = s.requestToken(ctx)
	s.client = tandoor.NewClient(s.baseURL, s.token)
}

func (s *TandoorSuite) provisionUser(ctx context.Context) {
	code, output := s.exec(ctx,
		[]string{venvPython, "manage.py", "createsuperuser", "--noinput"},
		tcexec.WithEnv([]string{
			"DJANGO_SUPERUSER_USERNAME=" + tandoorUser,
			"DJANGO_SUPERUSER_EMAIL=" + tandoorEmail,
			"DJANGO_SUPERUSER_PASSWORD=" + tandoorPassword,
		}),
	)
	s.Require().Zero(code, "createsuperuser exited %d: %s", code, output)
}

func (s *TandoorSuite) provisionSpace(ctx context.Context) {
	script := fmt.Sprintf(provisionSpaceScript, tandoorUser, tandoorSpace)

	code, output := s.exec(ctx, []string{venvPython, "manage.py", "shell", "-c", script})
	s.Require().Zero(code, "space provisioning exited %d: %s", code, output)
}

func (s *TandoorSuite) exec(ctx context.Context, command []string, options ...tcexec.ProcessOption) (int, string) {
	container, err := s.stack.ServiceContainer(ctx, tandoorContainer)
	s.Require().NoError(err, "resolve the %s container", tandoorContainer)

	code, reader, err := container.Exec(ctx, command, append(options, tcexec.Multiplexed())...)
	s.Require().NoError(err, "exec %s", command[1])

	output, err := io.ReadAll(reader)
	s.Require().NoError(err, "read the %s output", command[1])

	return code, string(output)
}

func (s *TandoorSuite) requestToken(ctx context.Context) string {
	credentials := map[string]string{
		"username": tandoorUser,
		"password": tandoorPassword,
	}

	var response struct {
		Token string `json:"token"`
	}

	s.request(ctx, http.MethodPost, "/api-token-auth/", credentials, &response)
	s.Require().NotEmpty(response.Token, "the token response contained no token")

	return response.Token
}
