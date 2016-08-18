package main

import "engo.io/ecs"

type HealthSystem struct {
	healths map[*ecs.BasicEntity]*HealthComponent
}

type HealthComponent struct {
	Life int
	Dead bool
}

func (hs *HealthSystem) Update(dt float32) {
	for _, h := range hs.healths {
		if h.Life <= 0 {
			h.Dead = true
		}
	}
}

func (hs *HealthSystem) Add(e *ecs.BasicEntity, health *HealthComponent) {
	hs.healths[e] = health
}

func (hs *HealthSystem) Remove(entity ecs.BasicEntity) {
	delete(hs.healths, &entity)
}
