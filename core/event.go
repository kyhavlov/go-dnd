package core

import (
	"engo.io/ecs"
	log "github.com/Sirupsen/logrus"
	"github.com/kyhavlov/go-dnd/structs"
	"engo.io/engo/common"
	"engo.io/engo"
	"encoding/gob"
	"github.com/kyhavlov/go-dnd/mapgen"
	"image/color"
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
	gob.Register(&GameStart{})
	gob.Register(&SetPlayerID{})
	gob.Register(&NewPlayer{})
	gob.Register(&MoveEvent{})
	gob.Register(&PlayerAction{})
	gob.Register(&PlayerReady{})
	gob.Register(&TurnChange{})

}

// Starts the game, generating the map from the given seed
type GameStart struct {
	RandomSeed  int64
	PlayerCount int
}

func (gs GameStart) Process(w *ecs.World, dt float32) bool {
	log.Infof("Got random seed from server: %d", gs.RandomSeed)
	level := mapgen.GenerateMap(gs.RandomSeed)
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *UiSystem:
			sys.InitUI(w, gs.PlayerCount)
		case *TurnSystem:
			for i := 0; i < gs.PlayerCount; i++ {
				sys.PlayerReady[PlayerID(i)] = false
			}
		case *MapSystem:
			sys.Tiles = make([][]*structs.Tile, level.Width)
			for i, _ := range sys.Tiles {
				sys.Tiles[i] = make([]*structs.Tile, level.Height)
			}
		}
	}
	for _, tile := range level.Tiles {
		AddTile(w, tile)
	}
	for _, creature := range level.Creatures {
		AddCreature(w, creature)
	}
	return true
}

// Sets the PlayerID of the local InputSystem, so we know which player we are and what we control
type SetPlayerID struct {
	PlayerID
}

func (event *SetPlayerID) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *InputSystem:
			sys.PlayerID = event.PlayerID
			log.Info("player ID set to: ", event.PlayerID)
		}
	}

	return true
}

// Spawns a player with the given ID at the given GridPoint
type NewPlayer struct {
	PlayerID
	Life int
	structs.GridPoint
}

func (event *NewPlayer) Process(w *ecs.World, dt float32) bool {
	sheet := common.NewSpritesheetFromFile(structs.SpritesheetPath, structs.TileWidth, structs.TileWidth)

	player := structs.Creature{
		BasicEntity: ecs.NewBasic(),
	}
	player.SpaceComponent = common.SpaceComponent{
		Position: event.GridPoint.ToPixels(),
		Width:    structs.TileWidth,
		Height:   structs.TileWidth,
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

type PlayerAction struct {
	PlayerID PlayerID
	Action   Event
}

func (p *PlayerAction) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *TurnSystem:
			//log.Debugf("Setting action %v for player %d", reflect.TypeOf(p.Action), p.PlayerID)
			sys.PlayerActions[p.PlayerID] = &p.Action
		}
	}
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *UiSystem:
			switch action := p.Action.(type) {
			case *MoveEvent:
				var lines []*UiElement
				for i := 0; i < len(action.Path)-1; i++ {
					line := UiElement{BasicEntity: ecs.NewBasic()}
					current := action.Path[i].ToPixels()
					next := action.Path[i+1].ToPixels()
					start := current
					if current.X > next.X || current.Y > next.Y {
						start = next
					}
					w := float32(3)
					h := float32(3)
					if current.X != next.X {
						w = structs.TileWidth
					} else {
						h = structs.TileWidth
					}
					line.SpaceComponent = common.SpaceComponent{Position: engo.Point{start.X + structs.TileWidth/2, start.Y + structs.TileWidth/2}, Width: w, Height: h}
					line.RenderComponent = common.RenderComponent{Drawable: common.Rectangle{}, Color: color.RGBA{0, 255, 0, 255}}
					lines = append(lines, &line)
				}
				sys.UpdateActionIndicator(p.PlayerID, lines)
			}
		}
	}
	return true
}

type PlayerReady struct {
	PlayerID
	Ready bool
}

func (p *PlayerReady) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *TurnSystem:
			//log.Debugf("Setting ready to %v for player %d", p.Ready, p.PlayerID)
			sys.PlayerReady[p.PlayerID] = p.Ready
		}
	}
	return true
}

type TurnChange struct {
	PlayersTurn bool
}

func (t *TurnChange) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *TurnSystem:
			log.Infof("Setting turn back to players")
			sys.PlayersTurn = t.PlayersTurn
		}
	}
	return true
}