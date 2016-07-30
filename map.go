package main

import (
	"engo.io/ecs"
	"engo.io/engo/common"
	"engo.io/engo"
	"log"
	"math/rand"
)

const TileWidth = 64
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

func TileToPoint(x, y int) engo.Point {
	return engo.Point{float32(x*TileWidth), float32(y*TileWidth)}
}

func createMap(render *common.RenderSystem, input *InputSystem) {
	sheet := common.NewSpritesheetFromFile("textures/dungeon2x.png", TileWidth, TileWidth)

	if sheet == nil {
		log.Println("Unable to load texture file")
	}

	player := Player{BasicEntity: ecs.NewBasic()}
	player.SpaceComponent = common.SpaceComponent{
		Position: TileToPoint(3, 6),
		Width:    TileWidth,
		Height:   TileWidth,
	}
	player.RenderComponent = common.RenderComponent{
		Drawable: sheet.Cell(132),
		Scale:    engo.Point{1, 1},
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

	render.Add(&player.BasicEntity, &player.RenderComponent, &player.SpaceComponent)
	input.player = &player
}