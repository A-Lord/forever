package shared

import (
	"fmt"
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Alchemists' Stone 13503's equip spell: A_DUMMY 20, "Increases the effects that healing and mana
// potions have on the wearer by $s1%."
const alchemistsStoneSpellID int32 = 17619

// A potion: the spell its on-use row casts on the drinker, on the item's cooldown and the category
// it shares. Only the spec's combat potion is a major cooldown; the other potions are the rotation's.
func NewSpellDataPotion(itemID int32) {
	registerSpellDataConsumable(itemID, potionConsumable)
}

// A conjured item, a healthstone, mana gem, rune or tea, built the way a potion is. It is a major
// cooldown the manager uses only while it is the spec's chosen one.
func NewSpellDataConjured(itemID int32) {
	registerSpellDataConsumable(itemID, conjuredConsumable)
}

// Any other consumable whose on-use the potion builder runs, such as the Flame Deflector's shield: a
// major cooldown whenever the spec names it.
func NewSpellDataOnUseConsumable(itemID int32) {
	registerSpellDataConsumable(itemID, otherConsumable)
}

type spellDataConsumableKind uint8

const (
	potionConsumable spellDataConsumableKind = iota
	conjuredConsumable
	otherConsumable
)

// Why the consumable builders cannot run the spell, or nil where they can. Every effect has to land
// on the user and be a heal, an absorb, a gain of mana, rage or energy, damage to the user, or an
// aura the parser attaches.
func ConsumableUnsupported(row *spelldata.Spell) []string {
	_, unsupported := consumablePartsOf(row)
	return unsupported
}

type consumableParts struct {
	heal       *spelldata.Effect
	absorb     *spelldata.Effect
	gains      []*spelldata.Effect
	selfDamage *spelldata.Effect
	auraIdx    []int32
}

func consumablePartsOf(row *spelldata.Spell) (consumableParts, []string) {
	var parts consumableParts
	var unsupported []string
	if row == spelldata.Nil {
		return parts, []string{"the spell has no row in the store"}
	}

	for i := range row.Effects {
		e := &row.Effects[i]
		kind := e.Type.String()
		if e.Type == dbcenums.E_APPLY_AURA {
			kind = e.Aura.String()
		}

		switch {
		case e.Target[0] != dbcenums.TARGET_UNIT_CASTER:
			unsupported = append(unsupported, fmt.Sprintf("%s lands on implicit target %d, not the user", kind, e.Target[0]))
		case e.Type == dbcenums.E_HEAL, e.Type == dbcenums.E_HEAL_PCT,
			e.Type == dbcenums.E_APPLY_AURA && e.Aura == dbcenums.A_PERIODIC_HEAL:
			if parts.heal != nil {
				unsupported = append(unsupported, fmt.Sprintf("a second heal, %s", kind))
			}
			parts.heal = e
		case e.Type == dbcenums.E_APPLY_AURA && e.Aura == dbcenums.A_SCHOOL_ABSORB:
			if parts.absorb != nil {
				unsupported = append(unsupported, "a second absorb")
			}
			parts.absorb = e
		case e.Type == dbcenums.E_ENERGIZE, e.Type == dbcenums.E_ENERGIZE_PCT,
			e.Type == dbcenums.E_APPLY_AURA && e.Aura == dbcenums.A_PERIODIC_ENERGIZE:
			switch dbcenums.PowerType(e.Misc) {
			case dbcenums.POWER_MANA, dbcenums.POWER_RAGE, dbcenums.POWER_ENERGY:
				parts.gains = append(parts.gains, e)
			default:
				unsupported = append(unsupported, fmt.Sprintf("%s fills power type %d", kind, e.Misc))
			}
		case e.Type == dbcenums.E_SCHOOL_DAMAGE:
			if parts.selfDamage != nil {
				unsupported = append(unsupported, "a second hit on the user")
			}
			parts.selfDamage = e
		case e.ParsedAura():
			parts.auraIdx = append(parts.auraIdx, int32(i+1))
		default:
			unsupported = append(unsupported, fmt.Sprintf("%s, which no consumable builder models", kind))
		}

		ticks := e.Type == dbcenums.E_APPLY_AURA && (e.Aura == dbcenums.A_PERIODIC_HEAL || e.Aura == dbcenums.A_PERIODIC_ENERGIZE)
		if ticks && (e.PeriodMs <= 0 || row.DurationMs <= 0) {
			unsupported = append(unsupported, fmt.Sprintf("%s states no period or no duration to tick over", kind))
		}
	}

	if parts.heal != nil && parts.absorb != nil {
		unsupported = append(unsupported, "a heal beside an absorb")
	}
	if len(parts.auraIdx) > 0 && row.DurationMs <= 0 {
		unsupported = append(unsupported, fmt.Sprintf("%d states no duration for its aura", row.ID))
	}
	return parts, unsupported
}

func registerSpellDataConsumable(itemID int32, kind spellDataConsumableKind) {
	core.NewDataDrivenConsumableEffect(itemID, func(agent core.Agent) {
		character := agent.GetCharacter()
		consumable := core.GetConsumableByID(itemID)
		itemEffect := consumableOnUse(consumable)
		row := spelldata.MustFind(itemEffect.BuffId)
		parts, unsupported := consumablePartsOf(row)
		if len(unsupported) > 0 {
			panic(fmt.Sprintf("consumable %d casts %d, which the builder cannot run: %v", itemID, row.ID, unsupported))
		}

		actionID := core.ActionID{ItemID: itemID}
		potionBonus := 1.0
		if kind == potionConsumable && character.HasAlchStone() {
			potionBonus += spelldata.MustFind(alchemistsStoneSpellID).EffectN(1).Percent()
		}

		var config core.SpellConfig
		switch {
		case parts.heal != nil:
			config = spellDataProcHealSpell(character, row)
			config.DamageMultiplier *= potionBonus
		case parts.absorb != nil:
			config = spellDataAbsorbSpell(character, row, itemID)
		default:
			config = spelldata.SpellConfig(&character.Unit, row)
		}
		config.ActionID = actionID
		config.Flags &^= core.SpellFlagPassiveSpell | core.SpellFlagProc
		config.Flags |= core.SpellFlagNoOnCastComplete | core.SpellFlagAPL
		switch kind {
		case potionConsumable:
			// Encounter-only, so a prepull potion is the generic Potion action and the rotation cannot
			// name one the consumes do not.
			config.Flags |= core.SpellFlagEncounterOnly | core.SpellFlagPotion
		case conjuredConsumable:
			config.Flags |= core.SpellFlagConjured
		}

		config.Cast = spelldata.Cast(row)
		config.Cast.CD, config.Cast.SharedCD = consumableCooldowns(character, itemEffect, row)
		sharedCD := config.Cast.SharedCD.Duration

		cdType := core.CooldownTypeUnknown
		var applyAura func(*core.Simulation)
		if len(parts.auraIdx) > 0 {
			aura := character.GetOrRegisterAura(spelldata.AuraConfig(row))
			spelldata.ParseEffects(character, aura, row, spelldata.Effects(parts.auraIdx...))
			config.RelatedSelfBuff = aura
			applyAura = func(sim *core.Simulation) { aura.Activate(sim) }
			cdType = (&core.StatBuffAura{BuffedStatTypes: consumable.Stats.GetBuffedStatTypes()}).InferCDType()
		}

		var gains []*consumableGain
		for i := range row.Effects {
			e := &row.Effects[i]
			if e == parts.heal || e == parts.absorb {
				cdType = core.CooldownTypeSurvival
				continue
			}
			if !slices.Contains(parts.gains, e) {
				continue
			}
			gain := newConsumableGain(character, row, e, actionID, potionBonus)
			if gain == nil {
				continue
			}
			gains = append(gains, gain)
			switch {
			case gain.power != dbcenums.POWER_MANA:
				cdType = core.CooldownTypeDPS
			case cdType != core.CooldownTypeSurvival:
				cdType = core.CooldownTypeMana
			}
		}

		if parts.heal == nil && parts.absorb == nil && len(parts.auraIdx) == 0 && len(gains) == 0 {
			return
		}

		var selfDamage *core.Spell
		if parts.selfDamage != nil {
			selfDamage = character.RegisterSpell(consumableSelfDamageSpell(character, row, parts.selfDamage, actionID))
		}

		ownEffects := config.ApplyEffects
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			if ownEffects != nil {
				ownEffects(sim, &character.Unit, spell)
			}
			for i := range gains {
				gains[i].apply(sim)
			}
			if selfDamage != nil {
				selfDamage.Cast(sim, &character.Unit)
			}
			if applyAura != nil {
				applyAura(sim)
			}
			if sim.CurrentTime < 0 && sharedCD > 0 {
				spell.SharedCD.Set(sim.CurrentTime + sharedCD)
				character.UpdateMajorCooldowns()
			}
		}

		spell := character.RegisterSpell(config)

		consumables := character.Consumables
		mcd := core.MajorCooldown{Spell: spell, Type: cdType}
		if parts.heal == nil && parts.absorb == nil && len(parts.auraIdx) == 0 {
			mcd.ShouldActivate = func(sim *core.Simulation, character *core.Character) bool {
				for i := range gains {
					if !gains[i].fits() {
						return false
					}
				}
				return true
			}
		}

		switch kind {
		case potionConsumable:
			if consumables.PotId != itemID {
				return
			}
			spell.Flags |= core.SpellFlagCombatPotion
		case conjuredConsumable:
			shouldActivate := mcd.ShouldActivate
			mcd.ShouldActivate = func(sim *core.Simulation, character *core.Character) bool {
				return consumables.ConjuredId == itemID && (shouldActivate == nil || shouldActivate(sim, character))
			}
		}
		character.AddMajorCooldown(mcd)
	})
}

