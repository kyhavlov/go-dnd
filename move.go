package main

import (
	"engo.io/ecs"
	"engo.io/engo/common"
)

type MoveSystem struct {
	networkIds      map[*ecs.BasicEntity]NetworkID
	SpaceComponents map[NetworkID]*common.SpaceComponent
}

// New is the initialisation of the System
func (ms *MoveSystem) New(w *ecs.World) {
	ms.networkIds = make(map[*ecs.BasicEntity]NetworkID)
	ms.SpaceComponents = make(map[NetworkID]*common.SpaceComponent)
}

func (ms *MoveSystem) Update(dt float32) {}

func (ds *MoveSystem) Add(entity *ecs.BasicEntity, space *common.SpaceComponent, nid NetworkID) {
	ds.SpaceComponents[nid] = space
	ds.networkIds[entity] = nid
}

func (ms *MoveSystem) Remove(entity ecs.BasicEntity) {
	delete(ms.SpaceComponents, ms.networkIds[&entity])
	delete(ms.networkIds, &entity)
}
