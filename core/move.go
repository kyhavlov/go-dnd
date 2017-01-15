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

	Creatures map[structs.NetworkID]*structs.Creature
	CreatureLocations [][]*structs.Creature
}

// New is the initialisation of the System
func (ms *MoveSystem) New(w *ecs.World) {
	ms.networkIds = make(map[*ecs.BasicEntity]structs.NetworkID)
	ms.SpaceComponents = make(map[structs.NetworkID]*common.SpaceComponent)
	ms.Creatures = make(map[structs.NetworkID]*structs.Creature)
}

func (ms *MoveSystem) Update(dt float32) {}

func (ms *MoveSystem) Add(entity *ecs.BasicEntity, space *common.SpaceComponent, nid structs.NetworkID) {
	ms.SpaceComponents[nid] = space
	ms.networkIds[entity] = nid
}

func (ms *MoveSystem) AddCreature(creature *structs.Creature) {
	ms.SpaceComponents[creature.NetworkID] = &creature.SpaceComponent
	ms.networkIds[&creature.BasicEntity] = creature.NetworkID
	gridPoint := structs.PointToGridPoint(creature.Position)
	ms.CreatureLocations[gridPoint.X][gridPoint.Y] = creature
	ms.Creatures[creature.NetworkID] = creature
}

func (ms *MoveSystem) GetCreatureAt(point structs.GridPoint) *structs.Creature {
	return ms.CreatureLocations[point.X][point.Y]
}

func (ms *MoveSystem) Remove(entity ecs.BasicEntity) {
	delete(ms.SpaceComponents, ms.networkIds[&entity])
	delete(ms.networkIds, &entity)
}

// Moves the entity with the given Id along the path
type MoveEvent struct {
	Id   structs.NetworkID
	Path []structs.GridPoint

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