func consumableOnUse(consumable core.Consumable) *proto.ItemEffect {
	for _, effect := range consumable.ItemEffects {
		if effect.GetOnUse() != nil {
			return effect
		}
	}
	panic(fmt.Sprintf("consumable %d states no on-use row", consumable.Id))
}

// The item's own cooldown and the category cooldown it shares. -1 on the item's row, and a category
// of 0, defer to the spell's row: the Spellblasting potions 250932-250937 state neither and share
// potion category 4 for 120 s through their spells.
func consumableCooldowns(character *core.Character, itemEffect *proto.ItemEffect, row *spelldata.Spell) (core.Cooldown, core.Cooldown) {
	onUse := itemEffect.GetOnUse()
	cooldownMs, categoryMs, category := onUse.CooldownMs, onUse.CategoryCooldownMs, onUse.CategoryId
	if cooldownMs < 0 {
		cooldownMs = row.CooldownMs
	}
	if category <= 0 {
		category = int32(row.Category)
	}
	if categoryMs < 0 {
		categoryMs = row.CategoryCooldownMs
	}

	var cd core.Cooldown
	if cooldownMs > 0 {
		cd = core.Cooldown{Timer: character.NewTimer(), Duration: core.DurationFromMillis(cooldownMs)}
	}
	if category <= 0 || categoryMs <= 0 {
		return cd, core.Cooldown{}
	}
	return cd, core.Cooldown{
		Timer:    character.GetOrInitSpellCategoryTimer(category),
		Duration: core.DurationFromMillis(categoryMs),
	}
}

