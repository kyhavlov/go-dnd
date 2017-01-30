package core

import (
	"encoding/gob"
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/engoengine/math"
	"github.com/kyhavlov/go-dnd/mapgen"
	"github.com/kyhavlov/go-dnd/structs"
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
	gob.Register(&ResetPlayerActions{})
	gob.Register(&PlayerReady{})
	gob.Register(&TurnChange{})
	gob.Register(&Move{})
	gob.Register(&UseSkill{})
	gob.Register(&PickupItem{})
	gob.Register(&EquipItem{})
	gob.Register(&UnequipItem{})
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
			sys.MapInfo = level
			sys.Tiles = make([][]*structs.Tile, level.Width)
			for i, _ := range sys.Tiles {
				sys.Tiles[i] = make([]*structs.Tile, level.Height)
			}
			sys.CreatureLocations = make([][]*structs.Creature, level.Width)
			for i, _ := range sys.CreatureLocations {
				sys.CreatureLocations[i] = make([]*structs.Creature, level.Height)
			}
			sys.ItemLocations = make([][][]*structs.Item, level.Width)
			for i, _ := range sys.ItemLocations {
				sys.ItemLocations[i] = make([][]*structs.Item, level.Height)
				for j, _ := range sys.ItemLocations[i] {
					sys.ItemLocations[i][j] = make([]*structs.Item, 0)
				}
			}
		}
	}

	for _, tile := range level.Tiles {
		AddTile(w, tile)
	}
	for _, creature := range level.Creatures {
		AddCreature(w, creature)
	}

	// Make some test items
	staff := structs.NewItem("Sapphire Staff", structs.GridPoint{
		X: level.StartLoc.X + 1,
		Y: level.StartLoc.Y + 2,
	})
	staff.OnGround = true
	AddItem(w, staff)

	armor := structs.NewItem("Leather Armor", structs.GridPoint{
		X: level.StartLoc.X + 1,
		Y: level.StartLoc.Y + 2,
	})
	armor.OnGround = true
	AddItem(w, armor)

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
}

func (event *NewPlayer) Process(w *ecs.World, dt float32) bool {
	sheet := common.NewSpritesheetFromFile(structs.SpritesheetPath, structs.TileWidth, structs.TileWidth)

	var spawnLoc structs.GridPoint
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MapSystem:
			spawnLoc = sys.MapInfo.StartLoc
		}
	}
	spawnLoc.X += int(event.PlayerID)
	spawnLoc.Y += 3

	player := structs.NewCreature("Player", spawnLoc)
	player.IsPlayerTeam = true
	player.RenderComponent = common.RenderComponent{
		Drawable: sheet.Cell(player.Icon + int(event.PlayerID)),
	}
	AddCreature(w, player)

	isLocalPlayer := false
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MapSystem:
			sys.Players[event.PlayerID] = player
		case *InputSystem:
			if sys.PlayerID == event.PlayerID {
				sys.player = player
				isLocalPlayer = true
			}
		case *LightSystem:
			sys.Add(&player.BasicEntity, &DynamicLightSource{
				spaceComponent: &player.SpaceComponent,
				Brightness:     250,
			})
		case *TurnSystem:
			sys.PlayerReady[event.PlayerID] = false
		case *UiSystem:
			if isLocalPlayer {
				sys.UpdatePlayerDisplay()
				sys.SetupStatsDisplay(w)
			}
		}
	}

	// Start the camera on this player if it's ours
	if isLocalPlayer {
		engo.Mailbox.Dispatch(common.CameraMessage{Axis: common.XAxis, Value: player.SpaceComponent.Position.X, Incremental: false})
		engo.Mailbox.Dispatch(common.CameraMessage{Axis: common.YAxis, Value: player.SpaceComponent.Position.Y, Incremental: false})
	}

	log.Infof("New player added at %v, ID: %d", spawnLoc, event.PlayerID)

	return true
}

type PlayerAction struct {
	PlayerID PlayerID
	Action   Event
}

