package core

import (
	"fmt"
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	log "github.com/Sirupsen/logrus"
	"github.com/kyhavlov/go-dnd/structs"
)

const EquipmentHotkeys = "GHJKL"
const InventoryHotkeys = "ZXCVB"
const SkillHotkeys = "1234567890"

type UiElement struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

type DynamicText struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent

	UpdateFunc func() string
	lastValue  string
}

type UiSystem struct {
	dynamicTexts     map[*ecs.BasicEntity]*DynamicText
	actionIndicators map[PlayerID][]*UiElement

	equipmentFrames  [structs.EquipmentSlots]*common.SpaceComponent
	equipmentDisplay [structs.EquipmentSlots]*ecs.BasicEntity

	inventoryFrames  [structs.InventorySize]*common.SpaceComponent
	inventoryDisplay [structs.InventorySize]*ecs.BasicEntity

	skillIcons   map[string]common.Drawable
	skillFrames  [structs.SkillSlots]*common.SpaceComponent
	skillDisplay [structs.SkillSlots]*ecs.BasicEntity

	input  *InputSystem
	render *common.RenderSystem
}

func (us *UiSystem) Update(dt float32) {
	// Check for updates of dynamic text objects
	for _, text := range us.dynamicTexts {
		if text.UpdateFunc != nil {
			newValue := text.UpdateFunc()
			if newValue != text.lastValue {
				text.Drawable = common.Text{
					Font: text.Drawable.(common.Text).Font,
					Text: newValue,
				}
			}
			text.lastValue = newValue
		}
	}
}

func (us *UiSystem) Add(e *ecs.BasicEntity, text *DynamicText) {
	us.dynamicTexts[e] = text
	us.render.Add(e, &text.RenderComponent, &text.SpaceComponent)
}

func (us *UiSystem) Remove(entity ecs.BasicEntity) {
	delete(us.dynamicTexts, &entity)
}

// New is the initialisation of the System
func (us *UiSystem) New(w *ecs.World) {
	us.dynamicTexts = make(map[*ecs.BasicEntity]*DynamicText)
	us.actionIndicators = make(map[PlayerID][]*UiElement)

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			us.render = sys
		}
	}

	// Load skill icons
	sheet := common.NewSpritesheetFromFile(structs.SpritesheetPath, structs.TileWidth, structs.TileWidth)
	us.skillIcons = make(map[string]common.Drawable)
	for _, skill := range structs.GetAllSkills() {
		us.skillIcons[skill.Name] = sheet.Cell(skill.Icon)
	}
}

func (us *UiSystem) AddActionIndicators(player PlayerID, elems []*UiElement) {
	for _, elem := range elems {
		us.render.Add(&elem.BasicEntity, &elem.RenderComponent, &elem.SpaceComponent)
	}
	us.actionIndicators[player] = append(us.actionIndicators[player], elems...)
}

func (us *UiSystem) ResetActionIndicators(player PlayerID) {
	prev, ok := us.actionIndicators[player]
	if ok {
		for _, elem := range prev {
			us.render.Remove(elem.BasicEntity)
		}
	}
	us.actionIndicators[player] = nil
}

