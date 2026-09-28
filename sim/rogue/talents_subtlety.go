package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Package-level state the commented-out implementations used:
// var hemorrhageRank = spellData.Hemorrhage.ByID(26864)

func (rogue *Rogue) registerSubtletyTalents() {
	// Tier 1
	// Camouflage NYI
	// Master of Deception NYI
	rogue.registerOpportunity()

	// Tier 2
	// Setup NYI
	rogue.registerElusiveness()
	// Dirty Tricks NYI
	rogue.registerImprovedAmbush()

	// Tier 3
	rogue.registerInitiative()
	rogue.registerGhostlyStrike()
	// Improved Distract NYI

	// Tier 4
	// Heightened Senses NYI
	rogue.registerPremeditation()
	rogue.registerSerratedBlades()

	// Tier 5
	rogue.registerDirtyDeeds()
	rogue.registerPreparation()
	rogue.registerHemorrhage()

	// Tier 6
	rogue.registerQuietus()
	rogue.registerCutthroat()

	// Tier 7
	rogue.registerThousandCuts()
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerOpportunity() {
	if rogue.Talents.Opportunity == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.Opportunity.Rank(rogue.Talents.Opportunity))
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerInitiative() {
	if rogue.Talents.Initiative == 0 {
		return
	}

	initiativeRank := spellData.Initiative.Rank(rogue.Talents.Initiative)
	initMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: initiativeRank.ID})

	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Initiative - Trigger",
		ActionID:           core.ActionID{SpellID: initiativeRank.ID},
		Callback:           core.CallbackOnCastComplete,
		ClassSpellMask:     RogueSpellGarrote | RogueSpellAmbush,
		Outcome:            core.OutcomeLanded,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			rogue.AddComboPoints(sim, 1, initMetrics)
		},
	})
}

var ghostlyStrikeRank = spellData.GhostlyStrike.Highest()

func (rogue *Rogue) registerGhostlyStrike() {
	if !rogue.Talents.GhostlyStrike {
		return
	}

	config := spelldata.SpellConfig(&rogue.Unit, ghostlyStrikeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags |= SpellFlagBuilder
	config.ClassSpellMask = RogueSpellGhostlyStrike
	config.DamageMultiplier = core.TernaryFloat64(rogue.HasDagger(core.MainHand), ghostlyStrikeRank.EffectN(4).Percent(), ghostlyStrikeRank.EffectN(1).Percent())

	// Dodge aura not implemented; maybe if rogue tanking comes back

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target)) // Data suggests un-normalized

		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if result.Landed() {
			rogue.AddComboPoints(sim, int32(ghostlyStrikeRank.ProcEnergizeEffect().Roll(sim, core.CharacterLevel)), spell.ComboPointMetrics())
		} else {
			spell.IssueRefund(sim)
		}
	}

	rogue.GhostlyStrike = rogue.RegisterSpell(config)

	rogue.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(s *core.Simulation, slot proto.ItemSlot) {
		rogue.GhostlyStrike.DamageMultiplier = core.TernaryFloat64(rogue.HasDagger(core.MainHand), ghostlyStrikeRank.EffectN(4).Percent(), ghostlyStrikeRank.EffectN(1).Percent())
	})

	// The TBC implementation, kept for the port:
	// if !rogue.Talents.GhostlyStrike {
	// 	return
	// }
	//
	// pointMetric := rogue.NewComboPointMetrics(core.ActionID{SpellID: 14278})
	// rogue.GhostlyStrike = rogue.GetOrRegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 14278},
	// 	ClassSpellMask: RogueSpellGhostlyStrike,
	// 	ClassFlags:     spellData.GhostlyStrike.Highest().ClassFlags,
	// 	SpellSchool:    core.SpellSchoolPhysical,
	// 	DefenseType:    core.DefenseTypeMelee,
	// 	Flags:          core.SpellFlagAPL | core.SpellFlagMeleeMetrics,
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	MaxRange:       core.MaxMeleeRange,
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: time.Second,
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    rogue.NewTimer(),
	// 			Duration: time.Second * 20,
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost:   40,
	// 		Refund: 0.8,
	// 	},
	//
	// 	DamageMultiplier: 1.25,
	// 	ThreatMultiplier: 1,
	//
	// 	BonusCoefficient: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		rogue.BreakStealth(sim)
	//
	// 		// Dodge Aura NYI
	//
	// 		baseDamage := rogue.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
	// 		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
	// 		if result.Landed() {
	// 			rogue.AddComboPoints(sim, 1, pointMetric)
	// 		}
	// 	},
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerImprovedAmbush() {
	if rogue.Talents.ImprovedAmbush == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.ImprovedAmbush.Rank(rogue.Talents.ImprovedAmbush))

	// The TBC implementation, kept for the port:
	// if rogue.Talents.ImprovedAmbush == 0 {
	// 	return
	// }
	//
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_BonusCrit_Percent,
	// 	ClassMask:  RogueSpellAmbush,
	// 	FloatValue: spellData.ImprovedAmbush.ValueAt(rogue.Talents.ImprovedAmbush),
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerElusiveness() {
	if rogue.Talents.Elusiveness == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.Elusiveness.Rank(rogue.Talents.Elusiveness))

	// The TBC implementation, kept for the port:
	// if rogue.Talents.Elusiveness == 0 {
	// 	return
	// }
	//
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_Cooldown_Flat,
	// 	ClassMask: RogueSpellVanish,
	// 	TimeValue: time.Second * 45 * time.Duration(rogue.Talents.Elusiveness),
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerSerratedBlades() {
	if rogue.Talents.SerratedBlades == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.SerratedBlades.Rank(rogue.Talents.SerratedBlades))

	// The TBC implementation, kept for the port:
	// if rogue.Talents.SerratedBlades == 0 {
	// 	return
	// }
	//
	// rogue.AddStat(stats.ArmorPenetration, 186*float64(rogue.Talents.SerratedBlades))
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	ClassMask:  RogueSpellRupture,
	// 	FloatValue: spellData.SerratedBlades.Effect(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_DOT)).FractionAt(rogue.Talents.SerratedBlades),
	// })
}

