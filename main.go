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

// The main scene for the game, representing a level
type DungeonScene struct {
	// The server room to use if we're the server, nil if we're not
	serverRoom *ServerRoom

	// Channels to send/receive network messages
	incoming chan NetworkMessage
	outgoing chan NetworkMessage
}

// A unique identifier for the scene
func (*DungeonScene) Type() string { return "dnd" }

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

	mapSystem := &MapSystem{}
	input := &InputSystem{
		mapSystem: mapSystem,
		outgoing:  scene.outgoing,
	}

	ui := &UiSystem{
		input: input,
	}

	event := &EventSystem{
		world: world,

		eventHistory: make([]Event, 0),

		incoming: scene.incoming,
		outgoing: scene.outgoing,
	}

	// Generate the map and add our player if we're the server, otherwise just ask the server to
	// make a player for us. In the future the player entities should probably get created in a loop
	// at world creation because it'll be launched from a lobby and we'll have the info to do that.
	if scene.serverRoom != nil {
		event.serverRoom = scene.serverRoom
		SendMessage(input.outgoing, NetworkMessage{
			Events: GenerateMap(3434323421999),
		})

		SendMessage(input.outgoing, NetworkMessage{
			Events: []Event{&NewPlayerEvent{
				GridPoint: GridPoint{
					X: 6,
					Y: 4,
				},
			}},
		})
		log.Info("created map")
	} else {
		SendMessage(input.outgoing, NetworkMessage{
			NewPlayer: true,
		})
	}

	world.AddSystem(render)

	// Add our input systems (mouse/camera/camera scroll/input)
	world.AddSystem(&common.MouseSystem{})
	engo.Input.RegisterAxis(engo.DefaultHorizontalAxis, engo.AxisKeyPair{engo.A, engo.D})
	engo.Input.RegisterAxis(engo.DefaultVerticalAxis, engo.AxisKeyPair{engo.W, engo.S})
	world.AddSystem(common.NewKeyboardScroller(400, engo.DefaultHorizontalAxis, engo.DefaultVerticalAxis))
	world.AddSystem(&common.MouseZoomer{-0.125})
	world.AddSystem(input)
	world.AddSystem(ui)

	// Add the game logic systems (event/move/network/map)
	world.AddSystem(event)
	world.AddSystem(&MoveSystem{})
	world.AddSystem(&NetworkSystem{})
	world.AddSystem(mapSystem)
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

	// Register our different Event implementations with the gob package for serialization
	// it's probably worth trying to keep the number of different events low for simplicity
	gob.Register(&MoveEvent{})
	gob.Register(&SetPlayerEvent{})
	gob.Register(&NewPlayerEvent{})
	gob.Register(&NewTileEvent{})
	gob.Register(&InitMapEvent{})
	gob.Register(&NewCreatureEvent{})

	scene := &DungeonScene{}

	// If we're the server, initialize a new server room and start listening for connections.
	// Then, hook the server's incoming channel to both our scene's outgoing and incoming channels
	// so that we can send our own actions directly to the server's input channel
	if len(os.Args) > 1 && os.Args[1] == "server" {
		serverRoom := StartServer(":8999")
		scene.incoming = serverRoom.incoming
		scene.outgoing = serverRoom.incoming
		scene.serverRoom = serverRoom
	} else {
		// If we're not a server, make a client and use its incoming/outgoing channels for the scene
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
