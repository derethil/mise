package video

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ValidateSuite struct {
	suite.Suite
}

func TestValidateSuite(t *testing.T) {
	suite.Run(t, new(ValidateSuite))
}

func (s *ValidateSuite) TestIsSupportedImportSource() {
	tests := []struct {
		host string
		want bool
	}{
		{"tiktok.com", true},
		{"www.tiktok.com", true},
		{"vm.tiktok.com", true},
		{"TikTok.com", true},
		{"youtube.com", true},
		{"m.youtube.com", true},
		{"instagram.com", true},
		{"www.instagram.com", true},
		{"Instagram.com", true},
		{"vimeo.com", false},
		{"", false},
		{"nottiktok.com", false},
		{"tiktok.com.evil.test", false},
		{"faketiktok.com", false},
		{"instagram.com.evil.test", false},
		{"fakeinstagram.com", false},
	}

	for _, tt := range tests {
		s.Run(tt.host, func() {
			s.Equal(tt.want, IsSupportedImportSource(tt.host))
		})
	}
}

func (s *ValidateSuite) TestIsSupportedBrowser() {
	s.True(IsSupportedBrowser("firefox"))
	s.True(IsSupportedBrowser("Firefox"), "browser names are matched case-insensitively")
	s.False(IsSupportedBrowser("netscape"))
	s.False(IsSupportedBrowser(""))
}
