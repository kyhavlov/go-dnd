package structs

import (
	"reflect"
	"testing"
)

func TestParseItem(t *testing.T) {
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

	expected := Item{
		Name:   "Sapphire Staff",
		Slot:   "weapon",
		Icon:   2345,
		Skills: []string{"fireball", "ice-armor"},
		Requirements: StatComponent{
			Strength:     20,
			Dexterity:    21,
			Intelligence: 22,
		},
		Bonuses: StatComponent{
			MaxLife:      30,
			Strength:     10,
			Dexterity:    11,
			Intelligence: 12,
			MaxStamina:   40,
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

func TestParseCreature(t *testing.T) {
	raw := `
creature "Goblin" {
  icon = 2345
  skills = ["fireball", "ice-armor"]
  stats {
    life = 30
    str = 10
    dex = 11
    int = 12
    stamina = 40
    stamina_regen = 3
  }
}`

	expected := Creature{
		Name:         "Goblin",
		Icon:         2345,
		InnateSkills: []string{"fireball", "ice-armor"},
		StatComponent: StatComponent{
			MaxLife:      30,
			Strength:     10,
			Dexterity:    11,
			Intelligence: 12,
			MaxStamina:   40,
			StaminaRegen: 3,
		},
	}

	data, err := ParseItems(raw)
	if err != nil {
		t.Fatal(err)
	}

	if len(data.Creatures) != 1 {
		t.Fatalf("bad: %v", len(data.Creatures))
	}

	if !reflect.DeepEqual(data.Creatures[0], expected) {
		t.Fatalf("bad: \n%v\n%v", data.Creatures[0], expected)
	}
}

func TestParseTile(t *testing.T) {
	raw := `
tile "Floor" {
  icons = [1, 2, 3]
}`

	expected := Tile{
		Name:  "Floor",
		Icons: []int{1, 2, 3},
	}

	data, err := ParseItems(raw)
	if err != nil {
		t.Fatal(err)
	}

	if len(data.Tiles) != 1 {
		t.Fatalf("bad: %v", len(data.Tiles))
	}

	if !reflect.DeepEqual(data.Tiles[0], expected) {
		t.Fatalf("bad: \n%v\n%v", data.Tiles[0], expected)
	}
}

func TestParseSkill(t *testing.T) {
	raw := `
skill "Fireball" {
  icon = 2761

  min_range = 1
  max_range = 5

  damage = 10
  stamina_cost = 10

  damage_bonuses {
    int = 0.2
  }

  effects {
    hits_perpendicular = 1
  }
}`

	expected := Skill{
		Name:        "Fireball",
		Icon:        2761,
		MinRange:    1,
		MaxRange:    5,
		Damage:      10,
		StaminaCost: 10,
		DamageBonuses: StatModifiers{
			Int: 0.2,
		},
		Effects: map[string]int{"hits_perpendicular": 1},
	}

	data, err := ParseItems(raw)
	if err != nil {
		t.Fatal(err)
	}

	if len(data.Skills) != 1 {
		t.Fatalf("bad: %v", len(data.Skills))
	}

	if !reflect.DeepEqual(data.Skills[0], expected) {
		t.Fatalf("bad: \n%v\n%v", data.Skills[0], expected)
	}
}
