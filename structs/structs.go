package structs

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/engoengine/math/imath"
	"image/color"
	"math/rand"
)

const SpritesheetPath = "textures/dungeon2x.png"
const TileWidth = 64

var Sprites *common.Spritesheet

func LoadSprites() {
	engo.Files.Load(SpritesheetPath)
	Sprites = common.NewSpritesheetFromFile(SpritesheetPath, TileWidth, TileWidth)
}

const MinBrightness = 80
const InventorySize = 5
const EquipmentSlots = 5
const SkillSlots = 10

type NetworkID uint64

type HealthComponent struct {
	Life int
	Dead bool
}

type StatComponent struct {
	Movement     int `hcl:"move"`
	MaxLife      int `hcl:"life"`
	Strength     int `hcl:"str"`
	Dexterity    int `hcl:"dex"`
	Intelligence int `hcl:"int"`
	MaxStamina   int `hcl:"stamina"`
	Stamina      int `hcl:"-"`
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

	OnGround bool `hcl:"-"`

	Name string `hcl:",key"`
	Icon int    `hcl:"icon"`

	Slot string
	Type ItemType `hcl:"-"`

	Skills []string

	Requirements StatComponent `hcl:"reqs"`
	Bonuses      StatComponent `hcl:"bonus"`
}

func NewItem(name string, coords GridPoint) *Item {
	item := GetItemData(name)
	item.OnGround = true
	item.BasicEntity = ecs.NewBasic()
	item.SpaceComponent = common.SpaceComponent{
		Position: coords.ToPixels(),
		Width:    TileWidth,
		Height:   TileWidth,
	}
	item.RenderComponent = common.RenderComponent{
		Drawable: Sprites.Cell(item.Icon),
		Scale:    engo.Point{0.5, 0.5},
	}

	return &item
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
	ecs.BasicEntity        `hcl:"-"`
	common.RenderComponent `hcl:"-"`
	common.SpaceComponent  `hcl:"-"`
	GridPoint              `hcl:"-"`

	Name  string `hcl:",key"`
	Icons []int
}

func NewTile(name string, coords GridPoint) *Tile {
	tile := GetTileData(name)
	tile.BasicEntity = ecs.NewBasic()
	tile.SpaceComponent = common.SpaceComponent{
		Position: coords.ToPixels(),
		Width:    TileWidth,
		Height:   TileWidth,
	}
	sprite := tile.Icons[rand.Intn(len(tile.Icons))]
	tile.RenderComponent = common.RenderComponent{
		Drawable: Sprites.Cell(sprite),
		Color:    color.Alpha{MinBrightness},
		Scale:    engo.Point{1, 1},
	}
	tile.RenderComponent.SetZIndex(-100)
	tile.GridPoint = coords

	return &tile
}

type Skill struct {
	Name string `hcl:",key"`

	Icon int

	MinRange      int  `hcl:"min_range"`
	MaxRange      int  `hcl:"max_range"`
	TargetsGround bool `hcl:"targets_ground"`

	Damage      int
	StaminaCost int `hcl:"stamina_cost"`

	DamageBonuses StatModifiers `hcl:"damage_bonuses"`

	Effects map[string]int
}

type StatModifiers struct {
	Str float64
	Dex float64
	Int float64
}

const CleaveEffect = "hits_perpendicular"
const PierceEffect = "pierces"
const AoeEffect = "aoe_radius"

type SkillTarget struct {
	ID       NetworkID
	Location GridPoint
}