// Update the items shown in the inventory display slots
func (us *UiSystem) UpdatePlayerDisplay() {
	equipment := us.input.player.Equipment
	for i := 0; i < structs.EquipmentSlots; i++ {
		if us.equipmentDisplay[i] != nil {
			us.render.Remove(*us.equipmentDisplay[i])
		}
		if equipment[i] == nil {
			us.equipmentDisplay[i] = nil
			continue
		}
		entity := ecs.NewBasic()
		us.equipmentDisplay[i] = &entity
		component := common.RenderComponent{Drawable: equipment[i].Drawable}
		component.SetShader(common.HUDShader)
		component.SetZIndex(3)
		log.Infof("Adding equipment item display")
		us.render.Add(&entity, &component, us.equipmentFrames[i])
	}

	inventory := us.input.player.Inventory
	for i := 0; i < structs.InventorySize; i++ {
		if us.inventoryDisplay[i] != nil {
			us.render.Remove(*us.inventoryDisplay[i])
		}
		if inventory[i] == nil {
			us.inventoryDisplay[i] = nil
			continue
		}
		entity := ecs.NewBasic()
		us.inventoryDisplay[i] = &entity
		component := common.RenderComponent{Drawable: inventory[i].Drawable}
		component.SetShader(common.HUDShader)
		component.SetZIndex(3)
		log.Infof("Adding inventory item display")
		us.render.Add(&entity, &component, us.inventoryFrames[i])
	}

	skills := us.input.player.GetSkills()
	log.Infof("player skill count: %d", len(skills))
	for i := 0; i < structs.SkillSlots; i++ {
		if us.skillDisplay[i] != nil {
			us.render.Remove(*us.skillDisplay[i])
		}
		if len(skills) <= i {
			continue
		}
		entity := ecs.NewBasic()
		us.skillDisplay[i] = &entity
		skill := skills[i]
		component := common.RenderComponent{Drawable: us.skillIcons[skill]}
		component.SetShader(common.HUDShader)
		component.SetZIndex(3)
		log.Infof("Adding skill display")
		us.render.Add(&entity, &component, us.skillFrames[i])
	}
}

func (us *UiSystem) InitUI(w *ecs.World, playerCount int) {
	// Add UI displays
	font := &common.Font{
		URL:  "fonts/Gamegirl.ttf",
		FG:   color.White,
		Size: 12,
	}
	if err := font.CreatePreloaded(); err != nil {
		panic(err)
	}

	us.setupInventoryDisplay(font)
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *TurnSystem:
			log.Infof("player count for ui: %d", playerCount)
			us.setupReadyIndicators(sys, font, playerCount)
		}
	}
}

