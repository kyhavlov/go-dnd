package core

import (
	"net"
	"os"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	log "github.com/Sirupsen/logrus"
	"github.com/kyhavlov/go-dnd/structs"
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
	engo.Files.Load(structs.SpritesheetPath)

	// Load a font
	if err := engo.Files.Load("fonts/Gamegirl.ttf"); err != nil {
		panic(err)
	}

	if err := LoadItems(); err != nil {
		panic(err)
	}
}

// Setup is called before the main loop starts. It allows you
// to add entities and systems to your Scene.
func (scene *DungeonScene) Setup(world *ecs.World) {
	render := &common.RenderSystem{}

	mapSystem := &MapSystem{}
	turn := &TurnSystem{}
	input := &InputSystem{
		mapSystem: mapSystem,
		outgoing:  scene.outgoing,
		turn:      turn,
	}

	ui := &UiSystem{
		input: input,
	}

	event := &EventSystem{
		world: world,

		eventHistory: make([]Event, 0),

		incoming:   scene.incoming,
		outgoing:   scene.outgoing,
		serverRoom: scene.serverRoom,
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
	world.AddSystem(&NetworkSystem{})
	world.AddSystem(mapSystem)
	world.AddSystem(&LightSystem{})
	world.AddSystem(turn)
	world.AddSystem(&HealthSystem{})
}

// If we're the server, initialize a new server room and start listening for connections.
// Then, hook the server's incoming channel to both our scene's outgoing and incoming channels
// so that we can send our own actions directly to the server's input channel
func (scene *DungeonScene) Start() {
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
		gameStart := <-client.incoming
		client.incoming <- gameStart
		scene.incoming = client.incoming
		scene.outgoing = client.outgoing
	}
}
