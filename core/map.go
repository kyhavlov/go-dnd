package core

import (
	"engo.io/ecs"
	"engo.io/engo/common"

	"github.com/kyhavlov/go-dnd/structs"
)

// The map system tracks the tiles of the map
type MapSystem struct {
	Tiles [][]*structs.Tile
}

func (ms *MapSystem) Add(tile *structs.Tile) {
	if ms.Tiles[tile.X][tile.Y] == nil {
		ms.Tiles[tile.X][tile.Y] = tile
	}
}

func (ms *MapSystem) Update(dt float32)             {}
func (ms *MapSystem) Remove(entity ecs.BasicEntity) {}

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
				sys.Add(tile)
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
