package core

import (
	"engo.io/ecs"
	"engo.io/engo/common"
	"github.com/kyhavlov/go-dnd/structs"
)

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