func (p *PlayerAction) Process(w *ecs.World, dt float32) bool {
	var mapSystem *MapSystem
	var ui *UiSystem

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *UiSystem:
			ui = sys
		}
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *TurnSystem:
			// Check that the action type is valid based on what they already have locked in
			switch p.Action.(type) {
			case *Move:
				if !sys.PlayerHasMove(p.PlayerID) {
					ui.ResetActionIndicators(p.PlayerID)
					switch sys.PlayerActions[p.PlayerID][0].(type) {
					case *Move:
						sys.PlayerActions[p.PlayerID][0] = p.Action
						sys.PlayerActions[p.PlayerID] = sys.PlayerActions[p.PlayerID][:1]
					default:
						sys.PlayerActions[p.PlayerID][1] = p.Action
					}
					for _, action := range sys.PlayerActions[p.PlayerID] {
						ui.AddActionIndicator(action, p.PlayerID, mapSystem)
					}
				} else {
					sys.PlayerActions[p.PlayerID] = append(sys.PlayerActions[p.PlayerID], p.Action)
					ui.AddActionIndicator(p.Action, p.PlayerID, mapSystem)
				}
			default:
				if (sys.PlayerHasMove(p.PlayerID) && len(sys.PlayerActions[p.PlayerID]) == 1) || len(sys.PlayerActions[p.PlayerID]) == 2 {
					ui.ResetActionIndicators(p.PlayerID)
					switch sys.PlayerActions[p.PlayerID][0].(type) {
					case *Move:
						sys.PlayerActions[p.PlayerID][1] = p.Action
					default:
						sys.PlayerActions[p.PlayerID][0] = p.Action
					}
					for _, action := range sys.PlayerActions[p.PlayerID] {
						ui.AddActionIndicator(action, p.PlayerID, mapSystem)
					}
				} else {
					sys.PlayerActions[p.PlayerID] = append(sys.PlayerActions[p.PlayerID], p.Action)
					ui.AddActionIndicator(p.Action, p.PlayerID, mapSystem)
				}
			}
		case *MapSystem:
			mapSystem = sys
		}
	}

	return true
}

type ResetPlayerActions struct {
	PlayerID PlayerID
}

func (e *ResetPlayerActions) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *TurnSystem:
			sys.PlayerActions[e.PlayerID] = nil
			sys.PlayerReady[e.PlayerID] = false
		}
	}
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *UiSystem:
			sys.ResetActionIndicators(e.PlayerID)
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
			sys.PlayerReady[p.PlayerID] = !sys.PlayerReady[p.PlayerID]
			if _, ok := sys.PlayerActions[p.PlayerID]; !ok {
				sys.PlayerActions[p.PlayerID] = nil
			}
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
		case *MapSystem:
			for _, creature := range sys.Creatures {
				if creature.IsPlayerTeam == t.PlayersTurn {
					creature.Stamina += creature.StaminaRegen
					if creature.Stamina > creature.MaxStamina {
						creature.Stamina = creature.MaxStamina
					}
				}
			}
		}
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *TurnSystem:
			if t.PlayersTurn {
				log.Info("Beginning player turn")
			} else {
				log.Info("Beginning enemy turn")
			}

			sys.PlayersTurn = t.PlayersTurn
		}
	}
	return true
}

type NamedEvent interface {
	Name() string
}

// Moves the entity with the given Id along the path
type Move struct {
	Id   structs.NetworkID
	Path []structs.GridPoint

	// Tracks the node of the path we're currently moving towards
	// private field so we don't send it over the network
	current int
}

// Pixels per frame to move entities
const speed = 3

func (move *Move) Name() string { return "Moving" }
func (move *Move) Process(w *ecs.World, dt float32) bool {
	var lights *LightSystem
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *LightSystem:
			lights = sys
		}
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MapSystem:
			// Check if the path needs to be ended early because of an occupying creature
			last := 0
			for i := len(move.Path) - 1; i > 0; i-- {
				if sys.GetCreatureAt(move.Path[i]) == nil {
					last = i
					break
				}
			}
			if last == 0 {
				return true
			}
			move.Path = move.Path[:last+1]

			current := &sys.SpaceComponents[move.Id].Position
			target := move.Path[move.current].ToPixels()
			if current.PointDistance(target) <= 3.0 {
				lights.needsUpdate = true
				current.X = target.X
				current.Y = target.Y
				move.current++
				if move.current == len(move.Path) {
					// Move the mapping to the new location
					creature, ok := sys.Creatures[move.Id]
					if ok {
						sys.CreatureLocations[move.Path[0].X][move.Path[0].Y] = nil
						sys.CreatureLocations[move.Path[move.current-1].X][move.Path[move.current-1].Y] = creature
					}

					return true
				}
			} else {
				xDiff, yDiff := math.Abs(current.X-target.X), math.Abs(current.Y-target.Y)
				if xDiff != 0 {
					directionX := (current.X - target.X) / xDiff
					current.X -= speed * directionX
				}
				if yDiff != 0 {
					directionY := (current.Y - target.Y) / yDiff
					current.Y -= speed * directionY
				}
			}
		}
	}

	return false
}

