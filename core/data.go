package core

import (
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/kyhavlov/go-dnd/structs"
	"io/ioutil"
)

var itemData map[string]structs.Item

type Data struct {
	Items []structs.Item `hcl:"item"`
}

func LoadItems() error {
	// Read the file contents
	bytes, err := ioutil.ReadFile("items.hcl")
	if err != nil {
		return fmt.Errorf("Error loading config file: %s", err)
	}

	data, err := ParseItems(string(bytes))
	if err != nil {
		return err
	}

	itemData = make(map[string]structs.Item)
	for _, item := range data.Items {
		if _, ok := itemData[item.Name]; ok {
			return fmt.Errorf("Error: got multiple sets of stats for item: '%s'", item.Name)
		}
		switch item.Slot {
		case "weapon":
			item.Type = structs.Weapon
		case "off-hand":
			item.Type = structs.OffHand
		case "armor":
			item.Type = structs.Armor
		case "helm":
			item.Type = structs.Helm
		case "accessory":
			item.Type = structs.Accessory
		}
		itemData[item.Name] = item
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
