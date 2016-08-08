package main

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	log "github.com/Sirupsen/logrus"
	"math/rand"
	"sort"
)

const SpritesheetPath = "textures/dungeon2x.png"
const TileWidth = 64
const MapWidth = 30
const MapHeight = 20

type Tile struct {
	ecs.BasicEntity
	NewTileAction
}

type GridPoint struct {
	X int
	Y int
}

func (gp *GridPoint) toPixels() engo.Point {
	return engo.Point{float32(gp.X * TileWidth), float32(gp.Y * TileWidth)}
}

func createMap(render *common.RenderSystem, input *InputSystem) NetworkMessage {
	tileCreations := make([]Action, 0)

	rooms, hallways := generateZone(987988)
	for _, room := range rooms {
		log.Println(room)
		for i := 0; i < room.Size; i++ {
			for j := 0; j < room.Size; j++ {
				newTile := &NewTileAction{
					SpaceComponent: common.SpaceComponent{
						Position: engo.Point{float32((room.X+i) * TileWidth), float32((room.Y+j) * TileWidth)},
						Width:    TileWidth,
						Height:   TileWidth,
					},
					Sprite: 861 + rand.Intn(8),
				}

				tileCreations = append(tileCreations, newTile)
			}
		}
	}

	for _, square := range hallways {
		newTile := &NewTileAction{
			SpaceComponent: common.SpaceComponent{
				Position: square.toPixels(),
				Width:    TileWidth,
				Height:   TileWidth,
			},
			Sprite: 861 + rand.Intn(8),
		}

		tileCreations = append(tileCreations, newTile)
	}

	return NetworkMessage{Actions: tileCreations}
}

type NewTileAction struct {
	common.RenderComponent
	common.SpaceComponent
	Sprite int
}

func (action *NewTileAction) Process(w *ecs.World, dt float32) bool {
	sheet := common.NewSpritesheetFromFile(SpritesheetPath, TileWidth, TileWidth)

	if sheet == nil {
		log.Fatalf("Unable to load texture file")
	}

	tile := Tile{}
	tile.BasicEntity = ecs.NewBasic()
	tile.SpaceComponent = action.SpaceComponent
	tile.RenderComponent = common.RenderComponent{
		Drawable: sheet.Cell(action.Sprite),
		Scale:    engo.Point{1, 1},
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&tile.BasicEntity, &tile.RenderComponent, &tile.SpaceComponent)
		}
	}

	return true
}

type RoomNode struct {
	Neighbors map[int]*RoomNode
	Id int
	GridPoint
	Size int
	visited bool
	depth int
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

func addNewRoom(id int, random *rand.Rand, edgeRoom *RoomNode) (*RoomNode, []GridPoint) {
	newRoom := &RoomNode{
		Neighbors: make(map[int]*RoomNode),
		Id: id,
		Size: 5,
		depth: edgeRoom.depth + 1,
	}

	// place room a random distance offset from selected edge room
	xOffset := -4 + random.Intn(8)
	yOffset := -4 + random.Intn(8)

	if xOffset > 0 {
		xOffset += edgeRoom.Size
	} else {
		xOffset -= newRoom.Size
	}

	if yOffset > 0 {
		yOffset += edgeRoom.Size
	} else {
		yOffset -= newRoom.Size
	}

	newRoom.GridPoint = GridPoint{edgeRoom.X + xOffset, edgeRoom.Y + yOffset}

	// generate a connecting hallway
	hallway := make([]GridPoint, 0)
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

	start := GridPoint{}
	end := GridPoint{}

	if leftRoom.X + leftRoom.Size > rightRoom.X && leftRoom.X < rightRoom.X + rightRoom.Size {
		start.X = rightRoom.X + random.Intn(leftRoom.X + leftRoom.Size - rightRoom.X)
		end.X = start.X
	} else {
		start.X = leftRoom.X + random.Intn(leftRoom.Size)
		end.X = rightRoom.X + random.Intn(rightRoom.Size)
	}

	if bottomRoom.Y + bottomRoom.Size > topRoom.Y && bottomRoom.Y < topRoom.Y + topRoom.Size {
		start.Y = topRoom.Y + random.Intn(bottomRoom.Y + bottomRoom.Size - topRoom.Y)
		end.Y = start.Y
	} else {
		start.Y = bottomRoom.Y + random.Intn(bottomRoom.Size)
		end.Y = topRoom.Y + random.Intn(topRoom.Size)
	}

	for i := start.Y; i <= end.Y; i++ {
		if bottomRoom == rightRoom {
			hallway = append(hallway, GridPoint{start.X, i})
		} else {
			hallway = append(hallway, GridPoint{end.X, i})
		}
	}

	for i := start.X; i <= end.X; i++ {
		hallway = append(hallway, GridPoint{i, start.Y})
	}

	log.Infof("Hallway for rooms (%d, %d) to (%d, %d), start: (%d, %d), end: (%d, %d)", edgeRoom.X, edgeRoom.Y, newRoom.X, newRoom.Y, start.X, start.Y, end.X, end.Y)
	log.Infof("%v", hallway)

	return newRoom, hallway
}

func roomIsValid(room *RoomNode, rooms []*RoomNode) bool {
	roomMax := GridPoint{
		X: room.X + room.Size,
		Y: room.Y + room.Size,
	}
	for _, otherRoom := range rooms {
		otherMax := GridPoint{
			X: otherRoom.X + otherRoom.Size,
			Y: otherRoom.Y + otherRoom.Size,
		}

		if roomMax.X > otherRoom.X && room.X < otherMax.X && roomMax.Y > otherRoom.Y && room.Y < otherMax.Y {
			return false
		}
	}
	return true
}

// Generates a map from a seed number, returns rooms and hallways
func generateZone(seed int64) (Rooms, []GridPoint) {
	random := rand.New(rand.NewSource(seed))
	rooms := make(Rooms, 0)
	hallways := make([]GridPoint, 0)

	dungeonLength := 7 + random.Intn(7)
	idInc := 0

	startingRoom := &RoomNode{
		Neighbors: make(map[int]*RoomNode),
		GridPoint: GridPoint{50, 50},
		Id: idInc,
		Size: 4,
		depth: 1,
	}
	rooms = append(rooms, startingRoom)

	depthReached := 1

	// Add rooms until the dungeon is the desired length
	for depthReached < dungeonLength {
		idInc++
		newRoom := &RoomNode{}
		madeNewRoom := false

		// try to spawn a new room, starting from the furthest room
		for _, edgeRoom := range rooms {
			newRoom, hallway := addNewRoom(idInc, random, edgeRoom)

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
			log.Info("Couldn't make new room, trying again")
			continue
		}

		log.Infof("Added new room with id %d and depth %d", newRoom.Id, newRoom.depth)

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

			spaces := ""
			for i := 0; i < current.depth; i++ {
				spaces = spaces + "-"
			}
			log.Infof("%s %d", spaces, current.Id)

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

	rooms = append(rooms, startingRoom)

	return rooms, hallways
}