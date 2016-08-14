package main

import (
	"engo.io/ecs"
	log "github.com/Sirupsen/logrus"
)

type EventSystem struct {
	world *ecs.World

	eventHistory []Event

	incoming   chan NetworkMessage
	outgoing   chan NetworkMessage
	serverRoom *ServerRoom
}

type Event interface {
	Process(*ecs.World, float32) bool
}

// New is the initialisation of the System
func (ds *EventSystem) New(w *ecs.World) {}

func (ds *EventSystem) Update(dt float32) {
	select {
	case message, ok := <-ds.incoming:
		if ok {
			if message.NewPlayer && ds.serverRoom != nil {
				log.Info("sending event history to new player ", message.Sender)
				history := NetworkMessage{
					Events: ds.eventHistory,
				}
				setID := &SetPlayerEvent{message.Sender}
				newPlayer := &NewPlayerEvent{
					PlayerID: message.Sender,
					GridPoint: GridPoint{
						X: 6,
						Y: 4,
					},
				}
				history.Events = append(history.Events, setID)

				newPlayer.Process(ds.world, dt)
				ds.serverRoom.SendToClient(message.Sender, history)
				ds.serverRoom.SendToAllClients(NetworkMessage{
					Events: []Event{newPlayer},
				})
			}

			for _, event := range message.Events {
				event.Process(ds.world, dt)
				ds.eventHistory = append(ds.eventHistory, event)
				if ds.serverRoom != nil {
					ds.serverRoom.SendToAllClients(message)
				}
			}
		} else {
			log.Fatal("channel closed")
		}
	default:
	}
}

func (ds *EventSystem) Remove(entity ecs.BasicEntity) {}
