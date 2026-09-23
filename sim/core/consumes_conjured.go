package core

import (
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

var ConjuredAuraTag = "Conjured"

func registerConjuredCD(agent Agent, consumes *proto.ConsumesSpec) {
	character := agent.GetCharacter()

	for _, conjuredId := range consumes.ConjuredItems {
		var conjuredMCD MajorCooldown
		switch conjuredId {
		case 22788:
			conjuredMCD = makeConjuredActivationSpell(conjuredId, character)

			flameCapProc := character.RegisterSpell(SpellConfig{
				ActionID:    conjuredMCD.Spell.ActionID,
				SpellSchool: SpellSchoolFire,
				DefenseType: DefenseTypeMagic, // Flamecap Fire (28715)
				ProcMask:    ProcMaskSpellDamage,
				Flags:       SpellFlagProc,

				DamageMultiplier: 1,
				ThreatMultiplier: 1,

				ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
					spell.CalcAndDealDamage(sim, target, 40, spell.OutcomeMagicHitAndCrit)
				},
			})

			procTrigger := character.MakeProcTriggerAura(ProcTrigger{
				Name:       "Flame Cap - Proc",
				ActionID:   conjuredMCD.Spell.ActionID,
				Duration:   time.Minute * 1,
				ProcChance: 0.185,
				ProcMask:   ProcMaskMeleeWhiteHit | ProcMaskRangedAuto,
				Outcome:    OutcomeLanded,
				Callback:   CallbackOnSpellHitDealt,
				Handler: func(sim *Simulation, spell *Spell, result *SpellResult) {
					flameCapProc.Cast(sim, result.Target)
				},
			})

			flameCapAura := character.NewTemporaryStatsAura("Flame Cap", conjuredMCD.Spell.ActionID, stats.Stats{stats.FireDamage: 80}, time.Minute)
			flameCapAura.AttachDependentAura(procTrigger)

			oldApplyEffects := conjuredMCD.Spell.ApplyEffects
			conjuredMCD.Spell.ApplyEffects = func(sim *Simulation, target *Unit, spell *Spell) {
				oldApplyEffects(sim, target, spell)
				flameCapAura.Activate(sim)
			}
			conjuredMCD.Spell.RelatedSelfBuff = flameCapAura.Aura
		default:
			conjuredMCD = makeConjuredActivationSpell(conjuredId, character)
		}

		if conjuredMCD.Spell != nil {
			oldShouldActivate := conjuredMCD.ShouldActivate
			conjuredMCD.ShouldActivate = func(sim *Simulation, character *Character) bool {
				return oldShouldActivate(sim, character) && consumes.ConjuredId == conjuredId
			}
			character.AddMajorCooldown(conjuredMCD)
		}
	}

}

func makeConjuredActivationSpell(conjuredId int32, character *Character) MajorCooldown {
	conjured := GetConsumableByID(conjuredId)
	categoryCooldownDuration := TernaryDuration(conjured.CategoryCooldownDuration > 0, conjured.CategoryCooldownDuration, time.Minute*2)
	mcd := makeConjuredActivationSpellInternal(conjured, character)

	if mcd.Spell != nil {
		mcd.Spell.Flags |= SpellFlagConjured | SpellFlagAPL
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

func makeConjuredActivationSpellInternal(conjured Consumable, character *Character) MajorCooldown {
	cooldownDuration := TernaryDuration(conjured.CooldownDuration > 0, conjured.CooldownDuration, time.Minute*2)

	conjuredCast := CastConfig{
		CD: Cooldown{
			Timer:    character.NewTimer(),
			Duration: cooldownDuration,
		},
		SharedCD: Cooldown{
			Timer:    character.GetOrInitSpellCategoryTimer(conjured.CategoryId),
			Duration: time.Minute * 2,
		},
	}

	actionID := ActionID{ItemID: conjured.Id}
	var aura *StatBuffAura
	mcd := MajorCooldown{
		Spell: character.GetOrRegisterSpell(SpellConfig{
			ActionID: actionID,
			Flags:    SpellFlagNoOnCastComplete,
			Cast:     conjuredCast,
		}),
	}
	if conjured.BuffDuration > 0 {
		// Add stat buff aura if applicable
		aura = character.NewTemporaryStatsAura(conjured.Name, actionID, conjured.Stats, conjured.BuffDuration)
		mcd.Spell.RelatedSelfBuff = aura.Aura
		mcd.Type = aura.InferCDType()
	}
	var gains []resourceGainConfig
	resourceMetrics := make(map[proto.ResourceType]*ResourceMetrics)

	for _, effectID := range conjured.EffectIds {
		e := GetSpellEffectByID(effectID)
		resourceType := e.GetResourceType()
		if e.Type == proto.EffectType_EffectTypeResourceGain && resourceType != 0 {
			if resourceType == proto.ResourceType_ResourceTypeMana && mcd.Type != CooldownTypeSurvival {
				mcd.Type = CooldownTypeMana
			} else if resourceType == proto.ResourceType_ResourceTypeHealth {
				mcd.Type = CooldownTypeSurvival
			} else {
				mcd.Type = CooldownTypeDPS
			}
			gains = append(gains, resourceGainConfig{
				resType: resourceType,
				min:     e.MinEffectSize,
				spread:  e.EffectSpread,
			})

			if _, exists := resourceMetrics[resourceType]; !exists {
				resourceMetrics[resourceType] = character.Metrics.NewResourceMetrics(actionID, resourceType)
			}
			// Preload resource types that are found on this item
			if resourceMetrics[resourceType] == nil {
				resourceMetrics[resourceType] = character.Metrics.NewResourceMetrics(actionID, resourceType)
			}
		}
	}

	mcd.Spell.ApplyEffects = func(sim *Simulation, _ *Unit, _ *Spell) {
		if aura != nil {
			aura.Activate(sim)
		}

		for _, config := range gains {
			gain := config.min + TernaryFloat64(config.spread > 1, sim.RandomFloat(conjured.Name)*config.spread, config.spread)
			switch config.resType {
			case proto.ResourceType_ResourceTypeHealth:
				gain *= character.PseudoStats.HealingTakenMultiplier
			case proto.ResourceType_ResourceTypeEnergy:
				// Thistle Tea 100 - 2 * max(0, CharacterLevel - 40) energy gain
				if conjured.Id == 7676 {
					gain -= 2 * max(0, CharacterLevel-40)
				}
			}
			character.ExecuteResourceGain(sim, config.resType, gain, resourceMetrics[config.resType])
		}
	}

	mcd.ShouldActivate = func(sim *Simulation, character *Character) bool {
		shouldActivate := true
		for _, config := range gains {
			switch config.resType {
			case proto.ResourceType_ResourceTypeMana:
				totalRegen := character.ManaRegenPerSecondWhileCasting() * 5
				manaGain := config.min + config.spread
				shouldActivate = character.MaxMana()-(character.CurrentMana()+totalRegen) >= manaGain
			case proto.ResourceType_ResourceTypeEnergy:
				if conjured.Id == 7676 {
					gain := (config.min + config.spread) - 2*max(0, CharacterLevel-40)
					shouldActivate = character.MaximumEnergy()-(character.CurrentEnergy()) >= gain
				}
			}
		}
		return shouldActivate
	}

	return mcd

}
