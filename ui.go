package main

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

// UI Framework
type UiPanel struct {
	ecs.BasicEntity
	bg         UiBackground
	uiElements []UiElement
	position   engo.Point
	height     float32
	width      float32
}

func (uipanel *UiPanel) UpdatePanel () {
	for _, e := range uipanel.uiElements {
		e.UpdateElement()
	}
}

type UiElement interface {
	UpdateElement()
}

type ElementUpdateFunc func(*UiElement)

type UiText struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
	updateFunc ElementUpdateFunc
}

func (uitext UiText) UpdateElement() {
	if uitext.updateFunc != nil {
		uitext.updateFunc(&uitext)
	}
}

type UiBackground struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

// UI SYSTEM
type UiSystem struct {
	panels []UiPanel
	inputSystem *InputSystem
}

//// Leaving this in for now since it might be a better way to initialize a system (used in render)
//func (us *UiSystem) New(w *ecs.World) {
//	w.AddSystem(&UiSystem{})
//}

func (us *UiSystem) Add(myPanel UiPanel) {
	us.panels = append(us.panels, myPanel)
}

func (us *UiSystem) Remove(e ecs.BasicEntity) {
	var del int = -1
	for index, entity := range us.panels {
		if entity.ID() == e.ID() {
			del = index
			break
		}
	}
	if del >= 0 {
		us.panels = append(us.panels[:del], us.panels[del+1:]...)
	}
}

func (us *UiSystem) Update(dt float32) {
	for _, p := range us.panels {
		p.UpdatePanel()
	}
}

func NewMouseCoordPanel(world *ecs.World, input *InputSystem) *UiPanel {

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

	panel := &UiPanel{
		BasicEntity: ecs.NewBasic(),
		bg:          bg,
		uiElements:  make([]UiElement, 0),
		position:    position,
		height:      height,
		width:       width,
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

	xCoord := &UiText{BasicEntity: ecs.NewBasic()}
	xCoord.RenderComponent.Drawable = common.Text{
		Font: fnt,
		Text: "Mouse X position is : ",
	}
	xCoord.SetShader(common.HUDShader)
	xCoord.SpaceComponent.Position.Set(position.X, position.Y+12)
	xCoord.RenderComponent.SetZIndex(2)

	// The updateFunc determines how xCoord is updated
	xCoord.updateFunc = func(){
		// Get the mousetracker
		x := input.mouseTracker.MouseX
		y := input.mouseTracker.MouseY

		xCoord.RenderComponent.Drawable.Text

	}

	panel.uiElements = append(panel.uiElements, xCoord)

	yCoord := &UiText{BasicEntity: ecs.NewBasic()}
	yCoord.RenderComponent.Drawable = common.Text{
		Font: fnt,
		Text: "Mouse Y position is : ",
	}
	yCoord.SetShader(common.HUDShader)
	yCoord.SpaceComponent.Position.Set(position.X, position.Y+36)
	yCoord.RenderComponent.SetZIndex(2)
	panel.uiElements = append(panel.uiElements, yCoord)

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&xCoord.BasicEntity, &xCoord.RenderComponent, &xCoord.SpaceComponent)
			sys.Add(&yCoord.BasicEntity, &yCoord.RenderComponent, &yCoord.SpaceComponent)
		}
	}

	return panel
}
