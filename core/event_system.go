package core

import (
	"engo.io/ecs"
	log "github.com/Sirupsen/logrus"
)

type EventSystem struct {
	world *ecs.World

	eventHistory []Event
	activeEvents []Event

	incoming   chan NetworkMessage
	outgoing   chan NetworkMessage
	serverRoom *ServerRoom
}

// New is the initialisation of the System
func (es *EventSystem) New(w *ecs.World) {}

func (es *EventSystem) Update(dt float32) {
	// Process currently active events in order, in serial, stopping if one can't complete
	for i := 0; i < len(es.activeEvents); i++ {
		event := es.activeEvents[0]

		if event.Process(es.world, dt) {
			es.eventHistory = append(es.eventHistory, event)
			es.activeEvents = es.activeEvents[1:]
		} else {
			break
		}
	}

	select {
	case message, ok := <-es.incoming:
		if ok {
			for _, event := range message.Events {
				es.activeEvents = append(es.activeEvents, event)
				if es.serverRoom != nil {
					//log.Infof("Sending event to all clients: %v", reflect.TypeOf(event))
				}
			}
			if es.serverRoom != nil {
				es.serverRoom.SendToAllClients(message)
			}
		} else {
			log.Fatal("channel closed")
		}
	default:
	}
}

func (es *EventSystem) AddEvents(events ...Event) {
	es.activeEvents = append(es.activeEvents, events...)
}
func (es *EventSystem) Remove(entity ecs.BasicEntity) {}