func (rogue *Rogue) registerPreparation() {
	if !rogue.Talents.Preparation {
		return
	}

	// Prep purely runs a server-side script, so no client data to reference.

	config := spelldata.SpellConfig(&rogue.Unit, spellData.Preparation.Highest(), spelldata.Flags(core.SpellFlagAPL|core.SpellFlagMCD))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if rogue.ColdBlood != nil {
			rogue.ColdBlood.CD.Set(0)
		}
		if rogue.Premeditation != nil {
			rogue.Premeditation.CD.Set(0)
		}
		if rogue.Vanish != nil {
			rogue.Vanish.CD.Set(0)
		}
		if rogue.GhostlyStrike != nil {
			rogue.GhostlyStrike.CD.Set(0)
		}
		if rogue.BladeFlurry != nil {
			rogue.BladeFlurry.CD.Set(0)
		}
		if rogue.AdrenalineRush != nil {
			rogue.AdrenalineRush.CD.Set(0)
		}
	}

	rogue.Preparation = rogue.RegisterSpell(config)

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell: rogue.Preparation,
		Type:  core.CooldownTypeDPS,
	})
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerDirtyDeeds() {
	if rogue.Talents.DirtyDeeds == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.DirtyDeeds.Rank(rogue.Talents.DirtyDeeds))

	// The TBC implementation, kept for the port:
	// if rogue.Talents.DirtyDeeds == 0 {
	// 	return
	// }
	//
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_PowerCost_Flat,
	// 	ClassMask: RogueSpellGarrote,
	// 	IntValue:  -10 * rogue.Talents.DirtyDeeds,
	// })
	//
	// ddAura := rogue.GetOrRegisterAura(core.Aura{
	// 	Label:    "Dirty Deeds",
	// 	ActionID: core.ActionID{SpellID: 14083},
	// 	Duration: core.NeverExpires,
	// }).AttachSpellMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	ClassMask:  RogueSpellsAll,
	// 	FloatValue: 0.1 * float64(rogue.Talents.DirtyDeeds),
	// })
	//
	// rogue.RegisterResetEffect(func(sim *core.Simulation) {
	// 	ddAura.Deactivate(sim)
	// 	sim.RegisterExecutePhaseCallback(func(sim *core.Simulation, isExecute int32) {
	// 		if isExecute == 35 {
	// 			ddAura.Activate(sim)
	// 		}
	// 	})
	// })
}

var hemorrhageRank = spellData.Hemorrhage.Highest()

