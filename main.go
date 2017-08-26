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
	log.Println("successfully connected")

	protocol := &Protocol{
		conn: &c,
	}
	defer protocol.Disconnect()

	var req *SC2APIProtocol.Request

	ourPlayer := &SC2APIProtocol.PlayerSetup{
		Type:       SC2APIProtocol.PlayerType_Participant.Enum(),
		Race:       SC2APIProtocol.Race_Terran.Enum(),
		Difficulty: SC2APIProtocol.Difficulty_VeryEasy.Enum(),
	}
	opponentPlayer := &SC2APIProtocol.PlayerSetup{
		Type: SC2APIProtocol.PlayerType_Computer.Enum(),
		Race: SC2APIProtocol.Race_Terran.Enum(),
	}

	req = &SC2APIProtocol.Request{
		Request: &SC2APIProtocol.Request_CreateGame{
			CreateGame: &SC2APIProtocol.RequestCreateGame{
				Map: &SC2APIProtocol.RequestCreateGame_BattlenetMapName{
					BattlenetMapName: "Antiga Shipyard",
				},
				PlayerSetup: []*SC2APIProtocol.PlayerSetup{ourPlayer, opponentPlayer},
			},
		},
	}
	/*req = &SC2APIProtocol.Request{
		Request: &SC2APIProtocol.Request_AvailableMaps{
			AvailableMaps: &SC2APIProtocol.RequestAvailableMaps{},
		},
	}*/
	/*req = &SC2APIProtocol.Request{
		Request: &SC2APIProtocol.Request_Ping{
			Ping: &SC2APIProtocol.RequestPing{},
		},
	}*/
	/*req = &SC2APIProtocol.Request{
		Request: &SC2APIProtocol.Request_Debug{
			Debug: &SC2APIProtocol.RequestDebug{},
		},
	}*/
	err = protocol.SendRequest(req)
	if err != nil {
		log.Fatal("Could not send request:", err)
	}

	_, err = protocol.ReadResponse()
	if err != nil {
		log.Fatal("Could not receive response:", err)
	}

	err = protocol.Disconnect()
	if err != nil {
		log.Fatal("Error disconnecting:", err)
	}

	log.Println("exiting")
}
