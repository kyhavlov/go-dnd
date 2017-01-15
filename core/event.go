package core

import (
	"engo.io/ecs"
	log "github.com/Sirupsen/logrus"
	"github.com/kyhavlov/go-dnd/structs"
	"engo.io/engo/common"
	"engo.io/engo"
	"encoding/gob"
)

// Event is the interface for things which affect world state, such as
// creature position, life, items, etc. It takes a time delta and returns
// whether it has completed.
type Event interface {
	Process(*ecs.World, float32) bool
}

// Register our different Event implementations with the gob package for serialization
// it's probably worth trying to keep the number of different events low for simplicity
func RegisterEvents() {
	gob.Register(&GameStartEvent{})
	gob.Register(&SetPlayerEvent{})
	gob.Register(&NewPlayerEvent{})
	gob.Register(&MoveEvent{})
	gob.Register(&PlayerAction{})
	gob.Register(&PlayerReady{})
	gob.Register(&TurnChangeEvent{})

}

type GameStartEvent struct {
	RandomSeed  int64
	PlayerCount int
}

func (gs GameStartEvent) Process(w *ecs.World, dt float32) bool {
	log.Infof("Got random seed from server: %d", gs.RandomSeed)
	GenerateMap(w, gs.RandomSeed)
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *UiSystem:
			sys.InitUI(w, gs.PlayerCount)
		case *TurnSystem:
			for i := 0; i < gs.PlayerCount; i++ {
				sys.PlayerReady[PlayerID(i)] = false
			}
		}
	}
	return true
}

// Spawns a player with the given ID at the given GridPoint
type NewPlayerEvent struct {
	PlayerID
	Life int
	GridPoint
}

func (event *NewPlayerEvent) Process(w *ecs.World, dt float32) bool {
	sheet := common.NewSpritesheetFromFile(SpritesheetPath, TileWidth, TileWidth)

	player := structs.Creature{
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