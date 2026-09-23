package core

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

var PotionAuraTag = "Potion"

func registerPotionCD(agent Agent, consumes *proto.ConsumesSpec) {
	character := agent.GetCharacter()
	defaultPotion := consumes.PotId

	for _, potionId := range consumes.Potions {
		potion := GetConsumableByID(potionId)
		if potion.Type == proto.ConsumableType_ConsumableTypePotion {
			potMCD := makePotionActivationSpell(potion.Id, character)
			if defaultPotion == potion.Id {
				potMCD.Spell.Flags |= SpellFlagCombatPotion
				character.AddMajorCooldown(potMCD)
			}
		}
	}
}

// Empty: the eight ids this carried are MoP-era alchemist stones and none of them is in
// this client's item database, so HasAlchStone was always false. The lookup is kept for
// whatever Forever's equivalent turns out to be.
var AlchStoneItemIDs = []int32{}

func (character *Character) HasAlchStone() bool {
	alchStoneEquipped := false
	for _, itemID := range AlchStoneItemIDs {
		alchStoneEquipped = alchStoneEquipped || character.HasTrinketEquipped(itemID)
	}
	return character.HasProfession(proto.Profession_Alchemy) && alchStoneEquipped
}

func makePotionActivationSpell(potionId int32, character *Character) MajorCooldown {
	potion := GetConsumableByID(potionId)
	categoryCooldownDuration := TernaryDuration(potion.CategoryCooldownDuration > 0, potion.CategoryCooldownDuration, time.Minute*2)
	mcd := makePotionActivationSpellInternal(potion, character)

	if mcd.Spell != nil {
		// Mark as 'Encounter Only' so that users are forced to select the generic Potion
		// placeholder action instead of specific potion spells, in APL prepull. This
		// prevents a mismatch between Consumes and Rotation settings.
		mcd.Spell.Flags |= SpellFlagEncounterOnly | SpellFlagPotion | SpellFlagAPL
		oldApplyEffects := mcd.Spell.ApplyEffects
		mcd.Spell.ApplyEffects = func(sim *Simulation, target *Unit, spell *Spell) {
			oldApplyEffects(sim, target, spell)
			if sim.CurrentTime < 0 {
				spell.SharedCD.Set(sim.CurrentTime + categoryCooldownDuration)
				character.UpdateMajorCooldowns()
			}
		}
	}

	return mcd
}

type resourceGainConfig struct {
	resType  proto.ResourceType
	min      float64
	spread   float64
	period   time.Duration // Duration between ticks; 0 for one-shot gains.
	duration time.Duration // Total duration of periodic gains.
}

