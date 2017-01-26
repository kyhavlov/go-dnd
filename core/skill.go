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
	"fireball": 2761,
}

func NewSkill(id string, sourceId structs.NetworkID, targetId structs.NetworkID) Skill {
	switch id {
	case "basicattack":
		return &BasicAttack{sourceId, targetId}
	case "fireball":
		return &Fireball{sourceId, targetId}
	}
	return nil
}

type BasicAttack struct {
	AttackerId       structs.NetworkID
	TargetId structs.NetworkID
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
	AttackerId       structs.NetworkID
	TargetId structs.NetworkID
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

