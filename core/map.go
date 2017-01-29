package core

import (
	"engo.io/ecs"
	"engo.io/engo/common"

	"github.com/kyhavlov/go-dnd/mapgen"
	"github.com/kyhavlov/go-dnd/structs"
)

// The map system tracks the tiles of the map
type MapSystem struct {
	networkIds      map[*ecs.BasicEntity]structs.NetworkID
	SpaceComponents map[structs.NetworkID]*common.SpaceComponent

	MapInfo *mapgen.Map

	Tiles [][]*structs.Tile

	Creatures         map[structs.NetworkID]*structs.Creature
	CreatureLocations [][]*structs.Creature

	Items         map[structs.NetworkID]*structs.Item
	ItemLocations [][]*structs.Item
}

// New is the initialisation of the System
func (ms *MapSystem) New(w *ecs.World) {
	ms.networkIds = make(map[*ecs.BasicEntity]structs.NetworkID)
	ms.SpaceComponents = make(map[structs.NetworkID]*common.SpaceComponent)
	ms.Creatures = make(map[structs.NetworkID]*structs.Creature)
	ms.Items = make(map[structs.NetworkID]*structs.Item)
}

func (ms *MapSystem) Add(entity *ecs.BasicEntity, space *common.SpaceComponent, nid structs.NetworkID) {
	ms.SpaceComponents[nid] = space
	ms.networkIds[entity] = nid
}

func (ms *MapSystem) AddTile(tile *structs.Tile) {
	if ms.Tiles[tile.X][tile.Y] == nil {
		ms.Tiles[tile.X][tile.Y] = tile
	}
}

func (ms *MapSystem) AddCreature(creature *structs.Creature) {
	ms.SpaceComponents[creature.NetworkID] = &creature.SpaceComponent
	ms.networkIds[&creature.BasicEntity] = creature.NetworkID
	gridPoint := structs.PointToGridPoint(creature.Position)
	ms.CreatureLocations[gridPoint.X][gridPoint.Y] = creature
	ms.Creatures[creature.NetworkID] = creature
}

func (ms *MapSystem) GetCreatureAt(point structs.GridPoint) *structs.Creature {
	return ms.CreatureLocations[point.X][point.Y]
}

func (ms *MapSystem) AddItem(item *structs.Item) {
	ms.SpaceComponents[item.NetworkID] = &item.SpaceComponent
	ms.networkIds[&item.BasicEntity] = item.NetworkID
	gridPoint := structs.PointToGridPoint(item.Position)
	ms.ItemLocations[gridPoint.X][gridPoint.Y] = item
	ms.Items[item.NetworkID] = item
}

func (ms *MapSystem) GetItemAt(point structs.GridPoint) *structs.Item {
	return ms.ItemLocations[point.X][point.Y]
}
func (ms *MapSystem) Update(dt float32) {}
func (ms *MapSystem) Remove(entity ecs.BasicEntity) {
	delete(ms.SpaceComponents, ms.networkIds[&entity])
	delete(ms.networkIds, &entity)
}

func (ms *MapSystem) GetTileAt(point structs.GridPoint) *structs.Tile {
	return ms.Tiles[point.X][point.Y]
}

func (ms *MapSystem) MapWidth() int {
	return len(ms.Tiles)
}

func (ms *MapSystem) MapHeight() int {
	return len(ms.Tiles[0])
}

func AddTile(w *ecs.World, tile *structs.Tile) {
	added := false
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MapSystem:
			if sys.GetTileAt(tile.GridPoint) == nil {
				sys.AddTile(tile)
				added = true
			}
		case *LightSystem:
			sys.needsUpdate = true
		}
	}

	// don't add something to the render system unless we added it to the map
	// TODO: don't try to create overlapping tiles in the first place
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			if added {
				sys.Add(&tile.BasicEntity, &tile.RenderComponent, &tile.SpaceComponent)
			}
		}
	}
}
