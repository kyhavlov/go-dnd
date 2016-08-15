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
	mapSystem    *MapSystem

	player *Player
	PlayerID

	outgoing chan NetworkMessage
}

// Sets the PlayerID of the local InputSystem, so we know which player we are and what we control
type SetPlayerEvent struct {
	PlayerID
}

func (event *SetPlayerEvent) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *InputSystem:
			sys.PlayerID = event.PlayerID
			log.Info("player ID set to: ", event.PlayerID)
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
		gridPoint := GridPoint{
			X: int(input.mouseTracker.MouseComponent.MouseX / TileWidth),
			Y: int(input.mouseTracker.MouseComponent.MouseY / TileWidth),
		}
		newLocation := engo.Point{
			float32(gridPoint.X * TileWidth),
			float32(gridPoint.Y * TileWidth),
		}

		if input.mapSystem.GetTileAt(gridPoint) != nil {
			move := &MoveEvent{input.player.NetworkID, newLocation}
			SendMessage(input.outgoing, NetworkMessage{
				Events: []Event{move},
			})

			log.Info("Sent move event to input.outgoing")
		}
	}
}

func (*InputSystem) Remove(ecs.BasicEntity) {}
