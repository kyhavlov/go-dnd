package main

import (
	"engo.io/ecs"
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

	player *Creature
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

		if input.mapSystem.GetTileAt(gridPoint) != nil {
			start := input.mapSystem.GetTileAt(PointToGridPoint(input.player.SpaceComponent.Position))
			path := GetPath(start, input.mapSystem.GetTileAt(gridPoint), input.mapSystem.Tiles)

			if len(path) < 17 {
				input.outgoing <- NetworkMessage{
					Events: []Event{&MoveEvent{
						Id:   input.player.NetworkID,
						Path: path,
					}},
				}
			} else {
				log.Info("Tried to move too far")
			}
		}
	}
}

func (*InputSystem) Remove(ecs.BasicEntity) {}
