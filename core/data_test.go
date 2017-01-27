package core

import (
	"github.com/kyhavlov/go-dnd/structs"
	"reflect"
	"testing"
)

func TestParseItems(t *testing.T) {
	raw := `
item "Sapphire Staff" {
  slot = "weapon"
  icon = 2345
  skills = ["fireball", "ice-armor"]
  reqs {
    str = 20
    dex = 21
    int = 22
  }
  bonus {
    life = 30
    str = 10
    dex = 11
    int = 12
    stamina = 40
    stamina_regen = 3
  }
}`

	expected := structs.Item{
		Name:   "Sapphire Staff",
		Slot:   "weapon",
		Icon:   2345,
		Skills: []string{"fireball", "ice-armor"},
		Requirements: structs.StatComponent{
			Strength:     20,
			Dexterity:    21,
			Intelligence: 22,
		},
		Bonuses: structs.StatComponent{
			MaxLife:      30,
			Strength:     10,
			Dexterity:    11,
			Intelligence: 12,
			Stamina:      40,
			StaminaRegen: 3,
		},
	}

	data, err := ParseItems(raw)
	if err != nil {
		t.Fatal(err)
	}

	if len(data.Items) != 1 {
		t.Fatalf("bad: %v", len(data.Items))
	}

	if !reflect.DeepEqual(data.Items[0], expected) {
		t.Fatalf("bad: \n%v\n%v", data.Items[0], expected)
	}
}