func (rogue *Rogue) registerHemorrhage() {
	if !rogue.Talents.Hemorrhage {
		return
	}

	config := spelldata.SpellConfig(&rogue.Unit, hemorrhageRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags |= SpellFlagBuilder
	config.ClassSpellMask = RogueSpellHemorrhage
	config.DamageMultiplier = core.TernaryFloat64(rogue.HasDagger(core.MainHand), hemorrhageRank.EffectN(5).Percent(), hemorrhageRank.EffectN(4).Percent())

	hemoAuras := rogue.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		hemoAura := target.GetOrRegisterAura(spelldata.AuraConfig(hemorrhageRank))
		spelldata.ParseEffects(nil, hemoAura, spellData.Hemorrhage.Highest(),
			spelldata.FromCaster(&rogue.Unit, DDBC_Hemorrhage, DDBC_Total))
		return hemoAura
	})

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := rogue.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		if result.Landed() {
			hemoAuras.Get(target).Activate(sim)
			rogue.AddComboPoints(sim, int32(hemorrhageRank.ProcEnergizeEffect().Roll(sim, core.CharacterLevel)), spell.ComboPointMetrics())
		} else {
			spell.IssueRefund(sim)
		}
	}

	rogue.Hemorrhage = rogue.GetOrRegisterSpell(config)

	rogue.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(s *core.Simulation, slot proto.ItemSlot) {
		rogue.Hemorrhage.DamageMultiplier = core.TernaryFloat64(rogue.HasDagger(core.MainHand), hemorrhageRank.EffectN(5).Percent(), hemorrhageRank.EffectN(4).Percent())
	})

	// The TBC implementation, kept for the port:
	// if !rogue.Talents.Hemorrhage {
	// 	return
	// }
	//
	// pointMetric := rogue.NewComboPointMetrics(core.ActionID{SpellID: hemorrhageRank.ID})
	// rogue.Hemorrhage = rogue.GetOrRegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: hemorrhageRank.ID},
	// 	ClassSpellMask: RogueSpellHemorrhage,
	// 	ClassFlags:     hemorrhageRank.ClassFlags,
	// 	SpellSchool:    hemorrhageRank.SpellSchool(),
	// 	DefenseType:    hemorrhageRank.DefenseTypeCore(),
	// 	Flags:          core.SpellFlagAPL | core.SpellFlagMeleeMetrics | SpellFlagBuilder,
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	MaxRange:       core.MaxMeleeRange,
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: hemorrhageRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost:   int32(hemorrhageRank.Cost()),
	// 		Refund: 0.8,
	// 	},
	//
	// 	DamageMultiplier: 1.1,
	// 	ThreatMultiplier: 1,
	//
	// 	BonusCoefficient: hemorrhageRank.DamageEffect().Coeff(),
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		rogue.BreakStealth(sim)
	//
	// 		baseDamage := rogue.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
	// 		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
	// 		if result.Landed() {
	// 			rogue.AddComboPoints(sim, 1, pointMetric)
	// 		}
	// 	},
	// })
}

var premeditationRank = spellData.Premeditation.Highest()

func (rogue *Rogue) registerPremeditation() {
	if !rogue.Talents.Premeditation {
		return
	}

	config := spelldata.SpellConfig(&rogue.Unit, premeditationRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagMCD))
	premedMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: premeditationRank.ID})

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.AddComboPoints(sim, int32(premeditationRank.ProcEnergizeEffect().Roll(sim, core.CharacterLevel)), premedMetrics)
	}

	rogue.Premeditation = rogue.GetOrRegisterSpell(config)

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell: rogue.Premeditation,
		Type:  core.CooldownTypeDPS,
	})

	// The TBC implementation, kept for the port:
	// if !rogue.Talents.Premeditation {
	// 	return
	// }
	//
	// comboMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 14183})
	// shouldTimeout := false
	//
	// premedAura := rogue.RegisterAura(core.Aura{
	// 	Label:    "Premed Timeout Aura",
	// 	Duration: time.Second * 10,
	//
	// 	OnGain: func(aura *core.Aura, sim *core.Simulation) {
	// 		shouldTimeout = true
	// 	},
	// 	OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
	// 		if spell.Flags.Matches(SpellFlagFinisher) && spell.ClassSpellMask == RogueSpellSliceAndDice {
	// 			shouldTimeout = false
	// 		}
	// 	},
	// 	OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		if spell.Flags.Matches(SpellFlagFinisher) && result.Landed() {
	// 			shouldTimeout = false
	// 		}
	// 	},
	// 	OnExpire: func(aura *core.Aura, sim *core.Simulation) {
	// 		// Remove 2 points because no finisher was casted
	// 		if shouldTimeout {
	// 			rogue.AddComboPoints(sim, -2, comboMetrics)
	// 			shouldTimeout = false
	// 		}
	// 	},
	// })
	//
	// rogue.Premeditation = rogue.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 14183},
	// 	Flags:          core.SpellFlagAPL | core.SpellFlagNoOnCastComplete,
	// 	ClassSpellMask: RogueSpellPremeditation,
	// 	ClassFlags:     spellData.Premeditation.Highest().ClassFlags,
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			Cost: 0,
	// 			GCD:  0,
	// 		},
	// 		IgnoreHaste: true,
	// 		CD: core.Cooldown{
	// 			Timer:    rogue.NewTimer(),
	// 			Duration: time.Minute * 2,
	// 		},
	// 	},
	// 	ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
	// 		return rogue.IsStealthed()
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
	// 		rogue.AddComboPoints(sim, 2, comboMetrics)
	// 		premedAura.Activate(sim)
	// 	},
	// })
	//
	// rogue.AddMajorCooldown(core.MajorCooldown{
	// 	Spell:              rogue.Premeditation,
	// 	Type:               core.CooldownTypeDPS,
	// 	Priority:           core.CooldownPriorityLow,
	// 	AllowSpellQueueing: true,
	// })
}

