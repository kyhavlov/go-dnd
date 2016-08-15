package main

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"

	log "github.com/Sirupsen/logrus"
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

// Moves the entity with the given ID to NewLocation
type MoveEvent struct {
	Id          NetworkID
	NewLocation engo.Point
}

func (move *MoveEvent) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MoveSystem:
			sys.SpaceComponents[move.Id].Position.X = move.NewLocation.X
			sys.SpaceComponents[move.Id].Position.Y = move.NewLocation.Y
			log.Info("player position set to: ", sys.SpaceComponents[move.Id].Position)
		}
	}

	return true
}
