package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/golang/protobuf/proto"
	. "github.com/grantmd/go-s2client"
	"github.com/grantmd/go-s2client/sc2proto"
)

var gamePort = flag.Int("GamePort", 5677, "Ladder server port")
var startPort = flag.Int("StartPort", 5690, "Ladder server game port")
var ladderServer = flag.String("LadderServer", "localhost", "Address of ladder server")
var opponentID = flag.String("OpponentId", "", "Ladder ID of opponent")

var quitRequested bool
var isMultiplayer bool
var realtime bool

var abilities []*SC2APIProtocol.AbilityData // Most Useful: AbilityId, Available, FootprintRadius
var units []*SC2APIProtocol.UnitTypeData    // Most Useful: UnitId, Available, MineralCost, VespeneCost, FoodRequired
var upgrades []*SC2APIProtocol.UpgradeData  // Most Useful: UpgradeId, MineralCost, VespeneCost
var buffs []*SC2APIProtocol.BuffData        // Most Useful: BuffId

var mapUnits [][]uint32

func main() {
	// Parse command line args and configure logging
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	realtime = false
	isMultiplayer = true
	// Connect to game server
	addr := fmt.Sprintf("%s:%d", *ladderServer, *gamePort)
	var c Conn
	log.Printf("Connecting to %s…", addr)

	err := c.Dial(&addr)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Println("Successfully connected!")

	protocol := &Protocol{
		conn: &c,
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

	// Join the game
	CurrentPort := *startPort
	CurrentPort++
	ReqGamePort := CurrentPort
	CurrentPort++
	ReqServerGamePort := CurrentPort
	CurrentPort++
	ReqServerBasePort := CurrentPort
	CurrentPort++
	ReqClientGamePort := CurrentPort
	CurrentPort++
	ReqClientBasePort := CurrentPort
	req := &SC2APIProtocol.Request{
		Request: &SC2APIProtocol.Request_JoinGame{
			JoinGame: &SC2APIProtocol.RequestJoinGame{
				Participation: &SC2APIProtocol.RequestJoinGame_Race{
					Race: SC2APIProtocol.Race_Terran,
				},
				Options: &SC2APIProtocol.InterfaceOptions{
					Raw:   proto.Bool(true),
					Score: proto.Bool(true),
				},
				ServerPorts: &SC2APIProtocol.PortSet{
					GamePort: proto.Int(ReqServerGamePort),
					BasePort: proto.Int(ReqServerBasePort),
				},
				ClientPorts: []*SC2APIProtocol.PortSet{
					&SC2APIProtocol.PortSet{
						GamePort: proto.Int(ReqClientGamePort),
						BasePort: proto.Int(ReqClientBasePort),
					},
				},

				SharedPort: proto.Int(ReqGamePort),
			},
		},
	}
	// Should change indentation here, probably need some auto-format tool at some point, but cba right now
	log.Println("Joining game…")
	err = protocol.SendRequest(req)
	if err != nil {
		log.Fatal("Could not send join game request:", err)
	}

	resp, err := protocol.ReadResponse()
	if err != nil {
		log.Fatal("Could not receive join game response:", err)
	}
	log.Println("Game joined:", resp)

	// Get game info
	req = &SC2APIProtocol.Request{
		Request: &SC2APIProtocol.Request_GameInfo{
			GameInfo: &SC2APIProtocol.RequestGameInfo{},
		},
	}
	log.Println("Requesting game info…")
	err = protocol.SendRequest(req)
	if err != nil {
		log.Fatal("Could not send game info request:", err)
	}

	resp, err = protocol.ReadResponse()
	if err != nil {
		log.Fatal("Could not receive game info response:", err)
	}
	if len(resp.GetGameInfo().PlayerInfo) > 1 {
		isMultiplayer = true
		log.Println("Game is multiplayer")
	}
	// TODO: StartRaw has tons of map data in raw.pb.go
	// Need to figure out what ImageData.Data is. Unit type per map unit? Top to bottom or left to right?
	// PathingGrid might be great for figuring out where to send units, but the game kind of figures this out for us
	// PlacementGrid is maybe indicative of where units are on map start?
	// PlayableArea is maybe where units are placeable vs full map size?
	mapX := resp.GetGameInfo().GetStartRaw().GetMapSize().GetX()
	mapY := resp.GetGameInfo().GetStartRaw().GetMapSize().GetY()
	mapUnits := make([][]uint32, mapY)
	for i := range mapUnits {
		mapUnits[i] = make([]uint32, mapX)
	}

	log.Println("Game data received")

	// Get game data
	req = &SC2APIProtocol.Request{
		Request: &SC2APIProtocol.Request_Data{
			Data: &SC2APIProtocol.RequestData{
				AbilityId:  proto.Bool(true),
				UnitTypeId: proto.Bool(true),
				UpgradeId:  proto.Bool(true),
				BuffId:     proto.Bool(true),
			},
		},
	}
	log.Println("Requesting game data")
	err = protocol.SendRequest(req)
	if err != nil {
		log.Fatal("Could not send game data request:", err)
	}

	resp, err = protocol.ReadResponse()
	if err != nil {
		log.Fatal("Could not receive game data response:", err)
	}
	abilities = resp.GetData().GetAbilities()
	units = resp.GetData().GetUnits()
	upgrades = resp.GetData().GetUpgrades()
	buffs = resp.GetData().GetBuffs()
	log.Println("Game data received")

	// TODO: Grab all units now and record on the map where they are?

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

		respObs := resp.GetObservation()
		// respObs.Actions - list of actions performed
		// respObs.ActionErrors - list of actions which did not complete
		// respObs.Observation - whole mess of observation data. see struct and notes below
		// respObs.PlayerResult - result of game, only if ended that step
		// respObs.Chat - chat messages received. could be fun. maybe take commands from chat?

		obs := respObs.GetObservation()
		rawData := obs.GetRawData()
		// obs.PlayerCommon - looks like player state. minerals, vespene, units, etc. very useful
		// obs.Alerts - critical end game actions like nuclear launch detected
		// obs.Abilities - list of available abilities? unclear. definitiion of AvailableAbility is in common.pb.go
		// obs.Score - I think how well you're doing, if you request score mode at game start
		// obs.RawData - Raw game data. Where the meat is of what we can see and do. Look at raw.pb.go for all the info
		// obs.FeatureLayerData - Probably not available unless you pick that game mode. Image based?
		// obs.RenderData - Full fidelity rendered image of the game. Not available yet
		// obs.UiData - Also not available yet

		// Print our rough state of game every 10 steps
		if obs.GetGameLoop()%100 == 0 {
			//log.Println(rawData.Units)
			log.Println(obs.PlayerCommon)
			log.Printf("%s Score: %d", SC2APIProtocol.Score_ScoreType_name[int32(*obs.Score.ScoreType)], int32(*obs.Score.Score))
		}

		// Are we done?
		if resp.GetStatus() == SC2APIProtocol.Status_ended {
			log.Println("Game over, man")
			log.Println(respObs.PlayerResult)
			log.Println(obs.Score)
			break
		}

		// Prep request for action in case we need it
		req = &SC2APIProtocol.Request{
			Request: &SC2APIProtocol.Request_Action{
				Action: &SC2APIProtocol.RequestAction{},
			},
		}
		action := req.GetAction()

		// Examine game state
		var unitType uint32
		var alliance SC2APIProtocol.Alliance
		var target *SC2APIProtocol.Unit
		for _, unit := range rawData.Units {
			unitType = unit.GetUnitType()
			alliance = unit.GetAlliance()

			if unitType == 48 { // Marine
				if alliance == SC2APIProtocol.Alliance_Self && len(unit.GetOrders()) == 0 {

					// This is for "MoveToBeacon"
					target = FindClosestUnit(rawData.Units, unit, SC2APIProtocol.Alliance_Neutral, 317) // beacon
					if target != nil {
						var abilityID int32 = 1 // "SMART". Could also be 16, which is "MOVE"
						a := &SC2APIProtocol.Action{
							ActionRaw: &SC2APIProtocol.ActionRaw{
								Action: &SC2APIProtocol.ActionRaw_UnitCommand{
									UnitCommand: &SC2APIProtocol.ActionRawUnitCommand{
										AbilityId: &abilityID,
										Target: &SC2APIProtocol.ActionRawUnitCommand_TargetWorldSpacePos{
											TargetWorldSpacePos: &SC2APIProtocol.Point2D{
												X: target.Pos.X,
												Y: target.Pos.Y,
											},
										},
										UnitTags: []uint64{unit.GetTag()},
									},
								},
							},
						}
						action.Actions = append(action.Actions, a)
						log.Printf("Moving marine %d to beacon %d", unit.GetTag(), target.GetTag())
						continue
					}

					// This is for "CollectMineralShards"
					target = FindClosestUnit(rawData.Units, unit, SC2APIProtocol.Alliance_Neutral, 1680) // mineral shard
					// TODO: Make sure this isn't already someone else's target, somehow
					if target != nil {
						var abilityID int32 = 1 // "SMART". Could also be 16, which is "MOVE"
						a := &SC2APIProtocol.Action{
							ActionRaw: &SC2APIProtocol.ActionRaw{
								Action: &SC2APIProtocol.ActionRaw_UnitCommand{
									UnitCommand: &SC2APIProtocol.ActionRawUnitCommand{
										AbilityId: &abilityID,
										Target: &SC2APIProtocol.ActionRawUnitCommand_TargetWorldSpacePos{
											TargetWorldSpacePos: &SC2APIProtocol.Point2D{
												X: target.Pos.X,
												Y: target.Pos.Y,
											},
										},
										UnitTags: []uint64{unit.GetTag()},
									},
								},
							},
						}
						action.Actions = append(action.Actions, a)
						log.Printf("Moving marine %d to mineral shard %d", unit.GetTag(), target.GetTag())
						continue
					}

					// Attack any enemy we can see
					target = FindClosestAnyEnemy(rawData.Units, unit)
					if target != nil {
						var abilityID int32 = 3674 // "ATTACK".
						a := &SC2APIProtocol.Action{
							ActionRaw: &SC2APIProtocol.ActionRaw{
								Action: &SC2APIProtocol.ActionRaw_UnitCommand{
									UnitCommand: &SC2APIProtocol.ActionRawUnitCommand{
										AbilityId: &abilityID,
										Target: &SC2APIProtocol.ActionRawUnitCommand_TargetUnitTag{
											TargetUnitTag: target.GetTag(),
										},
										UnitTags: []uint64{unit.GetTag()},
									},
								},
							},
						}
						action.Actions = append(action.Actions, a)
						log.Printf("Moving marine %d to attack enemy %d of type %d", unit.GetTag(), target.GetTag(), target.GetUnitType())
						continue
					}

					// Explore randomly (TODO: Could maybe better explore together)
					var abilityID int32 = 1 // "SMART". Could also be 16, which is "MOVE"
					offset := float32(32.0)
					rx := float32(*unit.Pos.X + rand.Float32()*offset)
					ry := float32(*unit.Pos.Y + rand.Float32()*offset)
					a := &SC2APIProtocol.Action{
						ActionRaw: &SC2APIProtocol.ActionRaw{
							Action: &SC2APIProtocol.ActionRaw_UnitCommand{
								UnitCommand: &SC2APIProtocol.ActionRawUnitCommand{
									AbilityId: &abilityID,
									Target: &SC2APIProtocol.ActionRawUnitCommand_TargetWorldSpacePos{
										TargetWorldSpacePos: &SC2APIProtocol.Point2D{
											X: &rx,
											Y: &ry,
										},
									},
									UnitTags: []uint64{unit.GetTag()},
								},
							},
						},
					}
					action.Actions = append(action.Actions, a)
					log.Printf("Moving marine %d to explore at %f,%f", unit.GetTag(), rx, ry)
					continue
				}
			}

			if unitType == 45 { // TERRAN SCV
				// This is for "CollectMineralsAndGas"
				if alliance == SC2APIProtocol.Alliance_Self {
					if obs.PlayerCommon.GetMinerals() >= 100 && obs.PlayerCommon.GetFoodCap()-obs.PlayerCommon.GetFoodUsed() <= 2 && AnyUnitHasOrder(rawData.Units, 45, 319) == false && HasActionQueued(action.Actions, 319) == false { // TODO: Way to find out cost programmatically?
						var abilityID int32 = 319 // "BUILD_SUPPLYDEPOT"

						offset := float32(15.0)
						rx := float32(*unit.Pos.X + rand.Float32()*offset)
						ry := float32(*unit.Pos.Y + rand.Float32()*offset)

						a := &SC2APIProtocol.Action{
							ActionRaw: &SC2APIProtocol.ActionRaw{
								Action: &SC2APIProtocol.ActionRaw_UnitCommand{
									UnitCommand: &SC2APIProtocol.ActionRawUnitCommand{
										AbilityId: &abilityID,
										Target: &SC2APIProtocol.ActionRawUnitCommand_TargetWorldSpacePos{
											TargetWorldSpacePos: &SC2APIProtocol.Point2D{
												X: &rx,
												Y: &ry,
											},
										},
										UnitTags: []uint64{unit.GetTag()},
									},
								},
							},
						}
						action.Actions = append(action.Actions, a)
						log.Printf("SCV %d building supply depot at %f,%f", unit.GetTag(), rx, ry)
						continue
					}

					if len(unit.GetOrders()) == 0 {
						if obs.PlayerCommon.GetMinerals() >= 150 && isMultiplayer && (CountUnitsOfType(rawData.Units, SC2APIProtocol.Alliance_Self, 21)) == 0 { // TODO: Way to find out cost programmatically? Check available actions instead of map name
							if AnyUnitHasOrder(rawData.Units, 45, 321) == false { // Only build one at a time
								var abilityID int32 = 321 // "BUILD_BARRACKS"

								offset := float32(30.0)
								rx := float32(*unit.Pos.X + rand.Float32()*offset)
								ry := float32(*unit.Pos.Y + rand.Float32()*offset)

								a := &SC2APIProtocol.Action{
									ActionRaw: &SC2APIProtocol.ActionRaw{
										Action: &SC2APIProtocol.ActionRaw_UnitCommand{
											UnitCommand: &SC2APIProtocol.ActionRawUnitCommand{
												AbilityId: &abilityID,
												Target: &SC2APIProtocol.ActionRawUnitCommand_TargetWorldSpacePos{
													TargetWorldSpacePos: &SC2APIProtocol.Point2D{
														X: &rx,
														Y: &ry,
													},
												},
												UnitTags: []uint64{unit.GetTag()},
											},
										},
									},
								}
								action.Actions = append(action.Actions, a)
								log.Printf("SCV %d building barracks at %f,%f", unit.GetTag(), rx, ry)
								continue
							}
						}

						if obs.PlayerCommon.GetMinerals() >= 75 { // TODO: Way to find out cost programmatically?
							if AnyUnitHasOrder(rawData.Units, 45, 320) == false { // Only build one at a time
								target = FindClosestUnit(rawData.Units, unit, SC2APIProtocol.Alliance_Neutral, 342) // vespene geyser
								if target != nil && IsUnitTypeAtPoint(rawData.Units, 20, *target.GetPos()) == false {
									var abilityID int32 = 320 // "BUILD_REFINERY"
									a := &SC2APIProtocol.Action{
										ActionRaw: &SC2APIProtocol.ActionRaw{
											Action: &SC2APIProtocol.ActionRaw_UnitCommand{
												UnitCommand: &SC2APIProtocol.ActionRawUnitCommand{
													AbilityId: &abilityID,
													Target: &SC2APIProtocol.ActionRawUnitCommand_TargetUnitTag{
														TargetUnitTag: target.GetTag(),
													},
													UnitTags: []uint64{unit.GetTag()},
												},
											},
										},
									}
									action.Actions = append(action.Actions, a)
									log.Printf("SCV %d building refinery at %d", unit.GetTag(), target.GetTag())
									continue
								}
							}
						}

						target = FindClosestUnit(rawData.Units, unit, SC2APIProtocol.Alliance_Self, 20) // terran refinery
						if target != nil && target.GetAssignedHarvesters() < target.GetIdealHarvesters() {
							var abilityID int32 = 3666 // "HARVEST_GATHER". There are other "harvest gather" abilities. What are they for?
							a := &SC2APIProtocol.Action{
								ActionRaw: &SC2APIProtocol.ActionRaw{
									Action: &SC2APIProtocol.ActionRaw_UnitCommand{
										UnitCommand: &SC2APIProtocol.ActionRawUnitCommand{
											AbilityId: &abilityID,
											Target: &SC2APIProtocol.ActionRawUnitCommand_TargetUnitTag{
												TargetUnitTag: target.GetTag(),
											},
											UnitTags: []uint64{unit.GetTag()},
										},
									},
								},
							}
							action.Actions = append(action.Actions, a)
							log.Printf("Moving SCV %d to finery %d", unit.GetTag(), target.GetTag())
							continue
						}

						target = FindClosestUnit(rawData.Units, unit, SC2APIProtocol.Alliance_Neutral, 341) // mineral field
						if target != nil && target.GetAssignedHarvesters() < 2 {
							var abilityID int32 = 3666 // "HARVEST_GATHER". There are other "harvest gather" abilities. What are they for?
							a := &SC2APIProtocol.Action{
								ActionRaw: &SC2APIProtocol.ActionRaw{
									Action: &SC2APIProtocol.ActionRaw_UnitCommand{
										UnitCommand: &SC2APIProtocol.ActionRawUnitCommand{
											AbilityId: &abilityID,
											Target: &SC2APIProtocol.ActionRawUnitCommand_TargetUnitTag{
												TargetUnitTag: target.GetTag(),
											},
											UnitTags: []uint64{unit.GetTag()},
										},
									},
								},
							}
							action.Actions = append(action.Actions, a)
							log.Printf("Moving SCV %d to mineral field %d", unit.GetTag(), target.GetTag())
							continue
						}
					}
				}
			}

			if unitType == 18 { // Terran command center
				// This is for "CollectMineralsAndGas"
				if alliance == SC2APIProtocol.Alliance_Self && len(unit.GetOrders()) == 0 && unit.GetBuildProgress() == 1.0 {
					if obs.PlayerCommon.GetMinerals() >= 400 && unit.GetAssignedHarvesters() > 0 && ((unit.GetIdealHarvesters()/2 <= unit.GetAssignedHarvesters() || unit.GetIdealHarvesters() <= unit.GetAssignedHarvesters()) && AnyUnitHasOrder(rawData.Units, 45, 318) == false && CountUnitsOfType(rawData.Units, SC2APIProtocol.Alliance_Self, 18) == 1) { // TODO: Way to find out cost programmatically?
						target = FindClosestUnit(rawData.Units, unit, SC2APIProtocol.Alliance_Self, 45) // SCV
						if target != nil {
							var abilityID int32 = 318 // "BUILD_COMMANDCENTER"

							offset := float32(20.0)
							rx := float32(*target.Pos.X + rand.Float32()*offset)
							ry := float32(*target.Pos.Y + rand.Float32()*offset)

							a := &SC2APIProtocol.Action{
								ActionRaw: &SC2APIProtocol.ActionRaw{
									Action: &SC2APIProtocol.ActionRaw_UnitCommand{
										UnitCommand: &SC2APIProtocol.ActionRawUnitCommand{
											AbilityId: &abilityID,
											Target: &SC2APIProtocol.ActionRawUnitCommand_TargetWorldSpacePos{
												TargetWorldSpacePos: &SC2APIProtocol.Point2D{
													X: &rx,
													Y: &ry,
												},
											},
											UnitTags: []uint64{target.GetTag()},
										},
									},
								},
							}
							action.Actions = append(action.Actions, a)
							log.Printf("SCV %d building command center at %f,%f", target.GetTag(), rx, ry)
							continue
						}
					}

					if obs.PlayerCommon.GetMinerals() >= 50 && obs.PlayerCommon.GetFoodCap() > obs.PlayerCommon.GetFoodUsed() && unit.GetIdealHarvesters() >= unit.GetAssignedHarvesters() { // TODO: Way to find out cost programmatically?
						var abilityID int32 = 524 // "TRAIN_SCV"
						a := &SC2APIProtocol.Action{
							ActionRaw: &SC2APIProtocol.ActionRaw{
								Action: &SC2APIProtocol.ActionRaw_UnitCommand{
									UnitCommand: &SC2APIProtocol.ActionRawUnitCommand{
										AbilityId: &abilityID,
										UnitTags:  []uint64{unit.GetTag()},
									},
								},
							},
						}
						action.Actions = append(action.Actions, a)
						log.Printf("Command center %d training SCV", unit.GetTag())
						continue
					}
				}
			}

			if unitType == 19 { // Supply depot
				// This is for "CollectMineralsAndGas"
				if alliance == SC2APIProtocol.Alliance_Self && unit.GetBuildProgress() == 1.0 { // TODO: Check available actions instead of map name
					var abilityID int32 = 556 // "MORPH_SUPPLYDEPOT_LOWER"
					a := &SC2APIProtocol.Action{
						ActionRaw: &SC2APIProtocol.ActionRaw{
							Action: &SC2APIProtocol.ActionRaw_UnitCommand{
								UnitCommand: &SC2APIProtocol.ActionRawUnitCommand{
									AbilityId: &abilityID,
									UnitTags:  []uint64{unit.GetTag()},
								},
							},
						},
					}
					action.Actions = append(action.Actions, a)
					log.Printf("Supply depot %d lowering", unit.GetTag())
					continue
				}
			}

			if unitType == 21 { // Barracks
				// This is for "BuildMarines" (or any multiplayer map)
				if alliance == SC2APIProtocol.Alliance_Self && len(unit.GetOrders()) == 0 && unit.GetBuildProgress() == 1.0 {
					if obs.PlayerCommon.GetMinerals() >= 50 && obs.PlayerCommon.GetFoodCap() > obs.PlayerCommon.GetFoodUsed() { // TODO: Way to find out cost programmatically?
						var abilityID int32 = 560 // "TRAIN_MARINE"
						a := &SC2APIProtocol.Action{
							ActionRaw: &SC2APIProtocol.ActionRaw{
								Action: &SC2APIProtocol.ActionRaw_UnitCommand{
									UnitCommand: &SC2APIProtocol.ActionRawUnitCommand{
										AbilityId: &abilityID,
										UnitTags:  []uint64{unit.GetTag()},
									},
								},
							},
						}
						action.Actions = append(action.Actions, a)
						log.Printf("Barracks %d training marine", unit.GetTag())
						continue
					}
				}
			}
		}

		if len(action.Actions) > 0 {
			// Send actions
			err = protocol.SendRequest(req)
			if err != nil {
				log.Fatal("Could not send action request:", err)
			}

			resp, err = protocol.ReadResponse()
			if err != nil {
				log.Fatal("Could not receive action response:", err)
			}

			log.Println(resp)
		}

		// Keep this reasonably paced
		//time.Sleep(100 * time.Millisecond)

		if realtime == false {
			// Prep request for action in case we need it
			req = &SC2APIProtocol.Request{
				Request: &SC2APIProtocol.Request_Step{
					Step: &SC2APIProtocol.RequestStep{
						Count: proto.Uint32(1),
					},
				},
			}

			err = protocol.SendRequest(req)
			if err != nil {
				log.Fatal("Could not send step request:", err)
			}

			resp, err = protocol.ReadResponse()
			if err != nil {
				log.Fatal("Could not receive step response:", err)
			}
		}
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

	resp, err = protocol.ReadResponse()
	if err != nil {
		log.Fatal("Could not receive leave response:", err)
	}
	log.Println(resp)
	log.Println("gg")

	// Disconnect
	log.Println("Disconnecting…")
	err = protocol.Disconnect()
	if err != nil {
		log.Fatal("Error disconnecting:", err)
	}

	log.Println("BYE!")

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

func FindClosestUnit(units []*SC2APIProtocol.Unit, ourUnit *SC2APIProtocol.Unit, desiredAlliance SC2APIProtocol.Alliance, desiredUnitType uint32) *SC2APIProtocol.Unit {
	var closestUnit *SC2APIProtocol.Unit
	var bestDistance float64
	for _, unit := range units {
		if unit.GetUnitType() != desiredUnitType {
			continue
		}

		if unit.GetAlliance() != desiredAlliance {
			continue
		}

		dx := float64(*ourUnit.Pos.X - *unit.Pos.X)
		dy := float64(*ourUnit.Pos.Y - *unit.Pos.Y)
		distance := math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2))

		if distance < bestDistance || bestDistance == 0 {
			bestDistance = distance
			closestUnit = unit
		}
	}
	return closestUnit
}

