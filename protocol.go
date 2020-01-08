package s2client

//
// Methods for communicating via request/response over the websocket
//

import (
	"github.com/golang/protobuf/proto"
	SC2APIProtocol "github.com/grantmd/go-s2client/sc2proto"
)

type Protocol struct {
	Conn *Conn
}

// SendRequest takes an API request struct, marshals it to JSON, and sends it over the connection
func (p *Protocol) SendRequest(req *SC2APIProtocol.Request) (err error) {
	data, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	return p.Conn.Write(data)
}

// ReadResponse reads a message off the connection and unmarshals it into an API response struct
func (p *Protocol) ReadResponse() (res *SC2APIProtocol.Response, err error) {
	data, err := p.Conn.Read()
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

// Disconnect wraps the underlying connection closure
func (p *Protocol) Disconnect() (err error) {
	return p.Conn.Close()
}
