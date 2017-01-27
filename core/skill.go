package core

import (
	log "github.com/Sirupsen/logrus"
	"github.com/kyhavlov/go-dnd/structs"
)

type Skill interface {
	TargetIsValid(*MapSystem) bool
	PerformSkillActions(*MapSystem)
	GetSourceAndTarget() (structs.NetworkID, structs.NetworkID)
}

var skillIcons = map[string]int{
	"basicattack": 3010,
	"fireball":    2761,
	"cleave":      2753,
}

func NewSkill(id string, sourceId structs.NetworkID, targetId structs.NetworkID) Skill {
	switch id {
	case "basicattack":
		return &BasicAttack{sourceId, targetId}
	case "fireball":
		return &Fireball{sourceId, targetId}
	case "cleave":
		return &Cleave{sourceId, targetId}
	}
	return nil
}

type BasicAttack struct {
	AttackerId structs.NetworkID
	TargetId   structs.NetworkID
}

func (s *BasicAttack) TargetIsValid(sys *MapSystem) bool {
	attacker := sys.Creatures[s.AttackerId]
	target := sys.Creatures[s.TargetId]

	a := structs.PointToGridPoint(attacker.SpaceComponent.Position)
	b := structs.PointToGridPoint(target.SpaceComponent.Position)

	return a.DistanceTo(b) == 1
}

func (s *BasicAttack) PerformSkillActions(sys *MapSystem) {
	creature := sys.Creatures[s.TargetId]
	creature.Life -= 20
	log.Infof("Creature id %d took %d damage, at %d life now", s.TargetId, 20, creature.Life)
}

func (s *BasicAttack) GetSourceAndTarget() (structs.NetworkID, structs.NetworkID) {
	return s.AttackerId, s.TargetId
}

// Fireball
type Fireball struct {
	AttackerId structs.NetworkID
	TargetId   structs.NetworkID
}

func (s *Fireball) TargetIsValid(sys *MapSystem) bool {
	attacker := sys.Creatures[s.AttackerId]
	target := sys.Creatures[s.TargetId]

	a := structs.PointToGridPoint(attacker.SpaceComponent.Position)
	b := structs.PointToGridPoint(target.SpaceComponent.Position)

	return a.DistanceTo(b) <= 5
}

func (s *Fireball) PerformSkillActions(sys *MapSystem) {
	attacker := sys.Creatures[s.AttackerId]
	damage := 10 + (attacker.Intelligence / 5)
	creature := sys.Creatures[s.TargetId]
	creature.Life -= damage
	log.Infof("Creature id %d took %d fireball damage, at %d life now", s.TargetId, damage, creature.Life)
}

func (s *Fireball) GetSourceAndTarget() (structs.NetworkID, structs.NetworkID) {
	return s.AttackerId, s.TargetId
}

// Cleave
type Cleave struct {
	AttackerId structs.NetworkID
	TargetId   structs.NetworkID
}

func (s *Cleave) TargetIsValid(sys *MapSystem) bool {
	attacker := sys.Creatures[s.AttackerId]
	target := sys.Creatures[s.TargetId]

	a := structs.PointToGridPoint(attacker.SpaceComponent.Position)
	b := structs.PointToGridPoint(target.SpaceComponent.Position)

	return a.DistanceTo(b) == 1
}

func (s *Cleave) PerformSkillActions(sys *MapSystem) {
	attacker := sys.Creatures[s.AttackerId]
	primaryTarget := sys.Creatures[s.TargetId]
	a := structs.PointToGridPoint(attacker.SpaceComponent.Position)
	b := structs.PointToGridPoint(primaryTarget.SpaceComponent.Position)

	targets := []*structs.Creature{primaryTarget}
	if a.X == b.X {
		left := sys.GetCreatureAt(structs.GridPoint{b.X - 1, b.Y})
		right := sys.GetCreatureAt(structs.GridPoint{b.X + 1, b.Y})

		if left != nil {
			targets = append(targets, left)
		}
		if right != nil {
			targets = append(targets, right)
		}
	} else {
		top := sys.GetCreatureAt(structs.GridPoint{b.X, b.Y - 1})
		bottom := sys.GetCreatureAt(structs.GridPoint{b.X, b.Y + 1})

		if top != nil {
			targets = append(targets, top)
		}
		if bottom != nil {
			targets = append(targets, bottom)
		}
	}

	for _, target := range targets {
		damage := 10 + (attacker.Strength / 10)
		target.Life -= damage
		log.Infof("Creature id %d took %d Cleave damage, at %d life now", target.NetworkID, damage, target.Life)
	}
}

func (s *Cleave) GetSourceAndTarget() (structs.NetworkID, structs.NetworkID) {
	return s.AttackerId, s.TargetId
}
