package test

import (
	"net/http"

	ws "github.com/gorilla/websocket"
)

func (s *AppTestSuite) TestTest() {
	pair := s.authenticate("john1967", "qwerty12345")
	conn := s.newWebsocketConn(pair.AccessToken)
	conn.Close()
}

func (s *AppTestSuite) newWebsocketConn(accessToken string) *ws.Conn {
	dialer := &ws.Dialer{}

	reqHeaders := http.Header{}
	reqHeaders.Add("Authorization", "Bearer "+accessToken)

	conn, _, err := dialer.Dial(s.chatURL(), reqHeaders)
	s.Require().NoError(err, "Failed to open websocket connection")

	return conn
}
