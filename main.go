package main

import (
	"engo.io/engo"
	"engo.io/ecs"
	"engo.io/engo/common"
)

type DungeonScene struct {}

// Type uniquely defines your game type
func (*DungeonScene) Type() string { return "dnd sim" }

// Preload is called before loading any assets from the disk,
// to allow you to register / queue them
func (*DungeonScene) Preload() {
	engo.Files.Load("textures/dungeon.png")
	engo.Files.Load("textures/dungeon2x.png")
}

// Setup is called before the main loop starts. It allows you
// to add entities and systems to your Scene.
func (*DungeonScene) Setup(world *ecs.World) {
	render := &common.RenderSystem{}

	input := &InputSystem{}
	createMap(render, input)

	world.AddSystem(render)
	world.AddSystem(&common.MouseSystem{})
	world.AddSystem(input)
}

func main() {
	opts := engo.RunOptions{
		Title: "Dragons and Dungeons",
		Width:  1200,
		Height: 800,
	}
	engo.Run(opts, &DungeonScene{})
}

