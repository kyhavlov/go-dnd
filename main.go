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
)

type DungeonScene struct {
	serverRoom *ServerRoom
	incoming chan NetworkMessage
	outgoing chan NetworkMessage
}

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
				Location: GridPoint{
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
	world.AddSystem(input)
	world.AddSystem(action)
	world.AddSystem(&MoveSystem{})
	world.AddSystem(&NetworkSystem{})
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
