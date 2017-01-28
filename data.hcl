// Items
item "Sapphire Staff" {
  slot = "weapon"
  icon = 1734
  skills = ["Fireball", "Cleave"]
  bonus {
    int = 12
    stamina_regen = 3
  }
}

item "Leather Armor" {
  slot = "armor"
  icon = 1378
  bonus {
    life = 10
    stamina_regen = 1
  }
}

// Creatures
creature "Player" {
  icon = 594

  stats {
    life = 20
    str = 12
    dex = 12
    int = 12
    stamina = 50
  }
}

creature "Skeleton" {
  icon = 533

  stats {
    life = 20
    str = 12
    dex = 12
    int = 12
    stamina = 50
  }
}

// Tiles
tile "Dungeon Floor" {
  icons = [861, 862, 863, 864, 865, 866, 867, 868]
}

// Skills
skill "Basic Attack" {
  icon = 3010

  min_range = 1
  max_range = 1

  damage = 10
  cost = 5

  damage_bonuses {
    str = 0.1
  }
}

skill "Fireball" {
  icon = 2761

  min_range = 1
  max_range = 5

  damage = 10
  cost = 10

  damage_bonuses {
    int = 0.2
  }
}

skill "Cleave" {
  icon = 2753

  min_range = 1
  max_range = 1

  damage = 10
  cost = 5

  damage_bonuses {
    str = 0.1
  }

  effects {
    hits_perpendicular = 1
  }
}