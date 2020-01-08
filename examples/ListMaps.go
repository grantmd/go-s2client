package main

import (
	"flag"
	"log"

	"github.com/grantmd/go-s2client"
)

var gameServer = flag.String("GameServer", "localhost", "Address of server")
var gamePort = flag.Int("GamePort", 12000, "Port of server")

func main() {
	// Parse command line args and configure logging
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	// Connect to game server
	protocol := s2client.Connect(*gameServer, *gamePort)
	defer protocol.Disconnect()

	// List the maps
	s2client.ListMaps(protocol)

	// Disconnect
	log.Println("Disconnecting…")
	err := protocol.Disconnect()
	if err != nil {
		log.Fatal("Error disconnecting:", err)
	}

	log.Println("BYE!")
}
