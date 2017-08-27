package main

import (
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/grantmd/go-s2client/sc2proto"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var addr = flag.String("addr", "localhost:5000", "sc2api server address")
var listMaps = flag.Bool("listMaps", false, "list available maps")
var mapName = flag.String("mapName", "", "name of battlenet map to play")
var mapPath = flag.String("mapPath", "", "path of local map to play")

var quitRequested bool

func main() {
	// Setup signal handling
	quitRequested = false
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		quitRequested = true
	}()

	// Parse command line args and configure logging
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	// Connect to game server
	var c Conn
	log.Printf("Connecting to %s…", *addr)

	err := c.Dial(addr)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Println("Successfully connected!")

	protocol := &Protocol{
		conn: &c,
	}
	defer protocol.Disconnect()

	// Start sending commands/reading responses
	var req *SC2APIProtocol.Request
	var resp *SC2APIProtocol.Response

	if *listMaps == true {
		req = &SC2APIProtocol.Request{
			Request: &SC2APIProtocol.Request_AvailableMaps{
				AvailableMaps: &SC2APIProtocol.RequestAvailableMaps{},
			},
		}
		log.Println("Listing available maps…")
		err = protocol.SendRequest(req)
		if err != nil {
			log.Fatal("Could not send available maps request:", err)
		}

		resp, err = protocol.ReadResponse()
		if err != nil {
			log.Fatal("Could not receive available maps response:", err)
		}

		availableMaps := resp.GetAvailableMaps()
		fmt.Println("Local maps:")
		for _, localMap := range availableMaps.LocalMapPaths {
			fmt.Println(localMap)
		}

		fmt.Println("\nBattlenet maps:")
		for _, bnetMap := range availableMaps.BattlenetMapNames {
			fmt.Println(bnetMap)
		}
	}

	if *mapName != "" || *mapPath != "" {
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

		// Set either a battlenet map or a local map from args
		if *mapName != "" {
			req = &SC2APIProtocol.Request{
				Request: &SC2APIProtocol.Request_CreateGame{
					CreateGame: &SC2APIProtocol.RequestCreateGame{
						Map: &SC2APIProtocol.RequestCreateGame_BattlenetMapName{
							BattlenetMapName: *mapName,
						},
						PlayerSetup: []*SC2APIProtocol.PlayerSetup{ourPlayer, opponentPlayer},
						DisableFog:  proto.Bool(false),
						Realtime:    proto.Bool(true),
					},
				},
			}
		}
		if *mapPath != "" {
			req = &SC2APIProtocol.Request{
				Request: &SC2APIProtocol.Request_CreateGame{
					CreateGame: &SC2APIProtocol.RequestCreateGame{
						Map: &SC2APIProtocol.RequestCreateGame_LocalMap{
							LocalMap: &SC2APIProtocol.LocalMap{
								MapPath: mapPath,
							},
						},
						PlayerSetup: []*SC2APIProtocol.PlayerSetup{ourPlayer, opponentPlayer},
						DisableFog:  proto.Bool(false),
						Realtime:    proto.Bool(true),
					},
				},
			}
		}

		log.Println("Starting game…")
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
		log.Println("Joining game…")
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
			// Do we want to be done?
			if quitRequested == true {
				break
			}

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

			respObs := resp.GetObservation()
			// respObs.Actions - list of actions performed
			// respObs.ActionErrors - list of actions which did not complete
			// respObs.Observation - whole mess of observation data. see struct
			// respObs.PlayerResult - result of game, only if ended that step
			// respObs.Chat - chat messages received. could be fun

			obs := respObs.GetObservation()
			// obs.PlayerCommon - looks like player state. minerals, vespene, units, etc. very useful
			// obs.Alerts - critical end game actions like nuclear launch detected
			// obs.Abilities - list of available abilities? unclear. definitiion of AvailableAbility is in common.pb.go
			// obs.Score - I think how well you're doing, if you request score mode at game start
			// obs.RawData - Raw game data. Where the meat is of what we can see and do. Look at raw.pb.go for all the info
			// obs.FeatureLayerData - Probably not available unless you pick that game mode. Image based?
			// obs.RenderData - Full fidelity rendered image of the game. Not available yet
			// obs.UiData - Also not available yet

			log.Printf("step: %s", obs.PlayerCommon)

			// Are we done?
			if resp.GetStatus() == SC2APIProtocol.Status_ended {
				log.Println("Game over, man")
				log.Println(respObs.PlayerResult)
				break
			}

			// Examine game state
			for _, unit := range obs.RawData.Units {
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
		log.Println("Leaving game…")
		err = protocol.SendRequest(req)
		if err != nil {
			log.Fatal("Could not send leave request:", err)
		}

		_, err = protocol.ReadResponse()
		if err != nil {
			log.Fatal("Could not receive leave response:", err)
		}
		log.Println("Game left:", resp)
	}

	// Disconnect
	log.Println("Disconnecting…")
	err = protocol.Disconnect()
	if err != nil {
		log.Fatal("Error disconnecting:", err)
	}

	log.Println("gg")

	// Extra requests I've tested
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
