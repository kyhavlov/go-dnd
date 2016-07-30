package main

import (
	"engo.io/ecs"
	"engo.io/engo/common"
	"engo.io/engo"
	"log"
	"math/rand"
)

const TileWidth = 32
const MapWidth = 30
const MapHeight = 20

type Tile struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

type TileSystem struct {
	tiles [MapWidth][MapHeight]Tile
}

func createMap(render *common.RenderSystem) {
	sheet := common.NewSpritesheetFromFile("textures/dungeon.png", TileWidth, TileWidth)

	if sheet == nil {
		log.Println("Unable to load texture file")
	}

	for i := 0; i < MapWidth; i++ {
		for j := 0; j < MapHeight; j++ {
			tile := Tile{BasicEntity: ecs.NewBasic()}
			tile.SpaceComponent = common.SpaceComponent{
				Position: engo.Point{float32(i*TileWidth), float32(j*TileWidth)},
				Width:    TileWidth,
				Height:   TileWidth,
			}

			tile.RenderComponent = common.RenderComponent{
				Drawable: sheet.Cell(861 + rand.Intn(8)),
				Scale:    engo.Point{1, 1},
			}

			render.Add(&tile.BasicEntity, &tile.RenderComponent, &tile.SpaceComponent)
		}
	}
}