func (us *UiSystem) setupInventoryDisplay(font *common.Font) {
	for i := 0; i < structs.EquipmentSlots; i++ {
		itemFrame := UiElement{BasicEntity: ecs.NewBasic()}
		itemFrame.SpaceComponent = common.SpaceComponent{Position: engo.Point{float32(24+64*i) + 4, 644 + 4}, Width: structs.TileWidth, Height: structs.TileWidth}
		itemFrame.RenderComponent = common.RenderComponent{Drawable: common.Rectangle{BorderWidth: 2, BorderColor: color.White}, Color: color.RGBA{200, 153, 0, 125}}
		itemFrame.SetShader(common.HUDShader)
		itemFrame.RenderComponent.SetZIndex(2)
		us.equipmentFrames[i] = &itemFrame.SpaceComponent
		us.render.Add(&itemFrame.BasicEntity, &itemFrame.RenderComponent, &itemFrame.SpaceComponent)

		hotkey := UiElement{BasicEntity: ecs.NewBasic()}
		hotkey.SpaceComponent = common.SpaceComponent{Position: engo.Point{float32(24+64*i) + 4, 644 + 7}, Width: structs.TileWidth, Height: structs.TileWidth}
		hotkey.RenderComponent = common.RenderComponent{Drawable: common.Text{Font: font, Text: string(EquipmentHotkeys[i])}, Color: color.White}
		hotkey.SetShader(common.HUDShader)
		hotkey.RenderComponent.SetZIndex(4)
		us.render.Add(&hotkey.BasicEntity, &hotkey.RenderComponent, &hotkey.SpaceComponent)
	}

	for i := 0; i < structs.InventorySize; i++ {
		itemFrame := UiElement{BasicEntity: ecs.NewBasic()}
		itemFrame.SpaceComponent = common.SpaceComponent{Position: engo.Point{float32(24+64*i) + 4, 712 + 4}, Width: structs.TileWidth, Height: structs.TileWidth}
		itemFrame.RenderComponent = common.RenderComponent{Drawable: common.Rectangle{BorderWidth: 2, BorderColor: color.White}, Color: color.RGBA{200, 153, 0, 125}}
		itemFrame.SetShader(common.HUDShader)
		itemFrame.RenderComponent.SetZIndex(2)
		us.inventoryFrames[i] = &itemFrame.SpaceComponent
		us.render.Add(&itemFrame.BasicEntity, &itemFrame.RenderComponent, &itemFrame.SpaceComponent)

		hotkey := UiElement{BasicEntity: ecs.NewBasic()}
		hotkey.SpaceComponent = common.SpaceComponent{Position: engo.Point{float32(24+64*i) + 4, 712 + 7}, Width: structs.TileWidth, Height: structs.TileWidth}
		hotkey.RenderComponent = common.RenderComponent{Drawable: common.Text{Font: font, Text: string(InventoryHotkeys[i])}, Color: color.White}
		hotkey.SetShader(common.HUDShader)
		hotkey.RenderComponent.SetZIndex(4)
		us.render.Add(&hotkey.BasicEntity, &hotkey.RenderComponent, &hotkey.SpaceComponent)
	}

	for i := 0; i < structs.SkillSlots; i++ {
		itemFrame := UiElement{BasicEntity: ecs.NewBasic()}
		itemFrame.SpaceComponent = common.SpaceComponent{Position: engo.Point{float32(534+64*i) + 4, 712 + 4}, Width: structs.TileWidth, Height: structs.TileWidth}
		itemFrame.RenderComponent = common.RenderComponent{Drawable: common.Rectangle{BorderWidth: 2, BorderColor: color.White}, Color: color.RGBA{200, 153, 0, 125}}
		itemFrame.SetShader(common.HUDShader)
		itemFrame.RenderComponent.SetZIndex(2)
		us.skillFrames[i] = &itemFrame.SpaceComponent
		us.render.Add(&itemFrame.BasicEntity, &itemFrame.RenderComponent, &itemFrame.SpaceComponent)

		hotkey := UiElement{BasicEntity: ecs.NewBasic()}
		hotkey.SpaceComponent = common.SpaceComponent{Position: engo.Point{float32(534+64*i) + 4, 712 + 7}, Width: structs.TileWidth, Height: structs.TileWidth}
		hotkey.RenderComponent = common.RenderComponent{Drawable: common.Text{Font: font, Text: string(SkillHotkeys[i])}, Color: color.White}
		hotkey.SetShader(common.HUDShader)
		hotkey.RenderComponent.SetZIndex(4)
		us.render.Add(&hotkey.BasicEntity, &hotkey.RenderComponent, &hotkey.SpaceComponent)
	}
}

func (us *UiSystem) setupReadyIndicators(sys *TurnSystem, font *common.Font, playerCount int) {
	for i := 0; i < playerCount; i++ {
		readyStatus := DynamicText{BasicEntity: ecs.NewBasic()}
		readyStatus.RenderComponent.Drawable = common.Text{
			Font: font,
		}
		readyStatus.SetShader(common.HUDShader)
		readyStatus.SpaceComponent.Position.Set(24, float32(120+(i*72)))
		readyStatus.RenderComponent.SetZIndex(2)
		playerNum := i + 1
		readyStatus.UpdateFunc = func() string {
			ready := sys.IsPlayerReady(PlayerID(playerNum - 1))
			status := "Not Ready"
			readyStatus.RenderComponent.Color = color.White
			if ready {
				status = "Ready"
				readyStatus.RenderComponent.Color = color.RGBA{0, 255, 0, 120}
			}
			return fmt.Sprintf("Player %d: %v", playerNum, status)
		}

		us.Add(&readyStatus.BasicEntity, &readyStatus)

		for j := 0; j < 2; j++ {
			actionStatus := DynamicText{BasicEntity: ecs.NewBasic()}
			actionStatus.RenderComponent.Drawable = common.Text{
				Font: font,
			}
			actionStatus.SetShader(common.HUDShader)
			actionStatus.SpaceComponent.Position.Set(24, float32(138+(i*72)+(j*18)))
			actionStatus.RenderComponent.SetZIndex(2)
			actionNum := j
			actionStatus.UpdateFunc = func() string {
				actionStatus.RenderComponent.Color = color.White
				if len(sys.PlayerActions[PlayerID(playerNum-1)]) > actionNum {
					action := sys.PlayerActions[PlayerID(playerNum-1)][actionNum]
					return "  - " + action.(NamedEvent).Name()
				}
				return ""
			}

			us.Add(&actionStatus.BasicEntity, &actionStatus)
		}
	}
}

