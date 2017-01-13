package main

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"

	log "github.com/Sirupsen/logrus"
)

type Creature struct {
	ecs.BasicEntity
	NetworkID
	HealthComponent
	common.SpaceComponent
	common.RenderComponent
	LightSource
}

func AddNewCreature(w *ecs.World, coords GridPoint, life int) {
	sheet := common.NewSpritesheetFromFile(SpritesheetPath, TileWidth, TileWidth)

	if sheet == nil {
		log.Fatalf("Unable to load texture file")
	}

	creature := Creature{}
	creature.BasicEntity = ecs.NewBasic()
	creature.SpaceComponent = common.SpaceComponent{
		Position: coords.toPixels(),
		Width:    TileWidth,
		Height:   TileWidth,
	}
	creature.RenderComponent = common.RenderComponent{
		Drawable: sheet.Cell(533),
		Scale:    engo.Point{1, 1},
	}
	creature.HealthComponent = HealthComponent{
		Life: life,
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *NetworkSystem:
			creature.NetworkID = sys.nextId()
		}
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&creature.BasicEntity, &creature.RenderComponent, &creature.SpaceComponent)
		case *MoveSystem:
			sys.Add(&creature.BasicEntity, &creature.SpaceComponent, creature.NetworkID)
		case *HealthSystem:
			sys.Add(&creature.BasicEntity, &creature.HealthComponent)
		}
	}
}
