package main

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	log "github.com/Sirupsen/logrus"
)

// Spawns a player with the given ID at the given GridPoint
type NewPlayerEvent struct {
	PlayerID
	Life int
	GridPoint
}

func (event *NewPlayerEvent) Process(w *ecs.World, dt float32) bool {
	sheet := common.NewSpritesheetFromFile(SpritesheetPath, TileWidth, TileWidth)

	player := Creature{
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
		case *LightSystem:
			sys.Add(&player.BasicEntity, &DynamicLightSource{
				spaceComponent: &player.SpaceComponent,
				Brightness:     250,
			})
		case *TurnSystem:
			sys.PlayerReady[event.PlayerID] = false
		}

	}

	log.Infof("New player added at %v, ID: %d", event.GridPoint, event.PlayerID)

	return true
}