func (us *UiSystem) SetupStatsDisplay(world *ecs.World) {
	position := engo.Point{24, 24}
	width := float32(320)
	height := float32(80)
	bgColor := color.RGBA{200, 153, 0, 125}

	// Create the panel background
	bg := UiElement{
		BasicEntity:     ecs.NewBasic(),
		RenderComponent: common.RenderComponent{Drawable: common.Rectangle{BorderWidth: 1, BorderColor: color.White}, Color: bgColor},
		SpaceComponent:  common.SpaceComponent{Position: position, Width: width, Height: height},
	}

	// Set the background camera-independent
	bg.RenderComponent.SetZIndex(1) // zIndex > 0 (default)
	bg.RenderComponent.SetShader(common.HUDShader)

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&bg.BasicEntity, &bg.RenderComponent, &bg.SpaceComponent)
		}
	}

	fnt := &common.Font{
		URL:  "fonts/Gamegirl.ttf",
		FG:   color.White,
		Size: 12,
	}
	err := fnt.CreatePreloaded()
	if err != nil {
		panic(err)
	}

	// Add text fields
	lifeDisplay := DynamicText{BasicEntity: ecs.NewBasic()}
	lifeDisplay.RenderComponent.Drawable = common.Text{
		Font: fnt,
	}
	lifeDisplay.SetShader(common.HUDShader)
	lifeDisplay.SpaceComponent.Position.Set(position.X+10, position.Y+12)
	lifeDisplay.RenderComponent.SetZIndex(2)

	lifeDisplay.UpdateFunc = func() string {
		return fmt.Sprintf("Life:    %d/%d", us.input.player.Life, us.input.player.GetEffectiveMaxLife())
	}
	us.Add(&lifeDisplay.BasicEntity, &lifeDisplay)

	staminaDisplay := DynamicText{BasicEntity: ecs.NewBasic()}
	staminaDisplay.RenderComponent.Drawable = common.Text{
		Font: fnt,
	}
	staminaDisplay.SetShader(common.HUDShader)
	staminaDisplay.SpaceComponent.Position.Set(position.X+10, position.Y+36)
	staminaDisplay.RenderComponent.SetZIndex(2)
	staminaDisplay.UpdateFunc = func() string {
		return fmt.Sprintf("Stamina: %d/%d", us.input.player.Stamina, us.input.player.MaxStamina)
	}
	us.Add(&staminaDisplay.BasicEntity, &staminaDisplay)

	statDisplay := DynamicText{BasicEntity: ecs.NewBasic()}
	statDisplay.RenderComponent.Drawable = common.Text{
		Font: fnt,
	}
	statDisplay.SetShader(common.HUDShader)
	statDisplay.SpaceComponent.Position.Set(position.X+10, position.Y+60)
	statDisplay.RenderComponent.SetZIndex(2)
	statDisplay.UpdateFunc = func() string {
		return fmt.Sprintf("Str %d  Dex %d  Int %d", us.input.player.GetEffectiveStrength(),
			us.input.player.GetEffectiveDexterity(), us.input.player.GetEffectiveIntelligence())
	}
	us.Add(&statDisplay.BasicEntity, &statDisplay)
}