func makePotionActivationSpellInternal(potion Consumable, character *Character) MajorCooldown {
	stoneMul := TernaryFloat64(character.HasAlchStone(), 1.4, 1.0)
	cooldownDuration := TernaryDuration(potion.CooldownDuration > 0, potion.CooldownDuration, time.Minute*2)

	potionCast := CastConfig{
		CD: Cooldown{
			Timer:    character.NewTimer(),
			Duration: cooldownDuration,
		},
		SharedCD: Cooldown{
			Timer:    character.GetPotionCD(),
			Duration: cooldownDuration,
		},
	}

	actionID := ActionID{ItemID: potion.Id}
	var aura *StatBuffAura
	mcd := MajorCooldown{
		Spell: character.GetOrRegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast:     potionCast,
		}),
	}
	if potion.BuffDuration > 0 {
		// Add stat buff aura if applicable
		aura = character.NewTemporaryStatsAura(potion.Name, actionID, potion.Stats, potion.BuffDuration)
		mcd.Spell.RelatedSelfBuff = aura.Aura
		mcd.Type = aura.InferCDType()
		mcd.BuffAura = aura
	}
	var gains []resourceGainConfig
	resourceMetrics := make(map[proto.ResourceType]*ResourceMetrics)

	// Stats applied by triggered auras (e.g. Fel Mana Potion's spell damage reduction).
	// These may be positive or negative.
	var auraStats stats.Stats
	var auraDuration time.Duration
	var auraSpellId int32

	for _, effectID := range potion.EffectIds {
		e := GetSpellEffectByID(effectID)
		resourceType := e.GetResourceType()
		isPeriodic := e.AuraPeriodMs > 0
		if resourceType != 0 && (isPeriodic || e.Type == proto.EffectType_EffectTypeResourceGain) {
			if resourceType == proto.ResourceType_ResourceTypeMana && mcd.Type != CooldownTypeSurvival {
				mcd.Type = CooldownTypeMana
			} else if resourceType == proto.ResourceType_ResourceTypeHealth {
				mcd.Type = CooldownTypeSurvival
			} else {
				mcd.Type = CooldownTypeDPS
			}
			gains = append(gains, resourceGainConfig{
				resType:  resourceType,
				min:      e.MinEffectSize,
				spread:   e.EffectSpread,
				period:   time.Duration(e.AuraPeriodMs) * time.Millisecond,
				duration: time.Duration(e.DurationMs) * time.Millisecond,
			})
			if _, exists := resourceMetrics[resourceType]; !exists {
				resourceMetrics[resourceType] = character.Metrics.NewResourceMetrics(actionID, resourceType)
			}
			// Preload resource types that are found on this item
			if resourceMetrics[resourceType] == nil {
				resourceMetrics[resourceType] = character.Metrics.NewResourceMetrics(actionID, resourceType)
			}
			continue
		}
		if effectStats := stats.FromProtoArray(e.Stats); effectStats != (stats.Stats{}) {
			auraStats = auraStats.Add(effectStats)
			auraDuration = max(auraDuration, time.Duration(e.DurationMs)*time.Millisecond)
			if auraSpellId == 0 {
				auraSpellId = e.SpellId
			}
		}
	}

	var debuffAura *StatBuffAura
	if auraDuration > 0 && auraStats != (stats.Stats{}) {
		debuffAura = character.NewTemporaryStatsAura(fmt.Sprintf("%s Debuff (%d)", potion.Name, auraSpellId), ActionID{SpellID: auraSpellId}, auraStats, auraDuration)
	}

	mcd.Spell.ApplyEffects = func(sim *Simulation, _ *Unit, _ *Spell) {
		if aura != nil {
			aura.Activate(sim)
		}
		if debuffAura != nil {
			debuffAura.Activate(sim)
		}
		for _, config := range gains {
			gain := config.min + sim.RandomFloat(potion.Name)*config.spread
			gain *= stoneMul
			if config.period > 0 {
				// Periodic gains (e.g. Fel Mana Potion) tick over the effect's duration.
				startPeriodicResourceGain(sim, character, config, gain, resourceMetrics[config.resType])
			} else {
				if config.resType == proto.ResourceType_ResourceTypeHealth {
					gain *= character.PseudoStats.HealingTakenMultiplier
				}
				character.ExecuteResourceGain(sim, config.resType, gain, resourceMetrics[config.resType])
			}
		}
	}

	mcd.ShouldActivate = func(sim *Simulation, character *Character) bool {
		shouldActivate := true
		for _, config := range gains {
			switch config.resType {
			case proto.ResourceType_ResourceTypeMana:
				totalRegen := character.ManaRegenPerSecondWhileCasting() * 5
				manaGain := config.min + config.spread
				manaGain *= stoneMul
				if config.period > 0 {
					manaGain *= float64(startPeriodicResourceGainTicks(config))
				}
				shouldActivate = character.MaxMana()-(character.CurrentMana()+totalRegen) >= manaGain
			}
		}
		return shouldActivate
	}

	return mcd

}

// startPeriodicResourceGain ticks a periodic resource gain (e.g. Fel Mana Potion's mana over
// time) across the effect's duration.
func startPeriodicResourceGain(sim *Simulation, character *Character, config resourceGainConfig, gainPerTick float64, metrics *ResourceMetrics) {
	if config.resType != proto.ResourceType_ResourceTypeMana && config.resType != proto.ResourceType_ResourceTypeHealth {
		return
	}
	if config.resType == proto.ResourceType_ResourceTypeHealth {
		gainPerTick *= character.PseudoStats.HealingTakenMultiplier
	}
	StartPeriodicAction(sim, PeriodicActionOptions{
		Period:   config.period,
		NumTicks: int(startPeriodicResourceGainTicks(config)),
		Priority: ActionPriorityDOT,
		OnAction: func(sim *Simulation) {
			character.ExecuteResourceGain(sim, config.resType, gainPerTick, metrics)
		},
	})
}

func startPeriodicResourceGainTicks(config resourceGainConfig) int32 {
	return int32(config.duration / config.period)
}
