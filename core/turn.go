package core

import (
	"engo.io/ecs"
	log "github.com/Sirupsen/logrus"
)

type TurnSystem struct {
	PlayerActions map[PlayerID][]Event
	PlayerReady   map[PlayerID]bool
	PlayersTurn   bool

	event *EventSystem
	ui    *UiSystem
}

func (ts *TurnSystem) IsPlayerReady(id PlayerID) bool {
	return ts.PlayerReady[id]
}

func (ts *TurnSystem) PlayerHasMove(id PlayerID) bool {
	for _, action := range ts.PlayerActions[id] {
		switch action.(type) {
		case *Move:
			return false
		}
	}
	return true
}

func (ts *TurnSystem) PlayerMovingFirst(id PlayerID) bool {
	if len(ts.PlayerActions[id]) == 0 {
		return false
	}

	switch ts.PlayerActions[id][0].(type) {
	case *Move:
		return true
	default:
		return false
	}
}

func (ts *TurnSystem) New(w *ecs.World) {
	ts.PlayerActions = make(map[PlayerID][]Event)
	ts.PlayerReady = make(map[PlayerID]bool)
	ts.PlayerReady[PlayerID(0)] = false
	ts.PlayersTurn = true

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *EventSystem:
			ts.event = sys
		case *UiSystem:
			ts.ui = sys
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
		for _, action := range ts.PlayerActions {
			if action == nil {
				allReady = false
			}
		}

		if allReady {
			playerCount := len(ts.PlayerActions)
			log.Infof("All %d players ready", playerCount)
			events := make([][]Event, playerCount)
			for id, event := range ts.PlayerActions {
				events[id] = event
			}
			for _, playerEvents := range events {
				ts.event.AddEvents(playerEvents...)
			}
			ts.event.AddEvents(&TurnChange{false})
			ts.event.AddEvents(&EnemyTurn{})
			for i := 0; i < playerCount; i++ {
				ts.PlayerActions[PlayerID(i)] = nil
				ts.PlayerReady[PlayerID(i)] = false
				ts.ui.ResetActionIndicators(PlayerID(i))
			}
			ts.PlayersTurn = false
		}
	}
}

func (ts *TurnSystem) Remove(entity ecs.BasicEntity) {}
