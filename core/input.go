package core

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	log "github.com/Sirupsen/logrus"
	"github.com/kyhavlov/go-dnd/structs"
)

type MouseTracker struct {
	ecs.BasicEntity
	common.MouseComponent
	common.SpaceComponent
}

type InputSystem struct {
	mouseTracker MouseTracker
	mapSystem    *MapSystem
	turn         *TurnSystem
	move         *MoveSystem

	player *structs.Creature
	PlayerID

	outgoing chan NetworkMessage
}

const ReadyKey = "ready"

// New is the initialisation of the System
func (input *InputSystem) New(w *ecs.World) {
	input.mouseTracker.BasicEntity = ecs.NewBasic()
	input.mouseTracker.MouseComponent = common.MouseComponent{Track: true}
	input.mouseTracker.SpaceComponent = common.SpaceComponent{}

	engo.Input.RegisterButton(ReadyKey, engo.T)

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.MouseSystem:
			sys.Add(&input.mouseTracker.BasicEntity, &input.mouseTracker.MouseComponent, &input.mouseTracker.SpaceComponent, nil)
		case *TurnSystem:
			input.turn = sys
		}
	}
}

func (input *InputSystem) Update(dt float32) {
	// One of three things can happen on left click: move, attack or pick up item
	if input.mouseTracker.MouseComponent.Clicked && input.player != nil && input.turn.PlayersTurn && !input.turn.PlayerReady[input.PlayerID] {
		gridPoint := structs.GridPoint{
			X: int(input.mouseTracker.MouseComponent.MouseX / structs.TileWidth),
			Y: int(input.mouseTracker.MouseComponent.MouseY / structs.TileWidth),
		}

		// If the target is occupied by an enemy, try to attack
		if input.mapSystem.GetTileAt(gridPoint) != nil {
			if target := input.move.GetCreatureAt(gridPoint); target != nil && !target.IsPlayerTeam {
				input.outgoing <- NetworkMessage{
					Events: []Event{&PlayerAction{
						PlayerID: input.PlayerID,
						Action: &Attack{
							Id:       input.player.NetworkID,
							TargetId: target.NetworkID,
						},
					}},
				}
			} else {
				if item := input.move.GetItemAt(gridPoint); item != nil && item.OnGround {
					input.outgoing <- NetworkMessage{
						Events: []Event{&PlayerAction{
							PlayerID: input.PlayerID,
							Action: &PickupItem{
								ItemId:     item.NetworkID,
								CreatureId: input.player.NetworkID,
							},
						}},
					}
				} else {
					start := input.mapSystem.GetTileAt(structs.PointToGridPoint(input.player.SpaceComponent.Position))
					path := GetPath(start, input.mapSystem.GetTileAt(gridPoint), input.mapSystem.Tiles, input.move.CreatureLocations, true)

					if len(path) < 17 {
						input.outgoing <- NetworkMessage{
							Events: []Event{&PlayerAction{
								PlayerID: input.PlayerID,
								Action: &Move{
									Id:   input.player.NetworkID,
									Path: path,
								},
							}},
						}
					} else {
						log.Info("Tried to move too far")
					}
				}
			}
		}
	}

	if engo.Input.Button(ReadyKey).JustPressed() && input.turn.PlayersTurn {
		input.outgoing <- NetworkMessage{
			Events: []Event{&PlayerReady{
				PlayerID: input.PlayerID,
			}},
		}
	}
}

func (*InputSystem) Remove(ecs.BasicEntity) {}
