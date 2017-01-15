package core

import (
	"engo.io/ecs"
	"engo.io/engo/common"

	"github.com/engoengine/math"
	"github.com/kyhavlov/go-dnd/structs"
)

type MoveSystem struct {
	networkIds      map[*ecs.BasicEntity]structs.NetworkID
	SpaceComponents map[structs.NetworkID]*common.SpaceComponent
}

// New is the initialisation of the System
func (ms *MoveSystem) New(w *ecs.World) {
	ms.networkIds = make(map[*ecs.BasicEntity]structs.NetworkID)
	ms.SpaceComponents = make(map[structs.NetworkID]*common.SpaceComponent)
}

func (ms *MoveSystem) Update(dt float32) {}

func (ms *MoveSystem) Add(entity *ecs.BasicEntity, space *common.SpaceComponent, nid structs.NetworkID) {
	ms.SpaceComponents[nid] = space
	ms.networkIds[entity] = nid
}

func (ms *MoveSystem) Remove(entity ecs.BasicEntity) {
	delete(ms.SpaceComponents, ms.networkIds[&entity])
	delete(ms.networkIds, &entity)
}

// Moves the entity with the given Id along the path
type MoveEvent struct {
	Id   structs.NetworkID
	Path []GridPoint

	// Tracks the node of the path we're currently moving towards
	// private field so we don't send it over the network
	current int
}

// Pixels per frame to move entities
const speed = 3

func (move *MoveEvent) Process(w *ecs.World, dt float32) bool {
	var lights *LightSystem
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *LightSystem:
			lights = sys
		}
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MoveSystem:
			current := &sys.SpaceComponents[move.Id].Position
			target := move.Path[move.current].toPixels()
			if current.PointDistance(target) <= 3.0 {
				lights.needsUpdate = true
				current.X = target.X
				current.Y = target.Y
				move.current++
				if move.current == len(move.Path) {
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
