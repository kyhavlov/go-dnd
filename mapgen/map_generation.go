package mapgen

import (
	"math/rand"
	"sort"

	"engo.io/engo"
	"engo.io/engo/common"
	log "github.com/Sirupsen/logrus"
	"github.com/kyhavlov/go-dnd/structs"
)

type RoomNode struct {
	Neighbors map[int]*RoomNode
	Id        int
	structs.GridPoint
	Size    int
	visited bool
	depth   int
}

func (room *RoomNode) Contains(point structs.GridPoint) bool {
	if room.X <= point.X && room.X+room.Size >= point.X && room.Y <= point.Y && room.Y+room.Size <= point.Y {
		return true
	}
	return false
}

type RoomQueue []*RoomNode

func (q *RoomQueue) Push(n *RoomNode) {
	*q = append(*q, n)
}

func (q *RoomQueue) Pop() (n *RoomNode) {
	n = (*q)[0]
	*q = (*q)[1:]
	return
}

func (q *RoomQueue) Len() int {
	return len(*q)
}

type Rooms []*RoomNode

func (m Rooms) Len() int {
	return len(m)
}
func (m Rooms) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
func (m Rooms) Less(i, j int) bool {
	return m[i].depth > m[j].depth
}

func addNewRoom(id int, random *rand.Rand, edgeRoom *RoomNode) (*RoomNode, []structs.GridPoint) {
	newRoom := &RoomNode{
		Neighbors: make(map[int]*RoomNode),
		Id:        id,
		Size:      5 + rand.Intn(6),
		depth:     edgeRoom.depth + 1,
	}

	// place room a random distance offset from selected edge room
	xOffset := -6 + random.Intn(edgeRoom.Size+12)
	yOffset := -6 + random.Intn(edgeRoom.Size+12)

	newRoom.GridPoint = structs.GridPoint{edgeRoom.X + xOffset, edgeRoom.Y + yOffset}

	// generate a connecting hallway
	hallway := make([]structs.GridPoint, 0)
	leftRoom := edgeRoom
	rightRoom := newRoom
	bottomRoom := edgeRoom
	topRoom := newRoom

	if newRoom.X < edgeRoom.X {
		leftRoom = newRoom
		rightRoom = edgeRoom
	}

	if newRoom.Y < edgeRoom.Y {
		bottomRoom = newRoom
		topRoom = edgeRoom
	}

	start := structs.GridPoint{}
	end := structs.GridPoint{}

	if leftRoom.X+leftRoom.Size > rightRoom.X && leftRoom.X < rightRoom.X+rightRoom.Size {
		start.X = rightRoom.X + random.Intn(leftRoom.X+leftRoom.Size-rightRoom.X)
		end.X = start.X
	} else {
		start.X = leftRoom.X + random.Intn(leftRoom.Size)
		end.X = rightRoom.X + random.Intn(rightRoom.Size)
	}

	if bottomRoom.Y+bottomRoom.Size > topRoom.Y && bottomRoom.Y < topRoom.Y+topRoom.Size {
		start.Y = topRoom.Y + random.Intn(bottomRoom.Y+bottomRoom.Size-topRoom.Y)
		end.Y = start.Y
	} else {
		start.Y = bottomRoom.Y + random.Intn(bottomRoom.Size)
		end.Y = topRoom.Y + random.Intn(topRoom.Size)
	}

	for i := start.Y; i <= end.Y; i++ {
		if bottomRoom == rightRoom {
			hallway = append(hallway, structs.GridPoint{start.X, i})
		} else {
			hallway = append(hallway, structs.GridPoint{end.X, i})
		}
	}

	for i := start.X; i <= end.X; i++ {
		hallway = append(hallway, structs.GridPoint{i, start.Y})
	}

	//log.Infof("Hallway for rooms (%d, %d) to (%d, %d), start: (%d, %d), end: (%d, %d)", edgeRoom.X, edgeRoom.Y, newRoom.X, newRoom.Y, start.X, start.Y, end.X, end.Y)
	//log.Infof("%v", hallway)

	return newRoom, hallway
}

func roomIsValid(room *RoomNode, rooms []*RoomNode) bool {
	roomMax := structs.GridPoint{
		X: room.X + room.Size,
		Y: room.Y + room.Size,
	}
	for _, otherRoom := range rooms {
		otherMax := structs.GridPoint{
			X: otherRoom.X + otherRoom.Size,
			Y: otherRoom.Y + otherRoom.Size,
		}

		// add 1 to the lengths so the walls of the rooms don't touch
		if roomMax.X+1 > otherRoom.X && room.X < otherMax.X+1 && roomMax.Y+1 > otherRoom.Y && room.Y < otherMax.Y+1 {
			return false
		}
	}
	return true
}

type Map struct {
	Tiles     []*structs.Tile
	Creatures []*structs.Creature
	Width     int
	Height    int
	StartLoc  structs.GridPoint
}

