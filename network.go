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

	Actions   []Action
}

func SendMessage(channel chan NetworkMessage, message NetworkMessage) {
	channel <- message
}

type NetworkSystem struct {
	networkIdCounter NetworkID
	myPlayerId int
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
		incoming: make(chan NetworkMessage),
		outgoing: make(chan NetworkMessage),
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

func (room *ServerRoom) NextId() PlayerID {
	room.idInc += 1
	return room.idInc
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
	id := room.NextId()
	room.clients[id] = client
	go func() {
		for {
			message := <-client.incoming
			log.Info("[server] new message from ", connection.RemoteAddr())
			message.Sender = id
			room.incoming <- message
		}
	}()
}

func (room *ServerRoom) Listen() {
	go func() {
		for {
			conn := <-room.joins
			log.Info("[server] new client connected from ", conn.RemoteAddr())
			room.Join(conn)
		}
	}()
}

func NewServerRoom() *ServerRoom {
	room := &ServerRoom{
		clients:  make(map[PlayerID]*Client),
		joins:    make(chan net.Conn),
		incoming: make(chan NetworkMessage),
	}

	room.Listen()

	return room
}

func runServer(listener net.Listener, room *ServerRoom) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("[server] Error accepting connection: %s", err)
		}
		room.joins <- conn
	}
}

func startServer() *ServerRoom {
	room := NewServerRoom()

	listener, err := net.Listen("tcp", ":8999")
	if err != nil {
		log.Errorf("[server] Error binding on port 8999: %s", err)
	}

	go runServer(listener, room)

	return room
}
