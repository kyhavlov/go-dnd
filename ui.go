package main

import (
	"fmt"
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	log "github.com/Sirupsen/logrus"
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
	input   *InputSystem
	render  *common.RenderSystem
}

type DynamicTextComponent struct {
	updateFunc func() string
	lastValue  string
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
	us.render.Add(e, &text.RenderComponent, &text.SpaceComponent)
}

func (us *UiSystem) Remove(entity ecs.BasicEntity) {
	delete(us.uiTexts, &entity)
}

// New is the initialisation of the System
func (us *UiSystem) New(w *ecs.World) {
	us.uiTexts = make(map[*ecs.BasicEntity]*UiText)
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			us.render = sys
		}
	}

	NewMouseCoordPanel(us, w)
}

func (us *UiSystem) InitUI(w *ecs.World, playerCount int) {
	// Add turn counters
	readyFont := &common.Font{
		URL:  "fonts/Gamegirl.ttf",
		FG:   color.White,
		Size: 12,
	}
	if err := readyFont.CreatePreloaded(); err != nil {
		panic(err)
	}
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *TurnSystem:
			log.Infof("player count for ui: %d", playerCount)
			for i := 0; i < playerCount; i++ {
				readyStatus := UiText{BasicEntity: ecs.NewBasic()}
				readyStatus.RenderComponent.Drawable = common.Text{
					Font: readyFont,
				}
				readyStatus.SetShader(common.HUDShader)
				readyStatus.SpaceComponent.Position.Set(24, float32(120+(i*18)))
				readyStatus.RenderComponent.SetZIndex(2)
				playerNum := i+1
				readyStatus.DynamicTextComponent = DynamicTextComponent{
					updateFunc: func() string {
						ready := sys.IsPlayerReady(PlayerID(playerNum-1))
						status := "Not Ready"
						readyStatus.RenderComponent.Color = color.White
						if ready {
							status = "Ready"
							readyStatus.RenderComponent.Color = color.RGBA{0, 255, 0, 120}
						}
						return fmt.Sprintf("Player %d: %v", playerNum, status)
					},
				}

				us.Add(&readyStatus.BasicEntity, &readyStatus)
			}
		}
	}
}

func NewMouseCoordPanel(uiSystem *UiSystem, world *ecs.World) {

	position := engo.Point{24, 24}
	width := float32(320)
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
		URL:  "fonts/Gamegirl.ttf",
		FG:   color.White,
		Size: 12,
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
	xCoord.SpaceComponent.Position.Set(position.X+10, position.Y+12)
	xCoord.RenderComponent.SetZIndex(2)

	xCoord.DynamicTextComponent = DynamicTextComponent{
		updateFunc: func() string {
			return fmt.Sprintf("Mouse X position is: %d", int(uiSystem.input.mouseTracker.MouseX))
		},
	}

	yCoord := UiText{BasicEntity: ecs.NewBasic()}
	yCoord.RenderComponent.Drawable = common.Text{
		Font: fnt,
	}
	yCoord.SetShader(common.HUDShader)
	yCoord.SpaceComponent.Position.Set(position.X+10, position.Y+36)
	yCoord.RenderComponent.SetZIndex(2)
	yCoord.DynamicTextComponent = DynamicTextComponent{
		updateFunc: func() string {
			return fmt.Sprintf("Mouse Y position is: %d", int(uiSystem.input.mouseTracker.MouseY))
		},
	}

	uiSystem.Add(&xCoord.BasicEntity, &xCoord)
	uiSystem.Add(&yCoord.BasicEntity, &yCoord)
}
