package core

import (
	log "github.com/Sirupsen/logrus"
	"github.com/engoengine/math/imath"
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

	maxRange := skill.MaxRange
	if skill.HasTag(structs.MeleeTag) {
		maxRange += 1
	}

	return a.DistanceTo(b) >= skill.MinRange && a.DistanceTo(b) <= maxRange
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
		if creature := sys.GetCreatureAt(targetLoc); creature != nil {
			targets = append(targets, creature)
		}

		xDiff := targetLoc.X - sourceLoc.X
		if xDiff != 0 {
			xDiff = imath.Abs(xDiff) / xDiff
		}

		yDiff := targetLoc.Y - sourceLoc.Y
		if yDiff != 0 {
			yDiff = imath.Abs(yDiff) / yDiff
		}

		target1 := structs.GridPoint{
			X: targetLoc.X + yDiff,
			Y: targetLoc.Y - xDiff,
		}
		if creature := sys.GetCreatureAt(target1); creature != nil {
			targets = append(targets, creature)
		}
		target2 := structs.GridPoint{
			X: targetLoc.X - yDiff,
			Y: targetLoc.Y + xDiff,
		}
		if creature := sys.GetCreatureAt(target2); creature != nil {
			targets = append(targets, creature)
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
			xDiff := targetLoc.X - sourceLoc.X
			if xDiff != 0 {
				xDiff = imath.Abs(xDiff) / xDiff
			}
			yDiff := targetLoc.Y - sourceLoc.Y
			if yDiff != 0 {
				yDiff = imath.Abs(yDiff) / yDiff
			}
			currentLoc := structs.GridPoint{
				X: targetLoc.X + i*xDiff,
				Y: targetLoc.Y + i*yDiff,
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
