package main

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"

	log "github.com/Sirupsen/logrus"
	"github.com/engoengine/math/imath"
	"image/color"
)

const SpritesheetPath = "textures/dungeon2x.png"
const TileWidth = 64

// GridPoint refers to a specific tile's coordinates; incrementing X by 1
// translates to going 1 tile to the right
type GridPoint struct {
	X int
	Y int
}

func (gp *GridPoint) toPixels() engo.Point {
	return engo.Point{float32(gp.X * TileWidth), float32(gp.Y * TileWidth)}
}

func PointToGridPoint(p engo.Point) GridPoint {
	return GridPoint{
		X: int(p.X / TileWidth),
		Y: int(p.Y / TileWidth),
	}
}

type Tile struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
	NewTileEvent
}

// The map system tracks the tiles and light levels of the map
type MapSystem struct {
	Tiles [][]*Tile
}

func (ms *MapSystem) Add(tile *Tile) {
	if ms.Tiles[tile.X][tile.Y] == nil {
		ms.Tiles[tile.X][tile.Y] = tile
	}
}

func (ms *MapSystem) Update(dt float32)             {}
func (ms *MapSystem) Remove(entity ecs.BasicEntity) {}

func (ms *MapSystem) GetTileAt(point GridPoint) *Tile {
	return ms.Tiles[point.X][point.Y]
}

// Initializes the map and sets camera settings based on the size
type InitMapEvent struct {
	Width, Height int
}

func (event *InitMapEvent) Process(w *ecs.World, dt float32) bool {
	common.CameraBounds.Max = engo.Point{
		X: float32(event.Width * TileWidth),
		Y: float32(event.Height * TileWidth),
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MapSystem:
			sys.Tiles = make([][]*Tile, event.Width)
			for i, _ := range sys.Tiles {
				sys.Tiles[i] = make([]*Tile, event.Height)
			}
		}
	}

	return true
}

// Adds a new tile to the map
type NewTileEvent struct {
	GridPoint
	Sprite int
}

func (event *NewTileEvent) Process(w *ecs.World, dt float32) bool {
	sheet := common.NewSpritesheetFromFile(SpritesheetPath, TileWidth, TileWidth)

	if sheet == nil {
		log.Fatalf("Unable to load texture file")
	}

	tile := Tile{}
	tile.BasicEntity = ecs.NewBasic()
	tile.SpaceComponent = common.SpaceComponent{
		Position: event.GridPoint.toPixels(),
		Width:    TileWidth,
		Height:   TileWidth,
	}
	tile.RenderComponent = common.RenderComponent{
		Drawable: sheet.Cell(event.Sprite),
		Color:    color.Alpha{250 - uint8(imath.Min((imath.Abs(event.X-6)+imath.Abs(event.Y-4))*20, 150))},
		Scale:    engo.Point{1, 1},
	}
	tile.GridPoint = event.GridPoint

	added := false
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MapSystem:
			if sys.GetTileAt(tile.GridPoint) == nil {
				sys.Add(&tile)
				added = true
			}
		}
	}

	// don't add something to the render system unless we added it to the map
	// TODO: don't try to create overlapping tiles in the first place
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			if added {
				sys.Add(&tile.BasicEntity, &tile.RenderComponent, &tile.SpaceComponent)
			}
		}
	}

	return true
}
