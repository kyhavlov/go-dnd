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

	engo.Input.RegisterButton(ReadyKey, engo.R)

	engo.Input.RegisterButton(string(EquipmentHotkeys[0]), engo.G)
	engo.Input.RegisterButton(string(EquipmentHotkeys[1]), engo.H)
	engo.Input.RegisterButton(string(EquipmentHotkeys[2]), engo.J)
	engo.Input.RegisterButton(string(EquipmentHotkeys[3]), engo.K)
	engo.Input.RegisterButton(string(EquipmentHotkeys[4]), engo.L)

	engo.Input.RegisterButton(string(InventoryHotkeys[0]), engo.Z)
	engo.Input.RegisterButton(string(InventoryHotkeys[1]), engo.X)
	engo.Input.RegisterButton(string(InventoryHotkeys[2]), engo.C)
	engo.Input.RegisterButton(string(InventoryHotkeys[3]), engo.V)
	engo.Input.RegisterButton(string(InventoryHotkeys[4]), engo.B)

	engo.Input.RegisterButton(string(SkillHotkeys[0]), engo.One)
	engo.Input.RegisterButton(string(SkillHotkeys[1]), engo.Two)
	engo.Input.RegisterButton(string(SkillHotkeys[2]), engo.Three)
	engo.Input.RegisterButton(string(SkillHotkeys[3]), engo.Four)
	engo.Input.RegisterButton(string(SkillHotkeys[4]), engo.Five)
	engo.Input.RegisterButton(string(SkillHotkeys[5]), engo.Six)
	engo.Input.RegisterButton(string(SkillHotkeys[6]), engo.Seven)
	engo.Input.RegisterButton(string(SkillHotkeys[7]), engo.Eight)
	engo.Input.RegisterButton(string(SkillHotkeys[8]), engo.Nine)
	engo.Input.RegisterButton(string(SkillHotkeys[9]), engo.Zero)

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
			if item := input.mapSystem.GetItemAt(gridPoint); item != nil && item.OnGround {
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
				path := GetPath(start, input.mapSystem.GetTileAt(gridPoint), input.mapSystem.Tiles, input.mapSystem.CreatureLocations, true)

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

	if engo.Input.Button(ReadyKey).JustPressed() && input.turn.PlayersTurn {
		input.outgoing <- NetworkMessage{
			Events: []Event{&PlayerReady{
				PlayerID: input.PlayerID,
			}},
		}
	}

	for i := 0; i < structs.InventorySize; i++ {
		if engo.Input.Button(string(InventoryHotkeys[i])).JustPressed() && input.turn.PlayersTurn {
			if input.player.Inventory[i] != nil {
				input.outgoing <- NetworkMessage{
					Events: []Event{&PlayerAction{
						PlayerID: input.PlayerID,
						Action: &EquipItem{
							InventorySlot: i,
							CreatureId:    input.player.NetworkID,
						}},
					},
				}
			}
		}
	}

	for i := 0; i < structs.EquipmentSlots; i++ {
		if engo.Input.Button(string(EquipmentHotkeys[i])).JustPressed() && input.turn.PlayersTurn {
			if input.player.Equipment[i] != nil {
				input.outgoing <- NetworkMessage{
					Events: []Event{&PlayerAction{
						PlayerID: input.PlayerID,
						Action: &UnequipItem{
							EquipSlot:  i,
							CreatureId: input.player.NetworkID,
						}},
					},
				}
			}
		}
	}

	for i := 0; i < structs.SkillSlots; i++ {
		if engo.Input.Button(string(SkillHotkeys[i])).JustPressed() && input.turn.PlayersTurn {
			gridPoint := structs.GridPoint{
				X: int(input.mouseTracker.MouseComponent.MouseX / structs.TileWidth),
				Y: int(input.mouseTracker.MouseComponent.MouseY / structs.TileWidth),
			}

			skills := GetCreatureSkills(input.player)
			if target := input.mapSystem.GetCreatureAt(gridPoint); target != nil && len(skills) > i {
				skill := NewSkill(skills[i], input.player.NetworkID, target.NetworkID)
				if skill.TargetIsValid(input.mapSystem) {
					input.outgoing <- NetworkMessage{
						Events: []Event{&PlayerAction{
							PlayerID: input.PlayerID,
							Action:   &UseSkill{skill},
						}},
					}
				} else {
					log.Info("Tried to attack invalid target")
				}
			}
		}
	}
}

func (*InputSystem) Remove(ecs.BasicEntity) {}
