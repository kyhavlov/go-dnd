package main

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"image/color"

	log "github.com/Sirupsen/logrus"
	"github.com/engoengine/math/imath"
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

func (gp *GridPoint) distanceTo(other GridPoint) int {
	return imath.Abs(gp.X-other.X) + imath.Abs(gp.Y-other.Y)
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
	GridPoint
	Sprite int
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

func (ms *MapSystem) MapWidth() int {
	return len(ms.Tiles)
}

func (ms *MapSystem) MapHeight() int {
	return len(ms.Tiles[0])
}

func AddNewTile(w *ecs.World, coords GridPoint, sprite int) bool {
	sheet := common.NewSpritesheetFromFile(SpritesheetPath, TileWidth, TileWidth)

	if sheet == nil {
		log.Fatalf("Unable to load texture file")
	}

	tile := Tile{}
	tile.BasicEntity = ecs.NewBasic()
	tile.SpaceComponent = common.SpaceComponent{
		Position: coords.toPixels(),
		Width:    TileWidth,
		Height:   TileWidth,
	}
	tile.RenderComponent = common.RenderComponent{
		Drawable: sheet.Cell(sprite),
		Color:    color.Alpha{MIN_BRIGHTNESS},
		Scale:    engo.Point{1, 1},
	}
	tile.GridPoint = coords

	added := false
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MapSystem:
			if sys.GetTileAt(coords) == nil {
				sys.Add(&tile)
				added = true
			}
		case *LightSystem:
			sys.needsUpdate = true
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