func FindFarthestUnit(units []*SC2APIProtocol.Unit, ourUnit *SC2APIProtocol.Unit, desiredAlliance SC2APIProtocol.Alliance, desiredUnitType uint32) *SC2APIProtocol.Unit {
	var farthestUnit *SC2APIProtocol.Unit
	var bestDistance float64
	for _, unit := range units {
		if unit.GetUnitType() != desiredUnitType {
			continue
		}

		if unit.GetAlliance() != desiredAlliance {
			continue
		}

		dx := float64(*ourUnit.Pos.X - *unit.Pos.X)
		dy := float64(*ourUnit.Pos.Y - *unit.Pos.Y)
		distance := math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2))

		if distance > bestDistance || bestDistance == 0 {
			bestDistance = distance
			farthestUnit = unit
		}
	}
	return farthestUnit
}

// This only works for non-enemy units
func AnyUnitHasOrder(units []*SC2APIProtocol.Unit, desiredUnitType uint32, desiredabilityID uint32) bool {
	for _, unit := range units {
		if unit.GetUnitType() != desiredUnitType {
			continue
		}

		for _, order := range unit.GetOrders() {
			if order.GetAbilityId() == desiredabilityID {
				return true
			}
		}
	}

	return false
}

func CountUnitsOfType(units []*SC2APIProtocol.Unit, desiredAlliance SC2APIProtocol.Alliance, desiredUnitType uint32) uint32 {
	var count uint32
	for _, unit := range units {
		if unit.GetUnitType() != desiredUnitType {
			continue
		}

		if unit.GetAlliance() != desiredAlliance {
			continue
		}

		count++
	}

	return count
}

