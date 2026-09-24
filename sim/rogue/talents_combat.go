package rogue

import "github.com/wowsims/forever/sim/core/stats"

func (rogue *Rogue) registerCombatTalents() {
	// Tier 1
	rogue.registerImprovedGouge()
	rogue.registerImprovedSinisterStrike()
	rogue.registerLightningReflexes()

	// Tier 2
	// Improved Slice and Dice implemented in slice_and_dice.go
	// Deflection NYI
	rogue.registerPrecision()

	// Tier 3
	// None in this tier implemented

	// Tier 4
	// Improved Kick NYI
	rogue.registerDualWieldSpecialization()

	// Tier 5
	rogue.registerBladeFlurry()

	// Tier 6
	// Blade Twisting NYI
	rogue.registerWeaponExpertise()
	rogue.registerAggression()

	// Tier 7
	rogue.registerAdrenalineRush()
	// Nerves of Steel NYI

	// Forever additions, not yet implemented.
	rogue.registerDeflection()
	rogue.registerEndurance()
	rogue.registerFlawlessExecution()
	rogue.registerHackAndSlash()
	rogue.registerImprovedKick()
	rogue.registerImprovedSprint()
	rogue.registerRiposte()
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerImprovedGouge() {
	if rogue.Talents.ImprovedGouge == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.ImprovedGouge == 0 {
	// 	return
	// }
	//
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_Cooldown_Flat,
	// 	ClassMask: RogueSpellGouge,
	// 	TimeValue: time.Millisecond * 500 * time.Duration(rogue.Talents.ImprovedGouge),
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerImprovedSinisterStrike() {
	if rogue.Talents.ImprovedSinisterStrike == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.ImprovedSinisterStrike == 0 {
	// 	return
	// }
	//
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_PowerCost_Flat,
	// 	ClassMask: RogueSpellSinisterStrike,
	// 	IntValue:  []int32{0, -3, -5}[rogue.Talents.ImprovedSinisterStrike],
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerLightningReflexes() {
	if rogue.Talents.LightningReflexes == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.LightningReflexes == 0 {
	// 	return
	// }
	//
	// rogue.AddStat(stats.DodgeRating, float64(rogue.Talents.LightningReflexes)*core.DodgeRatingPerDodgePercent)
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerPrecision() {
	if rogue.Talents.Precision == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.Precision == 0 {
	// 	return
	// }
	//
	// rogue.AddStat(stats.PhysicalHitPercent, float64(rogue.Talents.Precision))
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerDualWieldSpecialization() {
	if rogue.Talents.DualWieldSpecialization == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.DualWieldSpecialization == 0 {
	// 	return
	// }
	//
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	ProcMask:   core.ProcMaskMeleeOH,
	// 	FloatValue: spellData.DualWieldSpecialization.FractionAt(rogue.Talents.DualWieldSpecialization),
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerBladeFlurry() {
	if !rogue.Talents.BladeFlurry {
		return
	}

	// The TBC implementation, kept for the port:
	// if !rogue.Talents.BladeFlurry {
	// 	return
	// }
	//
	// var curDmg float64
	// bfHit := rogue.GetOrRegisterSpell(core.SpellConfig{
	// 	ActionID:    core.ActionID{SpellID: 22482},
	// 	SpellSchool: core.SpellSchoolPhysical,
	// 	ProcMask:    core.ProcMaskEmpty, // No proc mask, so it won't proc itself.
	// 	Flags:       core.SpellFlagIgnoreResists | core.SpellFlagIgnoreModifiers | core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete,
	//
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		spell.CalcAndDealDamage(sim, target, curDmg, spell.OutcomeAlwaysHit)
	// 	},
	// })
	//
	// rogue.BladeFlurryAura = rogue.GetOrRegisterAura(core.Aura{
	// 	Label:    "Blade Flurry",
	// 	ActionID: core.ActionID{SpellID: 13877},
	// 	Duration: time.Second * 15,
	//
	// 	OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		if sim.ActiveTargetCount() < 2 {
	// 			return
	// 		}
	//
	// 		if result.Damage == 0 || !spell.ProcMask.Matches(core.ProcMaskMelee) {
	// 			return
	// 		}
	//
	// 		curDmg = result.Damage
	// 		bfHit.Cast(sim, rogue.Env.NextActiveTargetUnit(result.Target))
	// 		bfHit.SpellMetrics[result.Target.UnitIndex].Casts--
	// 	},
	// }).AttachMultiplyAttackSpeed(1.2)
	//
	// rogue.BladeFlurry = rogue.GetOrRegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 13877},
	// 	ClassSpellMask: RogueSpellBladeFlurry,
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: time.Second,
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    rogue.NewTimer(),
	// 			Duration: time.Minute * 2,
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost: 25,
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		rogue.BladeFlurryAura.Activate(sim)
	// 	},
	// })
	//
	// rogue.AddMajorCooldown(core.MajorCooldown{
	// 	Spell: rogue.BladeFlurry,
	// 	Type:  core.CooldownTypeDPS,
	// })
}

// "Reduces the chance for your attacks to be Dodged or Parried by 1%/2%." A_MOD_EXPERTISE
// carries the percent directly.
func (rogue *Rogue) registerWeaponExpertise() {
	if rogue.Talents.WeaponExpertise == 0 {
		return
	}

	rogue.AddStat(stats.ExpertisePercent, spellData.WeaponExpertise.ValueAt(rogue.Talents.WeaponExpertise))
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerAggression() {
	if rogue.Talents.Aggression == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.Aggression == 0 {
	// 	return
	// }
	//
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	ClassMask:  RogueSpellSinisterStrike | RogueSpellBackstab | RogueSpellEviscerate,
	// 	FloatValue: spellData.Aggression.FractionAt(rogue.Talents.Aggression),
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerAdrenalineRush() {
	if !rogue.Talents.AdrenalineRush {
		return
	}

	// The TBC implementation, kept for the port:
	// if !rogue.Talents.AdrenalineRush {
	// 	return
	// }
	//
	// rogue.AdrenalineRushAura = rogue.GetOrRegisterAura(core.Aura{
	// 	Label:    "Adrenaline Rush",
	// 	ActionID: core.ActionID{SpellID: 13750},
	// 	Duration: time.Second * 15,
	//
	// 	OnGain: func(aura *core.Aura, sim *core.Simulation) {
	// 		rogue.MultiplyEnergyRegenSpeed(sim, 2)
	// 	},
	// 	OnExpire: func(aura *core.Aura, sim *core.Simulation) {
	// 		rogue.MultiplyEnergyRegenSpeed(sim, 0.5)
	// 	},
	// })
	//
	// rogue.AdrenalineRush = rogue.GetOrRegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 13750},
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: RogueSpellAdrenalineRush,
	//
	// 	Cast: core.CastConfig{
	// 		IgnoreHaste: true,
	// 		CD: core.Cooldown{
	// 			Timer:    rogue.NewTimer(),
	// 			Duration: time.Minute * 5,
	// 		},
	// 		DefaultCast: core.Cast{
	// 			GCD: time.Second,
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		rogue.AdrenalineRushAura.Activate(sim)
	// 	},
	// })
	//
	// rogue.AddMajorCooldown(core.MajorCooldown{
	// 	Spell: rogue.AdrenalineRush,
	// 	Type:  core.CooldownTypeDPS,
	// })
}

// registerDeflection implements Deflection, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerDeflection() {
	if rogue.Talents.Deflection == 0 {
		return
	}
}

// registerEndurance implements Endurance, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerEndurance() {
	if rogue.Talents.Endurance == 0 {
		return
	}
}

// registerFlawlessExecution implements Flawless Execution, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerFlawlessExecution() {
	if !rogue.Talents.FlawlessExecution {
		return
	}
}

// registerHackAndSlash implements Hack and Slash, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerHackAndSlash() {
	if rogue.Talents.HackAndSlash == 0 {
		return
	}
}

// registerImprovedKick implements Improved Kick, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerImprovedKick() {
	if rogue.Talents.ImprovedKick == 0 {
		return
	}
}

// registerImprovedSprint implements Improved Sprint, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerImprovedSprint() {
	if rogue.Talents.ImprovedSprint == 0 {
		return
	}
}

// registerRiposte implements Riposte, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerRiposte() {
	if !rogue.Talents.Riposte {
		return
	}
}
