package main

import (
	"./sc2proto/sc2api.pb.go"
	"github.com/golang/protobuf/proto"
)

type Protocol struct {
	c *Conn
}