// One gain of mana, rage or energy, at once or every period of the row's duration. A character
// without the bar gets none.
type consumableGain struct {
	power dbcenums.PowerType
	apply func(*core.Simulation)
	// Whether the whole gain fits under the bar's maximum, allowing for five seconds of mana regen.
	fits func() bool
}

func newConsumableGain(character *core.Character, row *spelldata.Spell, e *spelldata.Effect, actionID core.ActionID, potionBonus float64) *consumableGain {
	power := dbcenums.PowerType(e.Misc)
	var maximum, current func() float64
	var add func(*core.Simulation, float64, *core.ResourceMetrics)
	var metrics *core.ResourceMetrics
	bonus := 1.0
	switch {
	case power == dbcenums.POWER_MANA && character.HasManaBar():
		maximum, current, add, metrics = character.MaxMana, character.CurrentMana, character.AddMana, character.NewManaMetrics(actionID)
		bonus = potionBonus
	case power == dbcenums.POWER_RAGE && character.HasRageBar():
		maximum, current, add, metrics = character.MaximumRage, character.CurrentRage, character.AddRage, character.NewRageMetrics(actionID)
	case power == dbcenums.POWER_ENERGY && character.HasEnergyBar():
		maximum, current, add, metrics = character.MaximumEnergy, character.CurrentEnergy, character.AddEnergy, character.NewEnergyMetrics(actionID)
	default:
		return nil
	}

	level := character.Level
	amount := func(roll func(*spelldata.Effect) float64) float64 {
		if e.Type == dbcenums.E_ENERGIZE_PCT {
			return maximum() * e.Percent() * bonus
		}
		value := roll(e)
		if power.InTenths() {
			value /= 10
		}
		return value * bonus
	}

	ticks := 1
	gain := &consumableGain{power: power}
	if e.Aura == dbcenums.A_PERIODIC_ENERGIZE {
		ticks = int(row.DurationMs / e.PeriodMs)
		gain.apply = func(sim *core.Simulation) {
			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period:   e.Period(),
				NumTicks: ticks,
				Priority: core.ActionPriorityDOT,
				OnAction: func(sim *core.Simulation) {
					add(sim, amount(func(e *spelldata.Effect) float64 { return e.Roll(sim, level) }), metrics)
				},
			})
		}
	} else {
		gain.apply = func(sim *core.Simulation) {
			add(sim, amount(func(e *spelldata.Effect) float64 { return e.Roll(sim, level) }), metrics)
		}
	}

	gain.fits = func() bool {
		room := maximum() - current()
		if power == dbcenums.POWER_MANA {
			room -= character.ManaRegenPerSecondWhileCasting() * 5
		}
		return room >= amount(func(e *spelldata.Effect) float64 { return e.Max(level) })*float64(ticks)
	}
	return gain
}

// The damage the spell deals its user, as its own spell so the hit is one the character takes.
func consumableSelfDamageSpell(character *core.Character, row *spelldata.Spell, e *spelldata.Effect, actionID core.ActionID) core.SpellConfig {
	config := spelldata.SpellConfig(&character.Unit, row, spelldata.Magic(core.ProcMaskEmpty), spelldata.Proc())
	config.ActionID = actionID.WithTag(1)
	config.Flags &^= core.SpellFlagHelpful
	config.DefenseType = damageDefenseType(config.DefenseType, config.SpellSchool, false)
	outcome := damageOutcome(config.DefenseType, row.CannotCrit(), OutcomeDefault)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, e.Roll(sim, character.Level), GetOutcome(spell, outcome))
	}
	return config
}
