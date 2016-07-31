package main

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	log "github.com/Sirupsen/logrus"
)

type Player struct {
	NetworkID
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

type NewPlayerAction struct {
	PlayerID
	Location GridPoint
}

func (action *NewPlayerAction) Process(w *ecs.World, dt float32) bool {
	sheet := common.NewSpritesheetFromFile("textures/dungeon2x.png", TileWidth, TileWidth)

	player := Player{
		BasicEntity: ecs.NewBasic(),
	}
	player.SpaceComponent = common.SpaceComponent{
		Position: action.Location.toPixels(),
		Width:    TileWidth,
		Height:   TileWidth,
	}
	player.RenderComponent = common.RenderComponent{
		Drawable: sheet.Cell(594 + int(action.PlayerID)),
		Scale:    engo.Point{1, 1},
	}
	player.RenderComponent.SetZIndex(100)

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *NetworkSystem:
			player.NetworkID = sys.nextId()
		}
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&player.BasicEntity, &player.RenderComponent, &player.SpaceComponent)
		case *MoveSystem:
			sys.Add(&player.BasicEntity, &player.SpaceComponent, player.NetworkID)
		case *InputSystem:
			if sys.PlayerID == action.PlayerID {
				sys.player = &player
			}
		}
	}

	log.Info("New player added at ", action.Location)

	return true
}
