package core

import (
	"engo.io/ecs"
	"engo.io/engo/common"
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
		case *MapSystem:
			sys.AddCreature(creature)
		case *HealthSystem:
			sys.Add(creature.BasicEntity, &creature.HealthComponent)
		}
	}
}

func GetCreatureSkills(creature *structs.Creature) []string {
	var skills []string
	skills = append(skills, creature.InnateSkills...)
	for _, item := range creature.Equipment {
		if item == nil {
			continue
		}
		skills = append(skills, item.Skills...)
	}
	return skills
}
