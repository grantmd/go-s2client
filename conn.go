package main

import (
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

type Conn struct {
	ws *websocket.Conn
}

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
	//defer c.ws.Close()

	return nil
}

func (c *Conn) Read() (message []byte, err error) {
	_, message, err = c.ws.ReadMessage()
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (c *Conn) Write(data []byte) (err error) {
	err = c.ws.WriteMessage(websocket.TextMessage, data)
	return err
}

func (c *Conn) Close() (err error) {
	err = c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	return err
}
