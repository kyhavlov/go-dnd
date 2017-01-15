package structs

import (
	"engo.io/ecs"
	"engo.io/engo/common"
)

type NetworkID uint64

type Creature struct {
	ecs.BasicEntity
	NetworkID
	HealthComponent
	common.SpaceComponent
	common.RenderComponent
}

type HealthComponent struct {
	Life int
	Dead bool
}
