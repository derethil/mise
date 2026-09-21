package tandoor

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

const (
	MinSupportedVersion = "2.6.13"
	MaxSupportedVersion = "2.6.15"
)

func (c *Client) VersionSupported(ctx context.Context) (version string, supported bool, err error) {
	version, err = c.Version(ctx)
	if err != nil {
		return "", false, err
	}

	supported = compareVersions(version, MinSupportedVersion) >= 0 && compareVersions(version, MaxSupportedVersion) <= 0
	return version, supported, nil
}

func (c *Client) Version(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/openapi/?format=json", c.rootURL)
	body, err := c.requestURL(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	if !gjson.ValidBytes(body) {
		return "", fmt.Errorf("invalid openapi schema response")
	}

	version := gjson.GetBytes(body, "info.version")
	if !version.Exists() {
		return "", fmt.Errorf("openapi schema is missing info.version")
	}

	return parseTandoorVersion(version.String()), nil
}

func compareVersions(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")

	for i := 0; i < len(as) || i < len(bs); i++ {
		var an, bn int
		if i < len(as) {
			an, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			bn, _ = strconv.Atoi(bs[i])
		}

		if an != bn {
			if an < bn {
				return -1
			}
			return 1
		}
	}

	return 0
}

// Tandoor version strings are prefixed with 0.0.0 for some reason e.g. 0.0.0 (2.6.13)
func parseTandoorVersion(raw string) string {
	start := strings.IndexByte(raw, '(')
	end := strings.IndexByte(raw, ')')
	if start == -1 || end == -1 || end < start {
		return raw
	}

	return raw[start+1 : end]
}
