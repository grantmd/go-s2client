package main

import (
	"net/url"

	"github.com/gorilla/websocket"
)

type Conn struct {
	ws *websocket.Conn
}

func (c *Conn) Dial(addr *string) (err error) {
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/sc2api"}
	c.ws, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	defer c.ws.Close()
	return err
}

func (c *Conn) Read() (message []byte, err error) {
	_, message, err = c.ws.ReadMessage()
	return message, err
}

func (c *Conn) Write(data []byte) (err error) {
	err = c.ws.WriteMessage(websocket.TextMessage, data)
	return err
}

func (c *Conn) Close() (err error) {
	err = c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	return err
}
