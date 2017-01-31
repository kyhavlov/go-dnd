package structs

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

const LifeIcon = 25

type Creature struct {
	ecs.BasicEntity        `hcl:"-"`
	NetworkID              `hcl:"-"`
	common.SpaceComponent  `hcl:"-"`
	common.RenderComponent `hcl:"-"`

	LifeIcon    ecs.BasicEntity `hcl:"-"`
	LifeDisplay ecs.BasicEntity `hcl:"-"`

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
	IsActivated  bool
}

func NewCreature(name string, coords GridPoint) *Creature {
	creature := GetCreatureData(name)
	creature.Life = creature.MaxLife
	creature.Stamina = creature.MaxStamina
	creature.InnateSkills = append([]string{"Basic Attack"}, creature.InnateSkills...)

	creature.BasicEntity = ecs.NewBasic()
	creature.LifeIcon = ecs.NewBasic()
	creature.LifeDisplay = ecs.NewBasic()
	creature.SpaceComponent = common.SpaceComponent{
		Position: coords.ToPixels(),
		Width:    TileWidth,
		Height:   TileWidth,
	}
	creature.RenderComponent = common.RenderComponent{
		Drawable: Sprites.Cell(creature.Icon),
		Scale:    engo.Point{1, 1},
	}
	creature.RenderComponent.SetZIndex(1)

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

func (c *Creature) GetEffectiveMovement() int {
	life := c.Movement
	for _, item := range c.Equipment {
		if item != nil {
			life += item.Bonuses.Movement
		}
	}
	return life
}

func (c *Creature) GetEffectiveMaxLife() int {
	life := c.MaxLife
	for _, item := range c.Equipment {
		if item != nil {
			life += item.Bonuses.MaxLife
		}
	}
	return life
}

func (c *Creature) GetEffectiveStrength() int {
	str := c.Strength
	for _, item := range c.Equipment {
		if item != nil {
			str += item.Bonuses.Strength
		}
	}
	return str
}

func (c *Creature) GetEffectiveDexterity() int {
	dex := c.Dexterity
	for _, item := range c.Equipment {
		if item != nil {
			dex += item.Bonuses.Dexterity
		}
	}
	return dex
}

func (c *Creature) GetEffectiveIntelligence() int {
	intelligence := c.Intelligence
	for _, item := range c.Equipment {
		if item != nil {
			intelligence += item.Bonuses.Intelligence
		}
	}
	return intelligence
}
