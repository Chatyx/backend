package test

import (
	"net/url"
)

func (s *AppTestSuite) apiURLFromPath(path string) string {
	u, err := url.Parse("http://" + s.conf.API.Listen)
	s.Require().NoError(err, "Failed to parse api url")

	return u.JoinPath(path).String()
}

func (s *AppTestSuite) chatURL() string {
	u, err := url.Parse("ws://" + s.conf.Chat.Listen)
	s.Require().NoError(err, "Failed to parse chat url")

	return u.String()
}
