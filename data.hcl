// Items
item "Sapphire Staff" {
  slot = "weapon"
  icon = 1734
  skills = ["Fireball", "Ice Storm"]
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

item "Ice Spear" {
  slot = "weapon"
  icon = 1836
  skills = ["Frozen Lance", "Cleave"]
  increases_melee_range = true
  bonus {
    int = 10
    str = 5
  }
}

// Creatures
creature "Player" {
  icon = 594

  stats {
    move = 8
    life = 40
    str = 13
    dex = 13
    int = 13
    stamina = 50
    stamina_regen = 3
  }
}

creature "Skeleton" {
  icon = 533

  stats {
    move = 5
    life = 20
    str = 12
    dex = 12
    int = 12
    stamina = 30
    stamina_regen = 3
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
  stamina_cost = 5

  damage_bonuses {
    str = 0.1
  }

  tags = ["melee"]
}

skill "Fireball" {
  icon = 2761

  min_range = 1
  max_range = 5

  damage = 10
  stamina_cost = 10

  damage_bonuses {
    int = 0.5
  }
}

skill "Cleave" {
  icon = 2753

  min_range = 1
  max_range = 1
  targets_ground = true

  damage = 10
  stamina_cost = 10

  damage_bonuses {
    str = 0.3
  }

  effects {
    hits_perpendicular = 1
  }
}

skill "Ice Storm" {
  icon = 2776

  min_range = 0
  max_range = 5
  targets_ground = true

  damage = 10
  stamina_cost = 15

  damage_bonuses {
    int = 0.15
  }

  effects {
    aoe_radius = 1
  }
}

skill "Frozen Lance" {
  icon = 2781

  min_range = 1
  max_range = 1

  damage = 10
  stamina_cost = 12

  damage_bonuses {
    int = 0.2
  }

  effects {
    pierces = 1
  }

  tags = ["melee"]
}