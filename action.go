package main

import (
	"engo.io/ecs"
	log "github.com/Sirupsen/logrus"
)

type ActionSystem struct {
	world *ecs.World

	actionHistory []Action

	incoming   chan NetworkMessage
	outgoing   chan NetworkMessage
	serverRoom *ServerRoom
}

type Action interface {
	Process(*ecs.World, float32) bool
}

// New is the initialisation of the System
func (ds *ActionSystem) New(w *ecs.World) {}

func (ds *ActionSystem) Update(dt float32) {
	select {
	case message, ok := <-ds.incoming:
		if ok {
			if message.NewPlayer && ds.serverRoom != nil {
				log.Info("sending action history to new player ", message.Sender)
				history := NetworkMessage{
					Actions: ds.actionHistory,
				}
				setID := &SetPlayerAction{message.Sender}
				newPlayer := &NewPlayerAction{
					PlayerID: message.Sender,
					GridPoint: GridPoint{
						X: 6,
						Y: 4,
					},
				}
				history.Actions = append(history.Actions, setID)

				newPlayer.Process(ds.world, dt)
				ds.serverRoom.SendToClient(message.Sender, history)
				ds.serverRoom.SendToAllClients(NetworkMessage{
					Actions: []Action{newPlayer},
				})
			}

			for _, action := range message.Actions {
				action.Process(ds.world, dt)
				ds.actionHistory = append(ds.actionHistory, action)
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

func (ds *ActionSystem) Remove(entity ecs.BasicEntity) {}
