item "Sapphire Staff" {
  slot = "weapon"
  icon = 1734
  skills = ["fireball", "cleave"]
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

creature "Goblin" {
  life = 20
  str = 12
  dex = 12
  int = 12
  stamina = 50

  equipment = ["Sapphire Staff"]
  skills = ["firestorm"]
}

skill "fireball" {
  icon = 2761
  range = 5
  damage = 10
  int_modifier = 0.2
}