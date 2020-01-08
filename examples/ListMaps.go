package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/grantmd/go-s2client"
	"github.com/grantmd/go-s2client/sc2proto"
)

var gamePort = flag.Int("GamePort", 5000, "Ladder server port")
var gameServer = flag.String("GameServer", "localhost", "Address of server")

var quitRequested bool
var isMultiplayer bool

var abilities []*SC2APIProtocol.AbilityData // Most Useful: AbilityId, Available, FootprintRadius
var units []*SC2APIProtocol.UnitTypeData    // Most Useful: UnitId, Available, MineralCost, VespeneCost, FoodRequired
var upgrades []*SC2APIProtocol.UpgradeData  // Most Useful: UpgradeId, MineralCost, VespeneCost
var buffs []*SC2APIProtocol.BuffData        // Most Useful: BuffId

var mapUnits [][]uint32

func main() {
	// Parse command line args and configure logging
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	isMultiplayer = true

	// Connect to game server
	addr := fmt.Sprintf("%s:%d", *gameServer, *gamePort)
	var c s2client.Conn
	log.Printf("Connecting to %s…", addr)

	err := c.Dial(&addr)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Println("Successfully connected!")

	protocol := &s2client.Protocol{
		Conn: &c,
	}
	defer protocol.Disconnect()

	// Setup signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("Quit requested")
		protocol.Disconnect()
		os.Exit(1)
	}()

	// Start sending commands/reading responses
	var req *SC2APIProtocol.Request
	var resp *SC2APIProtocol.Response

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

	// Disconnect
	log.Println("Disconnecting…")
	err = protocol.Disconnect()
	if err != nil {
		log.Fatal("Error disconnecting:", err)
	}

	log.Println("BYE!")
}
