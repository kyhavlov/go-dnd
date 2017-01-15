package core

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	log "github.com/Sirupsen/logrus"
	"image/color"
)

type TurnSystem struct {
	PlayerActions map[PlayerID]*Event
	PlayerReady   map[PlayerID]bool
	PlayersTurn   bool

	event *EventSystem
	ui    *UiSystem
}

func (ts *TurnSystem) IsPlayerReady(id PlayerID) bool {
	return ts.PlayerReady[id]
}

func (ts *TurnSystem) New(w *ecs.World) {
	ts.PlayerActions = make(map[PlayerID]*Event)
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
			events := make([]Event, playerCount)
			for id, event := range ts.PlayerActions {
				events[id] = *event
			}
			ts.event.AddEvents(events...)
			ts.event.AddEvents(&TurnChangeEvent{true})
			for i := 0; i < playerCount; i++ {
				ts.PlayerActions[PlayerID(i)] = nil
				ts.PlayerReady[PlayerID(i)] = false
				ts.ui.UpdateActionIndicator(PlayerID(i), nil)
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
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *UiSystem:
			switch action := p.Action.(type) {
			case *MoveEvent:
				var lines []*UiElement
				for i := 0; i < len(action.Path)-1; i++ {
					line := UiElement{BasicEntity: ecs.NewBasic()}
					current := action.Path[i].toPixels()
					next := action.Path[i+1].toPixels()
					start := current
					if current.X > next.X || current.Y > next.Y {
						start = next
					}
					w := float32(3)
					h := float32(3)
					if current.X != next.X {
						w = TileWidth
					} else {
						h = TileWidth
					}
					line.SpaceComponent = common.SpaceComponent{Position: engo.Point{start.X + TileWidth/2, start.Y + TileWidth/2}, Width: w, Height: h}
					line.RenderComponent = common.RenderComponent{Drawable: common.Rectangle{}, Color: color.RGBA{0, 255, 0, 255}}
					lines = append(lines, &line)
				}
				sys.UpdateActionIndicator(p.PlayerID, lines)
			}
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