// registerQuietus implements Quietus, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerQuietus() {
	if rogue.Talents.Quietus == 0 {
		return
	}

	// Talent data is just dummy auras. Need to implement the effects by ClassMask

	rank := spellData.Quietus.Rank(rogue.Talents.Quietus)

	executeBonus := rogue.RegisterAura(core.Aura{
		Label:    "Quietus Damage Bonus",
		ActionID: core.ActionID{SpellID: rank.ID}, // Use the talent's spell ID.
		Duration: core.NeverExpires,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		ClassMask:  RogueSpellQuietus, // Set to the spells the bonus should affect.
		FloatValue: 0.1,
	})

	rogue.RegisterResetEffect(func(sim *core.Simulation) {
		executeBonus.Deactivate(sim)
		sim.RegisterExecutePhaseCallback(func(sim *core.Simulation, isExecute int32) {
			if isExecute == 35 { // Change to the desired threshold.
				executeBonus.Activate(sim)
			}
		})
	})
}

// registerCutthroat implements Cutthroat, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerCutthroat() {
	if rogue.Talents.Cutthroat == 0 {
		return
	}

	cutthroatTrigger := spellData.CutthroatTriggered.Highest()

	rogue.CutthroatAura = rogue.RegisterAura(core.Aura{
		Label:    "Cutthroat",
		ActionID: core.ActionID{SpellID: cutthroatTrigger.ID},
		Duration: cutthroatTrigger.Duration(),
	})

	cutthroatRank := spellData.Cutthroat.Rank(rogue.Talents.Cutthroat)
	trigger := spelldata.ProcTrigger(&rogue.Character, cutthroatRank, func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		rogue.CutthroatAura.Activate(sim)
	})
	trigger.Name = "Cutthroat Trigger"
	trigger.TriggerImmediately = true
	// trigger.RequireDamageDealt = true - Unknown
	trigger.ClassSpellMask = RogueSpellBackstab
	rogue.MakeProcTriggerAura(trigger)
}

// registerThousandCuts implements Thousand Cuts, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerThousandCuts() {
	if !rogue.Talents.ThousandCuts {
		return
	}

	aura := rogue.RegisterAura(spelldata.AuraConfig(spellData.ThousandCutsTriggered.Highest()))

	costMod := rogue.AddDynamicMod(core.SpellModConfig{
		Kind:      core.SpellMod_PowerCost_Flat,
		IntValue:  0, // Set dynamically based on aura stacks.
		ClassMask: RogueSpellHemorrhage | RogueSpellBackstab,
	})

	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Thousand Cuts - Trigger",
		ActionID:           core.ActionID{SpellID: spellData.ThousandCuts.Highest().ID},
		Callback:           core.CallbackOnPeriodicDamageDealt,
		ClassSpellMask:     RogueSpellRupture,
		TriggerImmediately: true,
		ProcChance:         1,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			aura.Activate(sim)
			aura.AddStack(sim)
			costMod.Activate()
			costMod.UpdateIntValue(aura.GetStacks() * -3)
		},
	}).ExposeToAPL(spellData.ThousandCuts.Highest().ID)

	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Thousand Cuts - Cast Watcher",
		ActionID:           core.ActionID{SpellID: spellData.ThousandCuts.Highest().ID},
		Callback:           core.CallbackOnCastComplete,
		ClassSpellMask:     RogueSpellHemorrhage | RogueSpellBackstab,
		TriggerImmediately: true,
		ProcChance:         1,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			aura.SetStacks(sim, 0)
			aura.Deactivate(sim)
			costMod.Deactivate()
		},
	})
}
