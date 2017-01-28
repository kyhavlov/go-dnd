package core

import (
	log "github.com/Sirupsen/logrus"
	"github.com/kyhavlov/go-dnd/structs"
)

func SkillTargetIsValid(name string, sys *MapSystem, sourceID, targetID structs.NetworkID) bool {
	skill := structs.GetSkillData(name)
	source := sys.Creatures[sourceID]
	target := sys.Creatures[targetID]

	a := structs.PointToGridPoint(source.SpaceComponent.Position)
	b := structs.PointToGridPoint(target.SpaceComponent.Position)

	return a.DistanceTo(b) >= skill.MinRange && a.DistanceTo(b) <= skill.MaxRange
}

func PerformSkillActions(name string, sys *MapSystem, sourceID, targetID structs.NetworkID) {
	skill := structs.GetSkillData(name)
	source := sys.Creatures[sourceID]
	target := sys.Creatures[targetID]
	a := structs.PointToGridPoint(source.SpaceComponent.Position)
	b := structs.PointToGridPoint(target.SpaceComponent.Position)

	targets := []*structs.Creature{target}

	// Add extra targets if the skill has the cleave effect
	if _, ok := skill.Effects[structs.CleaveEffect]; ok {
		if a.X == b.X {
			left := sys.GetCreatureAt(structs.GridPoint{b.X - 1, b.Y})
			right := sys.GetCreatureAt(structs.GridPoint{b.X + 1, b.Y})

			if left != nil {
				targets = append(targets, left)
			}
			if right != nil {
				targets = append(targets, right)
			}
		} else {
			top := sys.GetCreatureAt(structs.GridPoint{b.X, b.Y - 1})
			bottom := sys.GetCreatureAt(structs.GridPoint{b.X, b.Y + 1})

			if top != nil {
				targets = append(targets, top)
			}
			if bottom != nil {
				targets = append(targets, bottom)
			}
		}
	}

	for _, t := range targets {
		damage := skill.Damage
		damage += int(skill.DamageBonuses.Str * float64(source.Strength))
		damage += int(skill.DamageBonuses.Int * float64(source.Intelligence))
		t.Life -= damage
		log.Infof("Creature id %d took %d damage from %s, at %d life now", t.NetworkID, damage, name, t.Life)
	}
}
