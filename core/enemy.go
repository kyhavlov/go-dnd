package core

import (
	"engo.io/ecs"
	"github.com/kyhavlov/go-dnd/structs"
	"sort"
)

func ProcessEnemyTurn(w *ecs.World) []Event {
	var actions []Event
	var creatures []int
	var mapSys *MapSystem

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MapSystem:
			for _, creature := range sys.Creatures {
				if !creature.IsPlayerTeam {
					creatures = append(creatures, int(creature.NetworkID))
				}
			}
			mapSys = sys
		}
	}
	sort.Ints(creatures)

	for _, id := range creatures {
		creature := mapSys.Creatures[structs.NetworkID(id)]
		action := ProcessCreatureTurn(creature, mapSys)
		if action != nil {
			actions = append(actions, action)
		}
	}

	return actions
}

const ActivationRange = 10

func ProcessCreatureTurn(creature *structs.Creature, sys *MapSystem) Event {
	var closest *structs.Creature
	dist := 9999
	for _, player := range sys.Players {
		a := sys.GetTileAt(structs.PointToGridPoint(creature.Position))
		b := sys.GetTileAt(structs.PointToGridPoint(player.Position))
		path := GetPath(a, b, sys.Tiles, sys.CreatureLocations, TeamAny)
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

	target := structs.PointToGridPoint(closest.Position)
	target.Y += 1

	if sys.GetTileAt(target) != nil && sys.GetCreatureAt(target) == nil {
		a := sys.GetTileAt(structs.PointToGridPoint(creature.Position))
		b := sys.GetTileAt(target)
		path := GetPath(a, b, sys.Tiles, sys.CreatureLocations, TeamEnemy)
		if len(path) > creature.Movement {
			path = path[:creature.Movement]
		}
		return &Move{
			Id:   creature.NetworkID,
			Path: path,
		}
	}

	return nil
}
