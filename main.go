package main

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"net"
	"os"

	"encoding/gob"
	log "github.com/Sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	//"engo.io/engo/demos/demoutils"
	//"image/color"
)

type DungeonScene struct {
	serverRoom *ServerRoom
	incoming chan NetworkMessage
	outgoing chan NetworkMessage
}

// A unique identifier for the scene
func (*DungeonScene) Type() string { return "dnd sim" }

// Preload is called before loading any assets from the disk,
// to allow you to register / queue them
func (*DungeonScene) Preload() {
	engo.Files.Load(SpritesheetPath)

	// Load a font
	err := engo.Files.Load("fonts/Roboto-Regular.ttf")
	if err != nil {
		panic(err)
	}
}

// Setup is called before the main loop starts. It allows you
// to add entities and systems to your Scene.
func (scene *DungeonScene) Setup(world *ecs.World) {
	render := &common.RenderSystem{}

	input := &InputSystem{
		outgoing: scene.outgoing,
	}
	action := &ActionSystem{
		world: world,

		actionHistory: make([]Action, 0),

		incoming: scene.incoming,
		outgoing: scene.outgoing,
	}

	if scene.serverRoom != nil {
		action.serverRoom = scene.serverRoom
		go SendMessage(input.outgoing, NetworkMessage{
			Actions: []Action{&NewPlayerAction{
				GridPoint: GridPoint{
					X: 6,
					Y: 4,
				},
			}},
		})

		go SendMessage(input.outgoing, createMap(render, input))
		log.Info("created map")
	} else {
		go SendMessage(input.outgoing, NetworkMessage{
			NewPlayer: true,
		})
	}

	world.AddSystem(render)

	world.AddSystem(&common.MouseSystem{})
	engo.Input.RegisterAxis(engo.DefaultHorizontalAxis, engo.AxisKeyPair{engo.A, engo.D})
	engo.Input.RegisterAxis(engo.DefaultVerticalAxis, engo.AxisKeyPair{engo.W, engo.S})
	world.AddSystem(common.NewKeyboardScroller(400, engo.DefaultHorizontalAxis, engo.DefaultVerticalAxis))
	world.AddSystem(&common.MouseZoomer{-0.125})
	world.AddSystem(input)

	world.AddSystem(action)
	world.AddSystem(&MoveSystem{})
	world.AddSystem(&NetworkSystem{})

	NewMouseCoordPanel(world)
}

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

	gob.Register(&MoveAction{})
	gob.Register(&SetPlayerAction{})
	gob.Register(&NewPlayerAction{})
	gob.Register(&NewTileAction{})

	scene := &DungeonScene{}
	if len(os.Args) > 1 && os.Args[1] == "server" {
		serverRoom := startServer()
		scene.incoming = serverRoom.incoming
		scene.outgoing = serverRoom.incoming
		scene.serverRoom = serverRoom
		log.Info("Hosting server")
	} else {
		conn, err := net.Dial("tcp", "127.0.0.1:8999")
		if err != nil {
			log.Fatalf("Error connecting to server: %s", err)
		} else {
			log.Info("Connected to server at ", conn.RemoteAddr())
		}
		client := NewClient(conn)
		scene.incoming = client.incoming
		scene.outgoing = client.outgoing
	}

	engo.Run(opts, scene)
}
