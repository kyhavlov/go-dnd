package main

import (
	"fmt"
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

type UiPanel struct {
	ecs.BasicEntity
	bg         UiBackground
	textfields []UiText
	position   engo.Point
	height     float32
	width      float32
}

type UiText struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
	DynamicTextComponent
}

type UiBackground struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

type UiSystem struct {
	uiTexts map[*ecs.BasicEntity]*UiText
	input *InputSystem
}

type DynamicTextComponent struct {
	updateFunc func() string
	lastValue string
}

func (us *UiSystem) Update(dt float32) {
	for _, text := range us.uiTexts {
		if text.updateFunc != nil {
			newValue := text.updateFunc()
			if newValue != text.lastValue {
				text.Drawable = common.Text{
					Font: text.Drawable.(common.Text).Font,
					Text: newValue,
				}
			}
			text.lastValue = newValue
		}
	}
}

func (us *UiSystem) Add(e *ecs.BasicEntity, text *UiText) {
	us.uiTexts[e] = text
}

func (us *UiSystem) Remove(entity ecs.BasicEntity) {
	delete(us.uiTexts, &entity)
}

// New is the initialisation of the System
func (us *UiSystem) New(w *ecs.World) {
	us.uiTexts = make(map[*ecs.BasicEntity]*UiText)

	NewMouseCoordPanel(us, w)
}

func NewMouseCoordPanel(uiSystem *UiSystem, world *ecs.World) {

	position := engo.Point{24, 24}
	width := float32(264)
	height := float32(72)
	bgColor := color.RGBA{200, 153, 0, 125}

	// Create the panel background
	bg := UiBackground{
		BasicEntity:     ecs.NewBasic(),
		RenderComponent: common.RenderComponent{Drawable: common.Rectangle{BorderWidth: 1, BorderColor: color.White}, Color: bgColor},
		SpaceComponent:  common.SpaceComponent{Position: position, Width: width, Height: height},
	}

	// Set the background camera-independent
	bg.RenderComponent.SetZIndex(1) // zIndex > 0 (default)
	bg.RenderComponent.SetShader(common.HUDShader)

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&bg.BasicEntity, &bg.RenderComponent, &bg.SpaceComponent)
		}
	}

	// Add text fields

	// Set font
	fnt := &common.Font{
		URL:  "fonts/Roboto-Regular.ttf",
		FG:   color.White,
		Size: 24,
	}
	err := fnt.CreatePreloaded()
	if err != nil {
		panic(err)
	}

	xCoord := UiText{BasicEntity: ecs.NewBasic()}
	xCoord.RenderComponent.Drawable = common.Text{
		Font: fnt,
	}
	xCoord.SetShader(common.HUDShader)
	xCoord.SpaceComponent.Position.Set(position.X, position.Y+12)
	xCoord.RenderComponent.SetZIndex(2)

	xCoord.DynamicTextComponent = DynamicTextComponent{
		updateFunc: func() string {
			return fmt.Sprintf("Mouse X position is : %d", int(uiSystem.input.mouseTracker.MouseX))
		},
	}

	yCoord := UiText{BasicEntity: ecs.NewBasic()}
	yCoord.RenderComponent.Drawable = common.Text{
		Font: fnt,
	}
	yCoord.SetShader(common.HUDShader)
	yCoord.SpaceComponent.Position.Set(position.X, position.Y+36)
	yCoord.RenderComponent.SetZIndex(2)
	yCoord.DynamicTextComponent = DynamicTextComponent{
		updateFunc: func() string {
			return fmt.Sprintf("Mouse Y position is : %d", int(uiSystem.input.mouseTracker.MouseY))
		},
	}

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&xCoord.BasicEntity, &xCoord.RenderComponent, &xCoord.SpaceComponent)
			sys.Add(&yCoord.BasicEntity, &yCoord.RenderComponent, &yCoord.SpaceComponent)
		}
	}

	uiSystem.Add(&xCoord.BasicEntity, &xCoord)
	uiSystem.Add(&yCoord.BasicEntity, &yCoord)
}
