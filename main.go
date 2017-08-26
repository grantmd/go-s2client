package main

import (
	"flag"
	"github.com/golang/protobuf/proto"
	"github.com/grantmd/go-s2client/sc2proto"
	"log"
	"time"
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

	// Create a new game
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
				DisableFog:  proto.Bool(false),
				Realtime:    proto.Bool(true),
			},
		},
	}
	err = protocol.SendRequest(req)
	if err != nil {
		log.Fatal("Could not send request:", err)
	}

	_, err = protocol.ReadResponse()
	if err != nil {
		log.Fatal("Could not receive response:", err)
	}

	// Join the game
	req = &SC2APIProtocol.Request{
		Request: &SC2APIProtocol.Request_JoinGame{
			JoinGame: &SC2APIProtocol.RequestJoinGame{
				Participation: &SC2APIProtocol.RequestJoinGame_Race{
					Race: SC2APIProtocol.Race_Terran,
				},
				Options: &SC2APIProtocol.InterfaceOptions{
					Raw: proto.Bool(true),
				},
			},
		},
	}
	err = protocol.SendRequest(req)
	if err != nil {
		log.Fatal("Could not send request:", err)
	}

	_, err = protocol.ReadResponse()
	if err != nil {
		log.Fatal("Could not receive response:", err)
	}

	// Game loop
	for {
		// Request observation
		req = &SC2APIProtocol.Request{
			Request: &SC2APIProtocol.Request_Observation{
				Observation: &SC2APIProtocol.RequestObservation{},
			},
		}
		err = protocol.SendRequest(req)
		if err != nil {
			log.Fatal("Could not send request:", err)
		}

		// Read the game state result
		resp, err := protocol.ReadResponse()
		if err != nil {
			log.Fatal("Could not receive response:", err)
		}

		obs := resp.GetObservation()

		// Are we done?
		if resp.GetStatus() == SC2APIProtocol.Status_ended {
			break
		}

		// Examine game state
		for _, unit := range obs.Observation.RawData.Units {
			if *unit.UnitType == 18 { // Command center

			}
		}

		// Send action

		// Keep this reasonably paced
		time.Sleep(100 * time.Millisecond)
	}

	// Leave game
	req = &SC2APIProtocol.Request{
		Request: &SC2APIProtocol.Request_LeaveGame{
			LeaveGame: &SC2APIProtocol.RequestLeaveGame{},
		},
	}
	err = protocol.SendRequest(req)
	if err != nil {
		log.Fatal("Could not send request:", err)
	}

	_, err = protocol.ReadResponse()
	if err != nil {
		log.Fatal("Could not receive response:", err)
	}

	// Disconnect
	err = protocol.Disconnect()
	if err != nil {
		log.Fatal("Error disconnecting:", err)
	}

	log.Println("exiting")

	// Extra requests I've tested
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
}
