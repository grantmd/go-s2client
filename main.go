package main

import (
	"flag"
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
}
