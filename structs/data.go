package structs

import (
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/hcl"
)

var itemData map[string]Item
var creatureData map[string]Creature
var tileData map[string]Tile
var skillData map[string]Skill

type Data struct {
	Items     []Item     `hcl:"item"`
	Creatures []Creature `hcl:"creature"`
	Tiles     []Tile     `hcl:"tile"`
	Skills    []Skill    `hcl:"skill"`
}

func LoadItems() error {
	// Read the file contents
	bytes, err := ioutil.ReadFile("data.hcl")
	if err != nil {
		return fmt.Errorf("Error loading config file: %s", err)
	}

	data, err := ParseItems(string(bytes))
	if err != nil {
		return err
	}

	skillData = make(map[string]Skill)
	for _, skill := range data.Skills {
		if _, ok := skillData[skill.Name]; ok {
			return fmt.Errorf("Error: got multiple sets of stats for skill: '%s'", skill.Name)
		}

		skillData[skill.Name] = skill
	}

	itemData = make(map[string]Item)
	for _, item := range data.Items {
		if _, ok := itemData[item.Name]; ok {
			return fmt.Errorf("Error: got multiple sets of stats for item: '%s'", item.Name)
		}
		for _, skill := range item.Skills {
			if _, ok := skillData[skill]; !ok {
				return fmt.Errorf("Error: item '%s' has unrecognized skill: '%s'", item.Name, skill)
			}
		}
		switch item.Slot {
		case "weapon":
			item.Type = Weapon
		case "off-hand":
			item.Type = OffHand
		case "armor":
			item.Type = Armor
		case "helm":
			item.Type = Helm
		case "accessory":
			item.Type = Accessory
		}
		itemData[item.Name] = item
	}

	creatureData = make(map[string]Creature)
	for _, creature := range data.Creatures {
		if _, ok := creatureData[creature.Name]; ok {
			return fmt.Errorf("Error: got multiple sets of stats for creature: '%s'", creature.Name)
		}

		creatureData[creature.Name] = creature
	}

	tileData = make(map[string]Tile)
	for _, tile := range data.Tiles {
		if _, ok := tileData[tile.Name]; ok {
			return fmt.Errorf("Error: got multiple sets of stats for tile: '%s'", tile.Name)
		}

		tileData[tile.Name] = tile
	}

	return nil
}

func ParseItems(raw string) (*Data, error) {
	var data Data
	if err := hcl.Decode(&data, raw); err != nil {
		return nil, err
	}

	return &data, nil
}

func GetItemData(name string) Item {
	return itemData[name]
}

func GetCreatureData(name string) Creature {
	return creatureData[name]
}

func GetTileData(name string) Tile {
	return tileData[name]
}

func GetSkillData(name string) Skill {
	return skillData[name]
}

func GetAllSkills() map[string]Skill {
	return skillData
}
