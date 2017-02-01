package core

import (
	"github.com/engoengine/math/imath"
	"github.com/kyhavlov/go-dnd/structs"
	//log "github.com/Sirupsen/logrus"
)

type Team int

const (
	TeamAny Team = iota
	TeamPlayer
	TeamEnemy
)

// Standard A* implementation for finding a shortest path between two map tiles.
// Comments stolen from wikipedia article on A*: https://en.wikipedia.org/wiki/A*_search_algorithm
func GetPath(start, goal *structs.Tile, tiles [][]*structs.Tile, creatures [][]*structs.Creature, team Team) []structs.GridPoint {
	// The set of nodes already evaluated
	closedSet := make(map[*structs.Tile]bool)
	// The set of currently discovered nodes still to be evaluated.
	// Initially, only the start node is known.
	openSet := make(map[*structs.Tile]bool)
	openSet[start] = true

	// For each node, which node it can most efficiently be reached from.
	// If a node can be reached from many nodes, cameFrom will eventually contain the
	// most efficient previous step.
	cameFrom := make(map[*structs.Tile]*structs.Tile)

	// For each node, the cost of getting from the start node to that node.
	gScore := make(map[*structs.Tile]int)
	// The cost of going from start to start is zero.
	gScore[start] = 0

	// For each node, the total cost of getting from the start node to the goal
	// by passing by that node. That value is partly known, partly heuristic.
	fScore := make(map[*structs.Tile]int)
	// For the first node, that value is completely heuristic.
	fScore[start] = getEstimatedDistance(start, goal)

	path := make([]structs.GridPoint, 0)

	// Define a function, sameTeam, for checking whether a creature is on a team we're allowed to move through
	var sameTeam func(x, y int) bool
	switch team {
	case TeamAny:
		sameTeam = func(x, y int) bool {
			return true
		}
	case TeamPlayer:
		sameTeam = func(x, y int) bool {
			return creatures[x][y] == nil || creatures[x][y].IsPlayerTeam
		}
	case TeamEnemy:
		sameTeam = func(x, y int) bool {
			return creatures[x][y] == nil || !creatures[x][y].IsPlayerTeam
		}
	}

	for len(openSet) > 0 {
		// Set current to the node in the open set with the lowest fScore
		// TODO: make the open set a priority queue instead of a plain map for better performance
		var current *structs.Tile = nil
		for tile, _ := range openSet {
			if current == nil {
				current = tile
				continue
			} else if fScore[tile] < fScore[current] {
				current = tile
			}
		}

		// If we've arrived at the goal, we can construct the path in reverse and return it
		if current == goal {
			path = append(path, current.GridPoint)
			for current != start {
				current = cameFrom[current]
				path = append(path, current.GridPoint)
			}

			// reverse the path so it goes from start to finish
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}

			return path
		}

		delete(openSet, current)
		closedSet[current] = true

		// Evaluate adjacent tiles to the current one
		for _, neighbor := range getNeighbors(current, tiles, sameTeam) {
			// Ignore the neighbors which are already evaluated.
			if _, ok := closedSet[neighbor]; ok {
				continue
			}

			// The distance from start to this neighbor
			tentativeGScore := gScore[current] + 1

			if _, ok := openSet[neighbor]; !ok {
				// New tile discovered that wasn't in the closed or open sets
				openSet[neighbor] = true
			} else if tentativeGScore >= gScore[neighbor] {
				// We didn't discover this tile along a better path
				continue
			}

			// This is the best path from the start that we've found for this tile, record its scores
			// and place it in cameFrom
			cameFrom[neighbor] = current
			gScore[neighbor] = tentativeGScore
			fScore[neighbor] = tentativeGScore + getEstimatedDistance(neighbor, goal)
		}
	}

	// Couldn't find a path, exhausted the search, return an empty list
	return path
}

func getEstimatedDistance(a, b *structs.Tile) int {
	return imath.Abs(a.X-b.X) + imath.Abs(a.Y-b.Y)
}

// TODO: re-use the neighbors slice instead of allocating a new one every time we call this
func getNeighbors(tile *structs.Tile, tiles [][]*structs.Tile, sameTeam func(x, y int) bool) []*structs.Tile {
	neighbors := make([]*structs.Tile, 0)

	if tile.X > 0 && tiles[tile.X-1][tile.Y] != nil && sameTeam(tile.X-1, tile.Y) {
		neighbors = append(neighbors, tiles[tile.X-1][tile.Y])
	}

	if tile.X < len(tiles)-1 && tiles[tile.X+1][tile.Y] != nil && sameTeam(tile.X+1, tile.Y) {
		neighbors = append(neighbors, tiles[tile.X+1][tile.Y])
	}

	if tile.Y > 0 && tiles[tile.X][tile.Y-1] != nil && sameTeam(tile.X, tile.Y-1) {
		neighbors = append(neighbors, tiles[tile.X][tile.Y-1])
	}

	if tile.Y < len(tiles[0])-1 && tiles[tile.X][tile.Y+1] != nil && sameTeam(tile.X, tile.Y+1) {
		neighbors = append(neighbors, tiles[tile.X][tile.Y+1])
	}

	return neighbors
}
