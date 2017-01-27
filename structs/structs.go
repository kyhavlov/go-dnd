package structs

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/engoengine/math/imath"
)

const SpritesheetPath = "textures/dungeon2x.png"
const TileWidth = 64

const MinBrightness = 80
const InventorySize = 5
const EquipmentSlots = 5
const SkillSlots = 10

type NetworkID uint64

type Creature struct {
	ecs.BasicEntity
	NetworkID
	common.SpaceComponent
	common.RenderComponent

	HealthComponent
	StatComponent

	Equipment    [EquipmentSlots]*Item
	Inventory    [InventorySize]*Item
	InnateSkills []string
	Skills       []string

	IsPlayerTeam bool
}

type HealthComponent struct {
	MaxLife int
	Life    int
	Dead    bool
}

type StatComponent struct {
	MaxLife      int `hcl:"life"`
	Strength     int `hcl:"str"`
	Dexterity    int `hcl:"dex"`
	Intelligence int `hcl:"int"`
	Stamina      int
	StaminaRegen int `hcl:"stamina_regen"`
}

type ItemType int

const (
	Weapon = iota
	OffHand
	Armor
	Helm
	Accessory
)

type Item struct {
	ecs.BasicEntity        `hcl:"-"`
	NetworkID              `hcl:"-"`
	common.SpaceComponent  `hcl:"-"`
	common.RenderComponent `hcl:"-"`

	Name string `hcl:",key"`
	Icon int    `hcl:"icon"`

	OnGround bool `hcl:"-"`

	Slot string
	Type ItemType `hcl:"-"`

	Skills []string

	Requirements StatComponent `hcl:"reqs"`
	Bonuses      StatComponent `hcl:"bonus"`
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
