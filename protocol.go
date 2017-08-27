package main

//
// Methods for communicating via request/response over the websocket
//

import (
	"github.com/golang/protobuf/proto"
	"github.com/grantmd/go-s2client/sc2proto"
)

type Protocol struct {
	conn *Conn
}

func (p *Protocol) SendRequest(req *SC2APIProtocol.Request) (err error) {
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

	res = &SC2APIProtocol.Response{}
	err = proto.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}

	// TODO: https://github.com/Blizzard/s2client-proto/blob/master/docs/protocol.md#protocol-errors
	// Should empty responses except for the `error` field populated return an error here? Or let callers handle it?

	return res, nil
}

func (p *Protocol) Disconnect() (err error) {
	return p.conn.Close()
}
