package main

import (
	"engo.io/ecs"
	"engo.io/engo/common"
)

type Player struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}