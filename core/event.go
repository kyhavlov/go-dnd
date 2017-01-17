package core

import (
	"encoding/gob"
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	log "github.com/Sirupsen/logrus"
	"github.com/kyhavlov/go-dnd/mapgen"
	"github.com/kyhavlov/go-dnd/structs"
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
	gob.Register(&PlayerAction{})
	gob.Register(&PlayerReady{})
	gob.Register(&TurnChange{})
	gob.Register(&Move{})
	gob.Register(&Attack{})
	gob.Register(&PickupItem{})
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
		case *MoveSystem:
			sys.CreatureLocations = make([][]*structs.Creature, level.Width)
			for i, _ := range sys.CreatureLocations {
				sys.CreatureLocations[i] = make([]*structs.Creature, level.Height)
			}
			sys.ItemLocations = make([][]*structs.Item, level.Width)
			for i, _ := range sys.ItemLocations {
				sys.ItemLocations[i] = make([]*structs.Item, level.Height)
			}
		}
	}
	for _, tile := range level.Tiles {
		AddTile(w, tile)
	}
	for _, creature := range level.Creatures {
		AddCreature(w, creature)
	}

	// Make a test item
	sheet := common.NewSpritesheetFromFile(structs.SpritesheetPath, structs.TileWidth, structs.TileWidth)
	if sheet == nil {
		log.Fatalf("Unable to load texture file")
	}
	coords := structs.GridPoint{
		X: 7,
		Y: 3,
	}
	item := structs.Item{
		Life:     20,
		OnGround: true,
	}
	item.BasicEntity = ecs.NewBasic()
	item.SpaceComponent = common.SpaceComponent{
		Position: coords.ToPixels(),
		Width:    structs.TileWidth,
		Height:   structs.TileWidth,
	}
	item.RenderComponent = common.RenderComponent{
		Drawable: sheet.Cell(1378),
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *NetworkSystem:
			item.NetworkID = sys.nextId()
		}
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&item.BasicEntity, &item.RenderComponent, &item.SpaceComponent)
		case *MoveSystem:
			sys.AddItem(&item)
		}
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
		BasicEntity:  ecs.NewBasic(),
		IsPlayerTeam: true,
	}
	player.HealthComponent = structs.HealthComponent{
		MaxLife: 50,
	}
	player.SpaceComponent = common.SpaceComponent{
		Position: event.GridPoint.ToPixels(),
		Width:    structs.TileWidth,
		Height:   structs.TileWidth,
	}
	player.RenderComponent = common.RenderComponent{
		Drawable: sheet.Cell(594 + int(event.PlayerID)),
	}
	player.RenderComponent.SetZIndex(1)

	AddCreature(w, &player)

	for _, system := range w.Systems() {
		switch sys := system.(type) {
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
	var move *MoveSystem
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *TurnSystem:
			//log.Debugf("Setting action %v for player %d", reflect.TypeOf(p.Action), p.PlayerID)
			sys.PlayerActions[p.PlayerID] = &p.Action
		case *MoveSystem:
			move = sys
		}
	}
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *UiSystem:
			switch action := p.Action.(type) {
			case *Move:
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
					offset := -4 + float32(p.PlayerID*5)
					line.SpaceComponent = common.SpaceComponent{Position: engo.Point{start.X + structs.TileWidth/2 + offset, start.Y + structs.TileWidth/2 + offset}, Width: w, Height: h}
					line.RenderComponent = common.RenderComponent{Drawable: common.Rectangle{}, Color: color.RGBA{0, 255, 0 + uint8(p.PlayerID*255), 255}}
					lines = append(lines, &line)
				}
				sys.UpdateActionIndicator(p.PlayerID, lines)
			case *Attack:
				circle := &UiElement{BasicEntity: ecs.NewBasic()}
				circle.SpaceComponent = common.SpaceComponent{Position: move.Creatures[action.TargetId].Position, Width: structs.TileWidth, Height: structs.TileWidth}
				circle.RenderComponent = common.RenderComponent{Drawable: common.Circle{BorderWidth: 3, BorderColor: color.RGBA{255, 0, 0, 255}}, Color: color.Transparent}
				sys.UpdateActionIndicator(p.PlayerID, []*UiElement{circle})
			case *PickupItem:
				itemCircle := &UiElement{BasicEntity: ecs.NewBasic()}
				itemCircle.SpaceComponent = common.SpaceComponent{Position: move.Items[action.ItemId].Position, Width: structs.TileWidth, Height: structs.TileWidth}
				itemCircle.RenderComponent = common.RenderComponent{Drawable: common.Circle{BorderWidth: 3, BorderColor: color.RGBA{0, 255, 0, 255}}, Color: color.Transparent}
				creatureCircle := &UiElement{BasicEntity: ecs.NewBasic()}
				creatureCircle.SpaceComponent = common.SpaceComponent{Position: move.Creatures[action.CreatureId].Position, Width: structs.TileWidth, Height: structs.TileWidth}
				creatureCircle.RenderComponent = common.RenderComponent{Drawable: common.Circle{BorderWidth: 3, BorderColor: color.RGBA{0, 255, 0, 255}}, Color: color.Transparent}
				sys.UpdateActionIndicator(p.PlayerID, []*UiElement{itemCircle, creatureCircle})
			}
		}
	}
	return true
}

type PlayerReady struct {
	PlayerID
}

func (p *PlayerReady) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *TurnSystem:
			//log.Debugf("Setting ready to %v for player %d", p.Ready, p.PlayerID)
			sys.PlayerReady[p.PlayerID] = !sys.PlayerReady[p.PlayerID]
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

type PickupItem struct {
	ItemId     structs.NetworkID
	CreatureId structs.NetworkID
}

func (p *PickupItem) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MoveSystem:
			item := sys.Items[p.ItemId]
			creature := sys.Creatures[p.CreatureId]
			// put the item in the first empty inventory slot
			for i, slot := range creature.Inventory {
				if slot == nil {
					creature.Inventory[i] = item
					break
				}
			}
			item.RenderComponent.Hidden = true
			item.OnGround = false
		case *UiSystem:
			if p.CreatureId == sys.input.player.NetworkID {
				sys.UpdateInventoryDisplay()
			}
		}
	}
	return true
}