// Generates a map from a seed number
func GenerateMap(seed int64) *Map {
	random := rand.New(rand.NewSource(seed))
	var rooms Rooms
	var hallways []structs.GridPoint

	dungeonLength := 7 + random.Intn(7)
	idInc := 0

	startingRoom := &RoomNode{
		Neighbors: make(map[int]*RoomNode),
		Id:        idInc,
		Size:      4,
		depth:     1,
	}
	rooms = append(rooms, startingRoom)

	depthReached := 1

	// Add rooms until the dungeon is the desired length
	for depthReached < dungeonLength {
		idInc++
		newRoom := &RoomNode{}
		hallway := make([]structs.GridPoint, 0)
		madeNewRoom := false

		// try to spawn a new room, starting from the furthest room
		for _, edgeRoom := range rooms {
			newRoom, hallway = addNewRoom(idInc, random, edgeRoom)

			if !roomIsValid(newRoom, rooms) {
				continue
			}

			edgeRoom.Neighbors[newRoom.Id] = newRoom
			newRoom.Neighbors[edgeRoom.Id] = edgeRoom

			rooms = append(rooms, newRoom)
			hallways = append(hallways, hallway...)

			madeNewRoom = true
			break
		}

		if !madeNewRoom {
			log.Debugf("Couldn't make new room, trying again")
			continue
		}

		log.Debugf("Added new room with id %d and depth %d", newRoom.Id, newRoom.depth)

		// Recalculate the depth of the rooms (the number of rooms away from the start they are)
		for _, room := range rooms {
			room.visited = false
		}

		queue := make(RoomQueue, 0)

		queue.Push(startingRoom)

		for queue.Len() > 0 {
			current := queue.Pop()
			if current.visited {
				continue
			}

			current.visited = true

			/*spaces := ""
			for i := 0; i < current.depth; i++ {
				spaces = spaces + "-"
			}
			log.Infof("%s %d", spaces, current.Id)*/

			for _, neighbor := range current.Neighbors {
				if !neighbor.visited {
					neighbor.depth = current.depth + 1
					queue.Push(neighbor)
				}
			}
		}

		sort.Sort(rooms)
		depthReached = rooms[0].depth
	}

	// Get the minimum X/Y values rooms were placed at so we can align the level to 0,0
	offset := structs.GridPoint{
		X: startingRoom.X,
		Y: startingRoom.Y,
	}

	// and the maximum X/Y values so we know the bounds of the level
	maxPoint := structs.GridPoint{}

	for _, room := range rooms {
		if room.X < offset.X {
			offset.X = room.X
			log.Debugf("New min X room: id %d gridpoint %v", room.Id, room.GridPoint)
		}
		if room.Y < offset.Y {
			offset.Y = room.Y
		}
		if room.X+room.Size > maxPoint.X {
			maxPoint.X = room.X + room.Size
		}
		if room.Y+room.Size > maxPoint.Y {
			maxPoint.Y = room.Y + room.Size
		}
	}

	for _, tile := range hallways {
		if tile.X < offset.X {
			offset.X = tile.X
		}
		if tile.Y < offset.Y {
			offset.Y = tile.Y
		}
		if tile.X > maxPoint.X {
			maxPoint.X = tile.X
		}
		if tile.Y > maxPoint.Y {
			maxPoint.Y = tile.Y
		}
	}

	log.Infof("Map bounds: %d wide, %d tall", maxPoint.X-offset.X, maxPoint.Y-offset.Y)

	// Initialize the map
	level := &Map{
		Width:  maxPoint.X - offset.X,
		Height: maxPoint.Y - offset.Y,
	}

	common.CameraBounds.Max = engo.Point{
		X: float32(level.Width * structs.TileWidth),
		Y: float32(level.Height * structs.TileWidth),
	}

	// Add tiles for the map based on the rooms generated
	sheet := common.NewSpritesheetFromFile(structs.SpritesheetPath, structs.TileWidth, structs.TileWidth)
	if sheet == nil {
		log.Fatalf("Unable to load texture file")
	}
	for _, room := range rooms {
		room.X -= offset.X
		room.Y -= offset.Y
		log.Debug(room)
		for i := 0; i < room.Size; i++ {
			for j := 0; j < room.Size; j++ {
				loc := structs.GridPoint{
					X: room.X + i,
					Y: room.Y + j,
				}

				level.Tiles = append(level.Tiles, structs.NewTile("Dungeon Floor", loc))
			}
		}
	}

	level.StartLoc = startingRoom.GridPoint

	// Next, do the hallways
	for _, tile := range hallways {
		tile.X -= offset.X
		tile.Y -= offset.Y

		level.Tiles = append(level.Tiles, structs.NewTile("Dungeon Floor", tile))
	}

	// Spawn creatures in some of the rooms
	for _, room := range rooms {
		if random.Intn(2) == 0 && room != startingRoom {
			count := 1 + random.Intn(3)
			for i := 0; i < count; i++ {
				coords := structs.GridPoint{
					X: room.X + random.Intn(room.Size),
					Y: room.Y + random.Intn(room.Size),
				}
				creature := structs.NewCreature("Skeleton", coords)
				level.Creatures = append(level.Creatures, creature)
			}
		}
	}

	return level
}
