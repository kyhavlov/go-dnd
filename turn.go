package main

import (
	"engo.io/ecs"
	log "github.com/Sirupsen/logrus"
)

type TurnSystem struct {
	PlayerActions map[PlayerID]*Event
	PlayerReady   map[PlayerID]bool
	PlayersTurn   bool

	eventSystem  *EventSystem
}

func (ts *TurnSystem) New(w *ecs.World) {
	ts.PlayerActions = make(map[PlayerID]*Event)
	ts.PlayerReady = make(map[PlayerID]bool)
	ts.PlayerReady[PlayerID(0)] = false
	ts.PlayersTurn = true

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *EventSystem:
			ts.eventSystem = sys
		}
	}
}

func (ts *TurnSystem) Update(dt float32) {
	if ts.PlayersTurn {
		allReady := true

		for _, ready := range ts.PlayerReady {
			if !ready {
				allReady = false
			}
		}

		if allReady {
			playerCount := len(ts.PlayerActions)
			log.Infof("All %d players ready", playerCount)
			events := make([]Event, playerCount)
			for id, event := range ts.PlayerActions {
				events[id] = *event
			}
			ts.eventSystem.AddEvents(events...)
			ts.eventSystem.AddEvents(&TurnChangeEvent{true})
			for i := 0; i < playerCount; i++ {
				ts.PlayerActions[PlayerID(i)] = nil
				ts.PlayerReady[PlayerID(i)] = false
			}
			ts.PlayersTurn = false
		}
	}
}

func (ts *TurnSystem) Remove(entity ecs.BasicEntity) {}

type PlayerAction struct {
	PlayerID PlayerID
	Action   Event
}

func (p *PlayerAction) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *TurnSystem:
			//log.Debugf("Setting action %v for player %d", reflect.TypeOf(p.Action), p.PlayerID)
			sys.PlayerActions[p.PlayerID] = &p.Action
		}
	}
	return true
}

type PlayerReady struct {
	PlayerID
	Ready bool
}

func (p *PlayerReady) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *TurnSystem:
			//log.Debugf("Setting ready to %v for player %d", p.Ready, p.PlayerID)
			sys.PlayerReady[p.PlayerID] = p.Ready
		}
	}
	return true
}

type TurnChangeEvent struct {
	PlayersTurn bool
}
func (t *TurnChangeEvent) Process(w *ecs.World, dt float32) bool {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *TurnSystem:
			log.Infof("Setting turn back to players")
			sys.PlayersTurn = t.PlayersTurn
		}
	}
	return true
}
