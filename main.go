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
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	var c Conn
	log.Printf("Connecting to %s", *addr)

	err := c.Dial(addr)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Println("Successfully connected!")

	protocol := &Protocol{
		conn: &c,
	}
	defer protocol.Disconnect()

	var req *SC2APIProtocol.Request
	var resp *SC2APIProtocol.Response

	// Create a new game
	ourPlayer := &SC2APIProtocol.PlayerSetup{
		Type: SC2APIProtocol.PlayerType_Participant.Enum(),
		Race: SC2APIProtocol.Race_Terran.Enum(),
	}
	opponentPlayer := &SC2APIProtocol.PlayerSetup{
		Type:       SC2APIProtocol.PlayerType_Computer.Enum(),
		Race:       SC2APIProtocol.Race_Terran.Enum(),
		Difficulty: SC2APIProtocol.Difficulty_VeryHard.Enum(),
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
	log.Println("Starting game")
	err = protocol.SendRequest(req)
	if err != nil {
		log.Fatal("Could not send game start request:", err)
	}

	resp, err = protocol.ReadResponse()
	if err != nil {
		log.Fatal("Could not receive game start response:", err)
	}
	log.Println("Game started:", resp)

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
	log.Println("Joining game")
	err = protocol.SendRequest(req)
	if err != nil {
		log.Fatal("Could not send join game request:", err)
	}

	resp, err = protocol.ReadResponse()
	if err != nil {
		log.Fatal("Could not receive join game response:", err)
	}
	log.Println("Game joined:", resp)

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
		resp, err = protocol.ReadResponse()
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
	log.Println("Leaving game")
	err = protocol.SendRequest(req)
	if err != nil {
		log.Fatal("Could not send leave request:", err)
	}

	_, err = protocol.ReadResponse()
	if err != nil {
		log.Fatal("Could not receive leave response:", err)
	}

	// Disconnect
	log.Println("Disconnecting")
	err = protocol.Disconnect()
	if err != nil {
		log.Fatal("Error disconnecting:", err)
	}

	log.Println("Exiting")

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
