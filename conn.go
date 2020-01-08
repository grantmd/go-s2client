package s2client

//
// The basic websocket connection code
//

import (
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

type Conn struct {
	ws *websocket.Conn
}

// Dial takes an address of the form 127.0.0.1:12000 and opens a
// websocket connection to a starcraft 2 server
func (c *Conn) Dial(addr *string) (err error) {
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/sc2api"}

	originURL := u
	originURL.Scheme = "http"
	origin := originURL.String()

	headers := make(http.Header)
	headers.Add("Origin", origin)

	c.ws, _, err = websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		return err
	}

	return nil
}

// Read wraps reading messages off the websocket api connection
func (c *Conn) Read() (message []byte, err error) {
	_, message, err = c.ws.ReadMessage()
	if err != nil {
		return nil, err
	}

	return message, nil
}

// Write wraps writing messages to the websocket api connection
func (c *Conn) Write(data []byte) (err error) {
	err = c.ws.WriteMessage(websocket.TextMessage, data)
	return err
}

// Close wraps closing the websocket api connection
func (c *Conn) Close() (err error) {
	err = c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	return err
}