func IsUnitTypeAtPoint(units []*SC2APIProtocol.Unit, desiredUnitType uint32, point SC2APIProtocol.Point) bool {
	for _, unit := range units {
		if unit.GetUnitType() != desiredUnitType {
			continue
		}

		unitPos := unit.GetPos()
		if unitPos.GetX() == point.GetX() && unitPos.GetY() == point.GetY() && unitPos.GetZ() == point.GetZ() {
			return true
		}
	}

	return false
}

func HasActionQueued(actions []*SC2APIProtocol.Action, abilityID int32) bool {
	for _, action := range actions {
		if action.ActionRaw.GetUnitCommand().GetAbilityId() == abilityID {
			return true
		}
	}

	return false
}

func AbilityIsAvailable(desiredabilityID uint32) (bool, error) {
	for _, ability := range abilities {
		if ability.GetAbilityId() == desiredabilityID {
			return ability.GetAvailable(), nil
		}
	}

	return false, errors.New("Ability not found")
}

func GetAbilityFootprintRadius(desiredabilityID uint32) (float32, error) {
	for _, ability := range abilities {
		if ability.GetAbilityId() == desiredabilityID {
			return ability.GetFootprintRadius(), nil
		}
	}

	return 0, errors.New("Ability not found")
}

func UnitIsAvailable(desiredUnitType uint32) (bool, error) {
	for _, unit := range units {
		if unit.GetUnitId() == desiredUnitType {
			return unit.GetAvailable(), nil
		}
	}

	return false, errors.New("Unit type not found")
}

