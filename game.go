package s2client

import (
	"fmt"
	"log"
	"sort"

	"github.com/golang/protobuf/proto"
	SC2APIProtocol "github.com/grantmd/go-s2client/sc2proto"
)

// Connect takes an ip address and port and connects to it
func Connect(ip string, port int) *Protocol {
	addr := fmt.Sprintf("%s:%d", ip, port)
	var c Conn
	log.Printf("Connecting to %s…", addr)

	err := c.Dial(&addr)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Println("Successfully connected!")

	protocol := &Protocol{
		Conn: &c,
	}

	return protocol
}

// ListMaps takes an open connection, asks the server for a list of available
// maps, and prints them out
func ListMaps(sc2 *Protocol) {
	// Start sending commands/reading responses
	var req *SC2APIProtocol.Request
	var resp *SC2APIProtocol.Response

	req = &SC2APIProtocol.Request{
		Request: &SC2APIProtocol.Request_AvailableMaps{
			AvailableMaps: &SC2APIProtocol.RequestAvailableMaps{},
		},
	}
	log.Println("Listing available maps…")
	err := sc2.SendRequest(req)
	if err != nil {
		log.Fatal("Could not send available maps request:", err)
	}

	resp, err = sc2.ReadResponse()
	if err != nil {
		log.Fatal("Could not receive available maps response:", err)
	}

	availableMaps := resp.GetAvailableMaps()
	sort.Strings(availableMaps.LocalMapPaths)
	fmt.Println("Local maps:")
	for _, localMap := range availableMaps.LocalMapPaths {
		fmt.Println(localMap)
	}

	sort.Strings(availableMaps.BattlenetMapNames)
	fmt.Println("\nBattlenet maps:")
	for _, bnetMap := range availableMaps.BattlenetMapNames {
		fmt.Println(bnetMap)
	}
}

// CreateLocalGame takes an existing open connection and a path to
// a "local map" (like the mini games), and tells the server to create
// a game for it
func CreateLocalGame(sc2 *Protocol, mapPath string, realtime bool) {
	createGame(sc2, "", mapPath, realtime)
}

// CreateBattlentGame takes an existing open connection and a Battle.Net
// map name, and tells the server to create a game for it
func CreateBattlenetGame(sc2 *Protocol, mapName string, realtime bool) {
	createGame(sc2, mapName, "", realtime)
}

func createGame(sc2 *Protocol, mapName string, mapPath string, realtime bool) {

	var req *SC2APIProtocol.Request
	var resp *SC2APIProtocol.Response

	if mapName != "" || mapPath != "" {
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
		if mapName != "" {
			req = &SC2APIProtocol.Request{
				Request: &SC2APIProtocol.Request_CreateGame{
					CreateGame: &SC2APIProtocol.RequestCreateGame{
						Map: &SC2APIProtocol.RequestCreateGame_BattlenetMapName{
							BattlenetMapName: mapName,
						},
						PlayerSetup: []*SC2APIProtocol.PlayerSetup{ourPlayer, opponentPlayer},
						DisableFog:  proto.Bool(false),
						Realtime:    proto.Bool(realtime),
					},
				},
			}
		}

		if mapPath != "" {
			req = &SC2APIProtocol.Request{
				Request: &SC2APIProtocol.Request_CreateGame{
					CreateGame: &SC2APIProtocol.RequestCreateGame{
						Map: &SC2APIProtocol.RequestCreateGame_LocalMap{
							LocalMap: &SC2APIProtocol.LocalMap{
								MapPath: &mapPath,
							},
						},
						PlayerSetup: []*SC2APIProtocol.PlayerSetup{ourPlayer, opponentPlayer},
						DisableFog:  proto.Bool(false),
						Realtime:    proto.Bool(realtime),
					},
				},
			}
		}

		log.Println("Starting game…")
		err := sc2.SendRequest(req)
		if err != nil {
			log.Fatal("Could not send game start request:", err)
		}
		log.Println("Request sent")

		resp, err = sc2.ReadResponse()
		if err != nil {
			log.Fatal("Could not receive game start response:", err)
		}
		log.Println("Game started:", resp)
		// TODO: Handle this: "Game started: create_game:<error:InvalidMapPath error_details:"map_path '/SC2/StarCraftII/maps/CollectMineralsAndGas.SC2Map' file doesn't exist." > status:launched"
	}
}

// List of unit/ability/upgrade/buff types:
// https://github.com/Blizzard/s2client-api/blob/master/include/sc2api/sc2_typeenums.h

// Best Scores:
// MoveToBeacon: 27
// CollectMineralShards: 121
// CollectMineralsAndGas: 6114
// BuildMarines: 30
// DefeatZerglingsAndBanelings: 67
// DefeatRoaches: 46
// FindAndDefeatZerglings: 10