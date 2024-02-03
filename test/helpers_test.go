package test

import (
	"net/url"
)

func (s *AppTestSuite) apiURLFromPath(path string) *url.URL {
	u, _ := url.Parse("http://" + s.conf.API.Listen)
	return u.JoinPath(path)
}

func (s *AppTestSuite) chatURL() *url.URL {
	u, _ := url.Parse("ws://" + s.conf.Chat.Listen)
	return u
}
