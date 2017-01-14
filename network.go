package main

import (
	"encoding/gob"
	"engo.io/ecs"
	log "github.com/Sirupsen/logrus"
	"net"
)

// Unique identifiers for referring to objects/players over the network
type PlayerID uint64
type NetworkID uint64

type NetworkMessage struct {
	Sender    PlayerID
	NewPlayer bool

	Events []Event
}

type NetworkSystem struct {
	networkIdCounter NetworkID
	myPlayerId       int
}

func (ns *NetworkSystem) nextId() NetworkID {
	ns.networkIdCounter += 1
	return ns.networkIdCounter
}
func (ns *NetworkSystem) Update(dt float32)             {}
func (ns *NetworkSystem) Remove(entity ecs.BasicEntity) {}

type Client struct {
	incoming chan NetworkMessage
	outgoing chan NetworkMessage
	reader   *gob.Decoder
	writer   *gob.Encoder
}

func (client *Client) Read() {
	for {
		message := NetworkMessage{}
		err := client.reader.Decode(&message)
		if err != nil {
			log.Errorf("Error reading from client connection: %s", err)
			break
		}
		client.incoming <- message
	}
}

func (client *Client) Write() {
	for data := range client.outgoing {
		err := client.writer.Encode(data)
		if err != nil {
			log.Errorf("Error writing to connection: %s", err)
		}
	}
}

func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}

func NewClient(connection net.Conn) *Client {
	writer := gob.NewEncoder(connection)
	reader := gob.NewDecoder(connection)

	client := &Client{
		incoming: make(chan NetworkMessage, 256),
		outgoing: make(chan NetworkMessage, 256),
		reader:   reader,
		writer:   writer,
	}

	client.Listen()

	return client
}

type ServerRoom struct {
	idInc    PlayerID
	clients  map[PlayerID]*Client
	joins    chan net.Conn
	incoming chan NetworkMessage
}

func (room *ServerRoom) SendToClient(pid PlayerID, message NetworkMessage) {
	room.clients[pid].outgoing <- message
}

func (room *ServerRoom) SendToAllClients(message NetworkMessage) {
	for _, client := range room.clients {
		client.outgoing <- message
	}
}

func (room *ServerRoom) Join(connection net.Conn) {
	client := NewClient(connection)
	room.idInc += 1
	id := room.idInc
	room.clients[id] = client
	go func() {
		for {
			message := <-client.incoming
			//log.Debugf("[server] new message from ", connection.RemoteAddr())
			message.Sender = id
			room.incoming <- message
		}
	}()
}

func newServerRoom() *ServerRoom {
	room := &ServerRoom{
		clients:  make(map[PlayerID]*Client),
		joins:    make(chan net.Conn, 0),
		incoming: make(chan NetworkMessage, 256),
	}

	return room
}

func runServer(listener net.Listener, room *ServerRoom, players int) {
	for i := 0; i < players; i++ {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("[server] Error accepting connection: %s", err)
		}
		log.Info("[server] new client connected from ", conn.RemoteAddr())
		room.Join(conn)
	}

	// Send the game start event and create players/assign player IDs
	events := []Event{GameStartEvent{
		RandomSeed: 3434323421999,
	}}
	for i := 0; i < players+1; i++ {
		events = append(events, &NewPlayerEvent{
			PlayerID: PlayerID(i),
			GridPoint: GridPoint{
				X: 6 + i,
				Y: 4,
			},
		})
	}
	room.incoming <- NetworkMessage{
		Events: events,
	}
	for i := 1; i <= players; i++ {
		room.SendToClient(PlayerID(i), NetworkMessage{
			Events: []Event{&SetPlayerEvent{PlayerID(i)}},
		})
	}
}

func StartServer(address string) *ServerRoom {
	room := newServerRoom()

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Errorf("[server] Error binding on %s %s", address, err)
	} else {
		log.Infof("Hosting server at %v", listener.Addr())
	}

	runServer(listener, room, 1)

	return room
}
