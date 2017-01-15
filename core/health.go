package core

import (
	"engo.io/ecs"
	log "github.com/Sirupsen/logrus"
	"github.com/kyhavlov/go-dnd/structs"
)

type HealthSystem struct {
	healths map[ecs.BasicEntity]*structs.HealthComponent
}

func (hs *HealthSystem) New(w *ecs.World) {
	hs.healths = make(map[ecs.BasicEntity]*structs.HealthComponent)
}

func (hs *HealthSystem) Update(dt float32) {
	for _, h := range hs.healths {
		if h.Life <= 0 && !h.Dead {
			h.Dead = true
			log.Info("Unit died")
		}
	}
}

func (hs *HealthSystem) Add(e ecs.BasicEntity, health *structs.HealthComponent) {
	hs.healths[e] = health
}

func (hs *HealthSystem) Remove(entity ecs.BasicEntity) {
	delete(hs.healths, entity)
}
