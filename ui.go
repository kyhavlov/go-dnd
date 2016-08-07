package main

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

type UiPanel struct {
	ecs.BasicEntity
	bg UiBackground
	textfields []UiText
	position engo.Point
	height float32
	width float32
}

type UiText struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

type UiBackground struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

func NewMouseCoordPanel(world *ecs.World) *UiPanel {

	position := engo.Point{24, 24}
	width := float32(264)
	height := float32(72)
	bgColor := color.RGBA{200, 153, 0, 125}

	// Create the panel background
	bg := UiBackground{
		BasicEntity: ecs.NewBasic(),
		RenderComponent: common.RenderComponent{Drawable: common.Rectangle{BorderWidth: 1, BorderColor: color.White}, Color: bgColor},
		SpaceComponent: common.SpaceComponent{Position: position, Width: width, Height: height},
	}

	// Set the background camera-independent
	bg.RenderComponent.SetZIndex(1) 			// zIndex > 0 (default)
	bg.RenderComponent.SetShader(common.HUDShader)

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&bg.BasicEntity, &bg.RenderComponent, &bg.SpaceComponent)
		}
	}

	panel := &UiPanel{
		BasicEntity : ecs.NewBasic(),
		bg: bg,
		textfields: make([]UiText, 0),
		position: position,
		height: height,
		width: width,
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
		Text: "Mouse X position is : ",
	}
	xCoord.SetShader(common.HUDShader)
	xCoord.SpaceComponent.Position.Set(position.X, position.Y + 12)
	xCoord.RenderComponent.SetZIndex(2)
	panel.textfields = append(panel.textfields,xCoord)

	yCoord := UiText{BasicEntity: ecs.NewBasic()}
	yCoord.RenderComponent.Drawable = common.Text{
		Font: fnt,
		Text: "Mouse Y position is : ",
	}
	yCoord.SetShader(common.HUDShader)
	yCoord.SpaceComponent.Position.Set(position.X, position.Y + 36)
	yCoord.RenderComponent.SetZIndex(2)
	panel.textfields = append(panel.textfields,yCoord)




	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&xCoord.BasicEntity, &xCoord.RenderComponent, &xCoord.SpaceComponent)
			sys.Add(&yCoord.BasicEntity, &yCoord.RenderComponent, &yCoord.SpaceComponent)
		}
	}

	return panel
}
