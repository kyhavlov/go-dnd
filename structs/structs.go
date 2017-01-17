package structs

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/engoengine/math/imath"
)

const SpritesheetPath = "textures/dungeon2x.png"
const TileWidth = 64

const MIN_BRIGHTNESS = 80
const INVENTORY_SIZE = 5

type NetworkID uint64

type Creature struct {
	ecs.BasicEntity
	NetworkID
	common.SpaceComponent
	common.RenderComponent

	HealthComponent
	StatComponent

	Inventory [INVENTORY_SIZE]*Item

	IsPlayerTeam bool
}

type HealthComponent struct {
	MaxLife int
	Life    int
	Dead    bool
}

type StatComponent struct {
	Strength     int
	Dexterity    int
	Intelligence int
	Stamina      int
}

type Item struct {
	ecs.BasicEntity
	NetworkID
	common.SpaceComponent
	common.RenderComponent

	OnGround  bool

	Life      int

	StrengthBonus     int
	DexterityBonus    int
	IntelligenceBonus int

	StaminaBonus      int
	StaminaRegenBonus int
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
