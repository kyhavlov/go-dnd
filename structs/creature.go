package structs

import (
	"engo.io/ecs"
	"engo.io/engo/common"
	"engo.io/engo"
)

type Creature struct {
	ecs.BasicEntity        `hcl:"-"`
	NetworkID              `hcl:"-"`
	common.SpaceComponent  `hcl:"-"`
	common.RenderComponent `hcl:"-"`

	HealthComponent `hcl:"-"`

	Name string `hcl:",key"`
	Icon int    `hcl:"icon"`

	StatComponent `hcl:"stats"`

	StartingItems []string `hcl:"items"`

	Equipment    [EquipmentSlots]*Item
	Inventory    [InventorySize]*Item
	InnateSkills []string `hcl:"skills"`
	Skills       []string `hcl:"-"`

	IsPlayerTeam bool
}

func NewCreature(name string, coords GridPoint) *Creature {
	creature := GetCreatureData(name)
	creature.BasicEntity = ecs.NewBasic()
	creature.SpaceComponent = common.SpaceComponent{
		Position: coords.ToPixels(),
		Width:    TileWidth,
		Height:   TileWidth,
	}
	creature.RenderComponent = common.RenderComponent{
		Drawable: sprites.Cell(creature.Icon),
		Scale:    engo.Point{1, 1},
	}
	creature.RenderComponent.SetZIndex(1)
	creature.InnateSkills = append([]string{"Basic Attack"}, creature.InnateSkills...)
	creature.Life = creature.MaxLife
	creature.Stamina = creature.MaxStamina

	return &creature
}

func (c *Creature) GetSkills() []string {
	var skills []string
	skills = append(skills, c.InnateSkills...)
	for _, item := range c.Equipment {
		if item == nil || len(item.Skills) == 0 {
			continue
		}
		skills = append(skills, item.Skills...)
	}
	return skills
}