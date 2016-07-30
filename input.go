package main

import (
	"engo.io/ecs"
	"engo.io/engo/common"
	"fmt"
)

type MouseTracker struct {
	ecs.BasicEntity
	common.MouseComponent
	common.SpaceComponent
}

type InputSystem struct {
	mouseTracker MouseTracker
	player *Player
}

// New is the initialisation of the System
func (cb *InputSystem) New(w *ecs.World) {
	cb.mouseTracker.BasicEntity = ecs.NewBasic()
	cb.mouseTracker.MouseComponent = common.MouseComponent{Track: true}
	cb.mouseTracker.SpaceComponent = common.SpaceComponent{}

	fmt.Println(cb.player.SpaceComponent.Position)

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.MouseSystem:
			sys.Add(&cb.mouseTracker.BasicEntity, &cb.mouseTracker.MouseComponent, &cb.mouseTracker.SpaceComponent, nil)
		}
	}
}

func (cb *InputSystem) Update(dt float32) {
	if cb.mouseTracker.MouseComponent.Clicked {
		cb.player.Position.X = float32(int(cb.mouseTracker.MouseComponent.MouseX/TileWidth)*TileWidth)
		cb.player.Position.Y = float32(int(cb.mouseTracker.MouseComponent.MouseY/TileWidth)*TileWidth)
		fmt.Println("player position set to: ", cb.player.Position)
	}
}

func (*InputSystem) Remove(ecs.BasicEntity) {}