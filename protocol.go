package main

import (
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/grantmd/go-s2client/sc2proto"
)

type Protocol struct {
	conn *Conn
}

func (p *Protocol) SendRequest(req *SC2APIProtocol.Request) (err error) {
	log.Printf("Request: %s", req)
	data, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	return p.conn.Write(data)
}

func (p *Protocol) ReadResponse() (res *SC2APIProtocol.Response, err error) {
	data, err := p.conn.Read()
	if err != nil {
		return nil, err
	}

	err = proto.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}

	log.Printf("Response: %s", res)
	return res, nil
}

func (p *Protocol) Disconnect() (err error) {
	return p.conn.Close()
}
