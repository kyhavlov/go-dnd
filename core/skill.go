package core

import (
	log "github.com/Sirupsen/logrus"
	"github.com/kyhavlov/go-dnd/structs"
)

func GetSkillTargetLocation(target structs.SkillTarget, sys *MapSystem) structs.GridPoint {
	loc := target.Location
	if target.ID != 0 {
		targetCreature := sys.Creatures[target.ID]
		loc = structs.PointToGridPoint(targetCreature.Position)
	}
	return loc
}

func CanUseSkill(name string, sys *MapSystem, sourceID structs.NetworkID, target structs.SkillTarget, sourceLoc *structs.GridPoint) bool {
	skill := structs.GetSkillData(name)
	source := sys.Creatures[sourceID]
	a := structs.PointToGridPoint(source.SpaceComponent.Position)
	if sourceLoc != nil {
		a = *sourceLoc
	}
	b := GetSkillTargetLocation(target, sys)

	if source.Stamina < skill.StaminaCost {
		return false
	}

	return a.DistanceTo(b) >= skill.MinRange && a.DistanceTo(b) <= skill.MaxRange
}

func PerformSkillActions(name string, sys *MapSystem, sourceID structs.NetworkID, target structs.SkillTarget) {
	// Get skill data and source creature
	skill := structs.GetSkillData(name)
	source := sys.Creatures[sourceID]
	var targets []*structs.Creature

	// Get locations of source and target
	sourceLoc := structs.PointToGridPoint(source.SpaceComponent.Position)
	targetLoc := target.Location
	if target.ID != 0 {
		targetCreature, ok := sys.Creatures[target.ID]
		// if the target creature died before the skill could be used, whiff and do nothing
		if !ok {
			return
		}
		targetLoc = structs.PointToGridPoint(targetCreature.Position)
		targets = append(targets, targetCreature)
	}

	// Add extra targets if the skill has the cleave effect
	if _, ok := skill.Effects[structs.CleaveEffect]; ok {
		if sourceLoc.X == targetLoc.X {
			left := sys.GetCreatureAt(structs.GridPoint{targetLoc.X - 1, targetLoc.Y})
			right := sys.GetCreatureAt(structs.GridPoint{targetLoc.X + 1, targetLoc.Y})

			if left != nil {
				targets = append(targets, left)
			}
			if right != nil {
				targets = append(targets, right)
			}
		} else {
			top := sys.GetCreatureAt(structs.GridPoint{targetLoc.X, targetLoc.Y - 1})
			bottom := sys.GetCreatureAt(structs.GridPoint{targetLoc.X, targetLoc.Y + 1})

			if top != nil {
				targets = append(targets, top)
			}
			if bottom != nil {
				targets = append(targets, bottom)
			}
		}
	}

	// Add extra targets in radius
	if radius, ok := skill.Effects[structs.AoeEffect]; ok {
		for i := 0; i < radius*2+1; i++ {
			for j := 0; j < radius*2+1; j++ {
				currentLoc := structs.GridPoint{
					X: targetLoc.X - radius + i,
					Y: targetLoc.Y - radius + j,
				}
				creature := sys.GetCreatureAt(currentLoc)
				if creature != nil {
					targets = append(targets, creature)
				}
			}
		}
	}

	// Add extra targets from piercing
	if pierceDistance, ok := skill.Effects[structs.PierceEffect]; ok {
		for i := 1; i <= pierceDistance; i++ {
			xDiff := i * (targetLoc.X - sourceLoc.X)
			yDiff := i * (targetLoc.Y - sourceLoc.Y)
			currentLoc := structs.GridPoint{
				X: targetLoc.X + xDiff,
				Y: targetLoc.Y + yDiff,
			}
			creature := sys.GetCreatureAt(currentLoc)
			if creature != nil {
				targets = append(targets, creature)
			}
		}
	}

	for _, t := range targets {
		damage := skill.Damage
		damage += int(skill.DamageBonuses.Str * float64(source.GetEffectiveStrength()))
		damage += int(skill.DamageBonuses.Dex * float64(source.GetEffectiveDexterity()))
		damage += int(skill.DamageBonuses.Int * float64(source.GetEffectiveIntelligence()))
		t.Life -= damage
		log.Infof("Creature id %d took %d damage from %s, at %d life now", t.NetworkID, damage, name, t.Life)
		if t.Life <= 0 {
			sys.RemoveCreature(t)
		}
	}

	source.Stamina -= skill.StaminaCost
}
