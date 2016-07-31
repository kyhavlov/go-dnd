package main

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	log "github.com/Sirupsen/logrus"
	"math/rand"
)

const SpritesheetPath = "textures/dungeon2x.png"
const TileWidth = 64
const MapWidth = 30
const MapHeight = 20

type Tile struct {
	ecs.BasicEntity
	NewTileAction
}

type GridPoint struct {
	X int
	Y int
}

func (gp *GridPoint) toPixels() engo.Point {
	return engo.Point{float32(gp.X * TileWidth), float32(gp.Y * TileWidth)}
}

func createMap(render *common.RenderSystem, input *InputSystem) NetworkMessage {
	tileCreations := make([]Action, 0)

	for i := 0; i < MapWidth; i++ {
		for j := 0; j < MapHeight; j++ {
			newTile := &NewTileAction{
				SpaceComponent: common.SpaceComponent{
					Position: engo.Point{float32(i * TileWidth), float32(j * TileWidth)},
					Width:    0,
					Height:   0,
				},
				Sprite: 861 + rand.Intn(8),
			}

			tileCreations = append(tileCreations, newTile)
		}
	}

	return NetworkMessage{Actions: tileCreations}
}

type NewTileAction struct {
	common.RenderComponent
	common.SpaceComponent
	Sprite int
}

func (action *NewTileAction) Process(w *ecs.World, dt float32) bool {
	sheet := common.NewSpritesheetFromFile(SpritesheetPath, TileWidth, TileWidth)

	if sheet == nil {
		log.Fatalf("Unable to load texture file")
	}

	tile := Tile{}
	tile.BasicEntity = ecs.NewBasic()
	tile.SpaceComponent = action.SpaceComponent
	tile.RenderComponent = common.RenderComponent{
		Drawable: sheet.Cell(action.Sprite),
		Scale:    engo.Point{1, 1},
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&tile.BasicEntity, &tile.RenderComponent, &tile.SpaceComponent)
		}
	}

	return true
}
