package core

import (
	"github.com/kyhavlov/go-dnd/structs"
)

const ActivationRange = 10

func ProcessCreatureTurn(id structs.NetworkID, sys *MapSystem) []Event {
	creature := sys.Creatures[id]
	var closest *structs.Creature
	creatureTile := sys.GetTileAt(structs.PointToGridPoint(creature.Position))
	dist := 9999
	// Find the closest player
	for _, player := range sys.Players {
		if player.Dead {
			continue
		}

		b := sys.GetTileAt(structs.PointToGridPoint(player.Position))
		path := GetPath(creatureTile, b, sys.Tiles, sys.CreatureLocations, TeamAny)
		if len(path) < dist {
			dist = len(path)
			closest = player
		}
	}

	if !creature.IsActivated {
		if dist <= ActivationRange {
			creature.IsActivated = true
		} else {
			return nil
		}
	}

	// Get potential squares around the target player to move to
	target := structs.PointToGridPoint(closest.Position)
	neighbors := getNeighbors(sys.GetTileAt(target), sys.Tiles, func(x, y int) bool { return true })
	var path []structs.GridPoint
	shortestPath := 9999
	for _, neighbor := range neighbors {
		if sys.GetCreatureAt(neighbor.GridPoint) != nil {
			continue
		}
		currentPath := GetPath(creatureTile, neighbor, sys.Tiles, sys.CreatureLocations, TeamEnemy)
		if len(currentPath) > 0 && len(currentPath) < shortestPath {
			path = currentPath
			shortestPath = len(path)
		}
	}

	if len(path) > creature.Movement {
		path = path[:creature.Movement]
	}
	var actions []Event
	if len(path) > 0 {
		actions = append(actions, &Move{
			Id:   creature.NetworkID,
			Path: path,
		})
	}

	// Try to use the creature's first skill
	skill := structs.GetSkillData(creature.GetSkills()[0])
	skillTarget := structs.SkillTarget{ID: closest.NetworkID}
	effectiveLoc := creatureTile.GridPoint
	if len(path) > 0 {
		effectiveLoc = path[len(path)-1]
	}
	if CanUseSkill(skill.Name, sys, creature.NetworkID, skillTarget, &effectiveLoc) {
		actions = append(actions, &UseSkill{
			SkillName: skill.Name,
			Source:    creature.NetworkID,
			Target:    skillTarget,
		})
	}

	return actions
}