func GetUnitMineralCost(desiredUnitType uint32) (uint32, error) {
	for _, unit := range units {
		if unit.GetUnitId() == desiredUnitType {
			return unit.GetMineralCost(), nil
		}
	}

	return 0, errors.New("Unit type not found")
}

func GetUnitVespeneCost(desiredUnitType uint32) (uint32, error) {
	for _, unit := range units {
		if unit.GetUnitId() == desiredUnitType {
			return unit.GetVespeneCost(), nil
		}
	}

	return 0, errors.New("Unit type not found")
}

func FindClosestAnyEnemy(units []*SC2APIProtocol.Unit, ourUnit *SC2APIProtocol.Unit) *SC2APIProtocol.Unit {
	var closestUnit *SC2APIProtocol.Unit
	var bestDistance float64
	for _, unit := range units {
		if unit.GetAlliance() != SC2APIProtocol.Alliance_Enemy {
			continue
		}

		dx := float64(*ourUnit.Pos.X - *unit.Pos.X)
		dy := float64(*ourUnit.Pos.Y - *unit.Pos.Y)
		distance := math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2))

		if distance < bestDistance || bestDistance == 0 {
			bestDistance = distance
			closestUnit = unit
		}
	}
	return closestUnit
}

func ListMaps(sc2 Protocol) {
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

func CreateGame(sc2 Protocol, mapName string, mapPath string) {

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
						Realtime:    proto.Bool(false),
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
						Realtime:    proto.Bool(false),
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
// CollectMineralShards: 109
// CollectMineralsAndGas: 5986
// BuildMarines: 30
// DefeatZerglingsAndBanelings: 67
// DefeatRoaches: 46
// FindAndDefeatZerglings: 10
