package structs

import (
	"engo.io/ecs"
	"engo.io/engo/common"
	"engo.io/engo"
	"github.com/engoengine/math/imath"
)

const SpritesheetPath = "textures/dungeon2x.png"
const TileWidth = 64

const MIN_BRIGHTNESS = 150

type NetworkID uint64

type Creature struct {
	ecs.BasicEntity
	NetworkID
	HealthComponent
	common.SpaceComponent
	common.RenderComponent

	IsPlayerTeam bool
}

type HealthComponent struct {
	Life int
	Dead bool
}

// GridPoint refers to a specific tile's coordinates; incrementing X by 1
// translates to going 1 tile to the right
type GridPoint struct {
	X int
	Y int
}

func (gp *GridPoint) ToPixels() engo.Point {
	return engo.Point{float32(gp.X * TileWidth), float32(gp.Y * TileWidth)}
}

func (gp *GridPoint) DistanceTo(other GridPoint) int {
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