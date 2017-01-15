package main

import (
	"engo.io/engo"

	log "github.com/Sirupsen/logrus"
	"github.com/kyhavlov/go-dnd/core"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func main() {
	// Set up logging
	formatter := new(prefixed.TextFormatter)
	formatter.ForceColors = true

	log.SetFormatter(formatter)
	log.SetLevel(log.DebugLevel)

	opts := engo.RunOptions{
		Title:  "Dragons and Dungeons",
		Width:  1200,
		Height: 800,
	}

	// Register the types of network message that will be sent
	core.RegisterEvents()

	scene := &core.DungeonScene{}
	scene.Start()

	engo.Run(opts, scene)
}
