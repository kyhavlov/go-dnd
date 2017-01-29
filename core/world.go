package core

import (
	"engo.io/ecs"
	"engo.io/engo/common"
	"github.com/kyhavlov/go-dnd/structs"
)

// Utility functions for adding game objects to all the necessary world systems

func AddCreature(w *ecs.World, creature *structs.Creature) {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *NetworkSystem:
			creature.NetworkID = sys.nextId()
		}
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&creature.BasicEntity, &creature.RenderComponent, &creature.SpaceComponent)
		case *MapSystem:
			sys.AddCreature(creature)
		case *HealthSystem:
			sys.Add(creature.BasicEntity, &creature.HealthComponent)
		}
	}
}

func AddItem(w *ecs.World, item *structs.Item) {
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
		case *MapSystem:
			sys.AddItem(item)
		}
	}
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