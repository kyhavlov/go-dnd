package main

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	log "github.com/Sirupsen/logrus"
)

type MouseTracker struct {
	ecs.BasicEntity
	common.MouseComponent
	common.SpaceComponent
}

type InputSystem struct {
	mouseTracker MouseTracker
	player       *Player
	outgoing     chan NetworkMessage
	PlayerID
}

// Sets the PlayerID of the local InputSystem, so we know which player we are and what we control
type SetPlayerAction struct {
	PlayerID
}

func (set *SetPlayerAction) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *InputSystem:
			sys.PlayerID = set.PlayerID
			log.Info("player position set to: ", set.PlayerID)
		}
	}

	return true
}

// Moves the entity with the given ID to NewLocation
type MoveAction struct {
	Id          NetworkID
	NewLocation engo.Point
}

func (move *MoveAction) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MoveSystem:
			sys.SpaceComponents[move.Id].Position.X = move.NewLocation.X
			sys.SpaceComponents[move.Id].Position.Y = move.NewLocation.Y
			log.Info("player position set to: ", sys.SpaceComponents[move.Id].Position)
		}
	}

	return true
}

// New is the initialisation of the System
func (input *InputSystem) New(w *ecs.World) {
	input.mouseTracker.BasicEntity = ecs.NewBasic()
	input.mouseTracker.MouseComponent = common.MouseComponent{Track: true}
	input.mouseTracker.SpaceComponent = common.SpaceComponent{}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.MouseSystem:
			sys.Add(&input.mouseTracker.BasicEntity, &input.mouseTracker.MouseComponent, &input.mouseTracker.SpaceComponent, nil)
		}
	}
}

func (input *InputSystem) Update(dt float32) {
	if input.mouseTracker.MouseComponent.Clicked && input.player != nil {
		newLocation := engo.Point{
			float32(int(input.mouseTracker.MouseComponent.MouseX/TileWidth) * TileWidth),
			float32(int(input.mouseTracker.MouseComponent.MouseY/TileWidth) * TileWidth),
		}

		move := &MoveAction{input.player.NetworkID, newLocation}
		go SendMessage(input.outgoing, NetworkMessage{
			Actions: []Action{move},
		})

		log.Info("Sent move action to input.outgoing")
	}
}

func (*InputSystem) Remove(ecs.BasicEntity) {}