type UseSkill struct {
	SkillName string
	Source    structs.NetworkID
	Target    structs.NetworkID
}

func (e *UseSkill) Name() string { return fmt.Sprintf("Using skill: %s", e.SkillName) }
func (e *UseSkill) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MapSystem:
			PerformSkillActions(e.SkillName, sys, e.Source, e.Target)
		}
	}

	return true
}

type PickupItem struct {
	ItemId     structs.NetworkID
	CreatureId structs.NetworkID
}

func (p *PickupItem) Name() string { return "Picking up item" }
func (p *PickupItem) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MapSystem:
			item := sys.Items[p.ItemId]
			// If the item is already picked up, exit early
			if !item.OnGround {
				return true
			}
			creature := sys.Creatures[p.CreatureId]
			// put the item in the first empty inventory slot
			for i, slot := range creature.Inventory {
				if slot == nil {
					creature.Inventory[i] = item
					break
				}
			}
			log.Infof("Inventory: %v", creature.Inventory)
			itemLoc := structs.PointToGridPoint(item.Position)
			itemPile := sys.ItemLocations[itemLoc.X][itemLoc.Y]
			for i, _ := range itemPile {
				if itemPile[i].NetworkID == p.ItemId {
					if len(itemPile) == 1 {
						sys.ItemLocations[itemLoc.X][itemLoc.Y] = []*structs.Item{}
					} else {
						itemPile[i] = itemPile[len(itemPile)-1]
						sys.ItemLocations[itemLoc.X][itemLoc.Y] = itemPile[:len(itemPile)-1]
					}
				}
			}
			item.RenderComponent.Hidden = true
			item.OnGround = false
		}
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *UiSystem:
			if p.CreatureId == sys.input.player.NetworkID {
				log.Info("updating inv ui")
				sys.UpdatePlayerDisplay()
			}
		}
	}
	return true
}

type EquipItem struct {
	InventorySlot int
	CreatureId    structs.NetworkID
}

func (e *EquipItem) Name() string { return "Equipping item" }
func (e *EquipItem) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MapSystem:
			creature := sys.Creatures[e.CreatureId]
			item := creature.Inventory[e.InventorySlot]
			equipped := creature.Equipment[item.Type]
			creature.Equipment[item.Type] = item
			creature.Inventory[e.InventorySlot] = equipped

			// Adjust max life/stamina based on bonuses
			creature.Life += item.Bonuses.MaxLife
			creature.Stamina += item.Bonuses.MaxStamina
		}
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *UiSystem:
			if e.CreatureId == sys.input.player.NetworkID {
				log.Info("updating inv ui")
				sys.UpdatePlayerDisplay()
			}
		}
	}
	return true
}

type UnequipItem struct {
	EquipSlot  int
	CreatureId structs.NetworkID
}

func (e *UnequipItem) Name() string { return "Unequip item" }
func (e *UnequipItem) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MapSystem:
			creature := sys.Creatures[e.CreatureId]
			item := creature.Equipment[e.EquipSlot]
			// put the item in the first empty inventory slot
			for i, slot := range creature.Inventory {
				if slot == nil {
					creature.Inventory[i] = item
					break
				}
			}
			creature.Equipment[e.EquipSlot] = nil

			// Adjust max life/stamina based on bonuses
			creature.Life -= item.Bonuses.MaxLife
			creature.Stamina -= item.Bonuses.MaxStamina
		}
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *UiSystem:
			if e.CreatureId == sys.input.player.NetworkID {
				log.Info("updating inv ui")
				sys.UpdatePlayerDisplay()
			}
		}
	}
	return true
}

type EnemyTurn struct{}

// Decide the enemy actions on the host to avoid desync
func (e *EnemyTurn) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *EventSystem:
			if sys.serverRoom != nil {
				actions := ProcessEnemyTurn(w)
				actions = append(actions, &TurnChange{true})
				sys.outgoing <- NetworkMessage{
					Events: actions,
				}
			}
		}
	}
	return true
}
