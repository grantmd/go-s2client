package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/grantmd/go-s2client"
)

var gamePort = flag.Int("GamePort", 5000, "Ladder server port")
var gameServer = flag.String("GameServer", "localhost", "Address of server")

func main() {
	// Parse command line args and configure logging
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

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

	s2client.ListMaps(protocol)

	// Disconnect
	log.Println("Disconnecting…")
	err = protocol.Disconnect()
	if err != nil {
		log.Fatal("Error disconnecting:", err)
	}

	log.Println("BYE!")
}
