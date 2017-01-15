package core

import (
	"engo.io/ecs"
	"engo.io/engo/common"
	log "github.com/Sirupsen/logrus"
	"github.com/kyhavlov/go-dnd/structs"
)

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
		case *MoveSystem:
			sys.AddCreature(creature)
		case *HealthSystem:
			sys.Add(creature.BasicEntity, &creature.HealthComponent)
		}
	}
}

type Attack struct {
	Id       structs.NetworkID
	TargetId structs.NetworkID
}

func (attack *Attack) Process(w *ecs.World, dt float32) bool {
	var creature *structs.Creature
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MoveSystem:
			creature = sys.Creatures[attack.TargetId]
			creature.Life -= 20
			log.Infof("Creature id %d took %d damage, at %d life now", attack.TargetId, 20, creature.Life)
		}
	}

	return true
}
