package core

import (
	"engo.io/ecs"
	"image/color"

	"github.com/engoengine/math/imath"
	//log "github.com/Sirupsen/logrus"
	"engo.io/engo/common"
	"github.com/kyhavlov/go-dnd/structs"
)

const LIGHT_DECREASE = 20

type LightSystem struct {
	mapSystem *MapSystem
	lights    map[*ecs.BasicEntity]LightSource

	needsUpdate bool
	timer       float32
}

type LightSource interface {
	GetLocation() structs.GridPoint
	GetBrightness() uint8
}

type BasicLightSource struct {
	structs.GridPoint

	// The starting brightness alpha value. 255 is full brightness
	Brightness uint8
}

func (b *BasicLightSource) GetLocation() structs.GridPoint { return b.GridPoint }
func (b *BasicLightSource) GetBrightness() uint8           { return b.Brightness }

type DynamicLightSource struct {
	spaceComponent *common.SpaceComponent
	Brightness     uint8
}

func (d *DynamicLightSource) GetLocation() structs.GridPoint {
	return structs.PointToGridPoint(d.spaceComponent.Position)
}
func (d *DynamicLightSource) GetBrightness() uint8 { return d.Brightness }

// New is the initialisation of the System
func (ls *LightSystem) New(w *ecs.World) {
	ls.lights = make(map[*ecs.BasicEntity]LightSource)

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *MapSystem:
			ls.mapSystem = sys
		}
	}

	e := ecs.NewBasic()
	ls.Add(&e, &BasicLightSource{
		GridPoint:  structs.GridPoint{18, 3},
		Brightness: 250,
	})

	ls.needsUpdate = true
}

func (ls *LightSystem) Update(dt float32) {
	if ls.needsUpdate {
		// Reset the light levels to minimum before recalculating
		for _, row := range ls.mapSystem.Tiles {
			for _, tile := range row {
				if tile != nil {
					tile.Color = color.Alpha{structs.MinBrightness}
				}
			}
		}

		// Increase the light of the tiles around the source in a diamond pattern,
		// with the light strength fading with distance from the source.
		for _, light := range ls.lights {
			radius := int((light.GetBrightness()-structs.MinBrightness)/LIGHT_DECREASE) + 1
			//log.Infof("radius: %d", radius)
			for i := 0; i <= radius*2; i++ {
				current := light.GetLocation()
				current.X -= radius
				current.Y = (current.Y - radius) + i

				for j := 0; j <= radius*2; j++ {
					//log.Infof("%d, %d distance to %d, %d: %d", current.X, current.Y, light.X, light.Y, current.distanceTo(&light.GridPoint))
					if current.X >= 0 && current.X <= ls.mapSystem.MapWidth() && current.Y >= 0 && current.Y <= ls.mapSystem.MapHeight() {
						dist := current.DistanceTo(light.GetLocation())

						if dist <= radius {
							if tile := ls.mapSystem.Tiles[current.X][current.Y]; tile != nil {
								lightStrength := (radius - dist) * LIGHT_DECREASE
								//log.Infof("lights at %d,%d updated to %d", current.X, current.Y, int(tile.Color.(color.Alpha).A) + lightStrength)
								tile.Color = color.Alpha{uint8(imath.Min(int(tile.Color.(color.Alpha).A)+lightStrength, 250))}
							}
						}
					}
					current.X += 1
				}
			}
		}

		//log.Info("lighting updated")

		ls.needsUpdate = false
	}
}

func (ls *LightSystem) Add(e *ecs.BasicEntity, light LightSource) {
	ls.lights[e] = light
	ls.needsUpdate = true
}

func (ls *LightSystem) Remove(entity ecs.BasicEntity) {
	delete(ls.lights, &entity)
}
