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

// Spawns a player with the given ID at the given GridPoint
type NewPlayerEvent struct {
	PlayerID
	GridPoint
}

func (event *NewPlayerEvent) Process(w *ecs.World, dt float32) bool {
	sheet := common.NewSpritesheetFromFile(SpritesheetPath, TileWidth, TileWidth)

	player := Player{
		BasicEntity: ecs.NewBasic(),
	}
	player.SpaceComponent = common.SpaceComponent{
		Position: event.GridPoint.toPixels(),
		Width:    TileWidth,
		Height:   TileWidth,
	}
	player.RenderComponent = common.RenderComponent{
		Drawable: sheet.Cell(594 + int(event.PlayerID)),
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
			if sys.PlayerID == event.PlayerID {
				sys.player = &player
			}
		}
	}

	log.Info("New player added at ", event.GridPoint)

	return true
}
