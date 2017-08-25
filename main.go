package main

import (
	"flag"
	"github.com/grantmd/go-s2client/sc2proto"
	"log"
)

var addr = flag.String("addr", "localhost:5000", "sc2api server address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	var c Conn
	log.Printf("connecting to %s", *addr)

	err := c.Dial(addr)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	log.Println("successfully connected")

	protocol := &Protocol{
		conn: &c,
	}

	req := &SC2APIProtocol.Request{
		Request: &SC2APIProtocol.Request_CreateGame{
			CreateGame: &SC2APIProtocol.RequestCreateGame{
				Map: &SC2APIProtocol.RequestCreateGame_BattlenetMapName{},
			},
		},
	}
	err = protocol.SendRequest(req)
	if err != nil {
		log.Fatal("Could not send request:", err)
	}

	res, err := protocol.ReadResponse()
	if err != nil {
		log.Fatal("Could not receive response:", err)
	}
	log.Printf("Got response: %s", res)
}
