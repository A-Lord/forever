package shaman

func (shaman *Shaman) registerElementalTalents() {
	// Tier 1
	shaman.applyConvection()
	shaman.applyConcussion()

	// Tier 2
	shaman.applyElementalWarding()
	shaman.applyReverberation()
	shaman.applyCallOfFlame()
	shaman.applyElementalDevastation()

	// Tier 3
	shaman.applyElementalFocus()
	shaman.applyElementalFury()

	// Tier 4
	shaman.applyImprovedFireNova()
	shaman.applyEyeOfTheStorm()
	shaman.applyCallOfThunder()

	// Tier 5
	shaman.applyElementalReach()
	shaman.applyLightningOverload()
	shaman.applyEarthbound()

	// Tier 6
	shaman.applyElementalAlacrity()

	// Tier 7
	shaman.applyLavaBurst()
}

// TODO: To be implemented. Port the TBC Call Of Flame implementation below; not yet verified against the Forever client.
func (shaman *Shaman) applyCallOfFlame() {
	if shaman.Talents.CallOfFlame == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if shaman.Talents.CallOfFlame == 0 {
	// 	return
	// }
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: spellData.CallOfFlame.Effect(dbcenums.A_ADD_PCT_MODIFIER, spelldata.SPELLMOD_DAMAGE).FractionAt(shaman.Talents.CallOfFlame),
	// 	ClassMask:  SpellMaskFireTotem,
	// })
}

// TODO: To be implemented. Port the TBC Call Of Thunder implementation below; not yet verified against the Forever client.
func (shaman *Shaman) applyCallOfThunder() {
	if !shaman.Talents.CallOfThunder {
		return
	}

	// The TBC implementation, kept for the port:
	// if !shaman.Talents.CallOfThunder {
	// 	return
	// }
	// // TODO: Forever collapses Call of Thunder from 2 ranks to 1; the confirmed crit bonus
	// // for the new single-rank version is unknown, so it is pinned to 0 until the Forever
	// // tooltip is known.
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_BonusCrit_Percent,
	// 	FloatValue: 0,
	// 	ClassMask:  SpellMaskLightningBolt | SpellMaskChainLightning | SpellMaskOverload,
	// })
}

// TODO: To be implemented. Port the TBC Concussion implementation below; not yet verified against the Forever client.
func (shaman *Shaman) applyConcussion() {
	if shaman.Talents.Concussion == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if shaman.Talents.Concussion == 0 {
	// 	return
	// }
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: spellData.Concussion.Effect(dbcenums.A_ADD_PCT_MODIFIER, spelldata.SPELLMOD_DAMAGE).FractionAt(shaman.Talents.Concussion),
	// 	ClassMask:  SpellMaskLightningBolt | SpellMaskChainLightning | SpellMaskOverload | SpellMaskShock,
	// })
}

// TODO: To be implemented. Port the TBC Convection implementation below; not yet verified against the Forever client.
func (shaman *Shaman) applyConvection() {
	if shaman.Talents.Convection == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if shaman.Talents.Convection == 0 {
	// 	return
	// }
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_PowerCost_Pct_Add,
	// 	FloatValue: -0.02 * float64(shaman.Talents.Convection),
	// 	ClassMask:  SpellMaskLightningBolt | SpellMaskChainLightning | SpellMaskOverload | SpellMaskShock,
	// })
}

// TODO: To be implemented. Port the TBC Elemental Devastation implementation below; not yet verified against the Forever client.
func (shaman *Shaman) applyElementalDevastation() {
	if shaman.Talents.ElementalDevastation == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if shaman.Talents.ElementalDevastation == 0 {
	// 	return
	// }
	// critBuffAura := shaman.RegisterAura(core.Aura{
	// 	Label:    "Elemental Devastation",
	// 	ActionID: core.ActionID{SpellID: 29178},
	// 	Duration: time.Second * 10,
	// }).AttachSpellMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_BonusCrit_Percent,
	// 	FloatValue: 3 * float64(shaman.Talents.ElementalDevastation),
	// 	ProcMask:   core.ProcMaskMelee,
	// })
	// shaman.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:             "Elemental Devastation Trigger",
	// 	CanProcFromProcs: true, // 29179/29180/30160 carry the bit.
	// 	Callback:         core.CallbackOnSpellHitDealt,
	// 	ProcMask:         core.ProcMaskSpellDamage,
	// 	Outcome:          core.OutcomeCrit,
	// 	Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
	// 		critBuffAura.Activate(sim)
	// 	},
	// })
	//
}

// TODO: To be implemented. Port the TBC Elemental Focus implementation below; not yet verified against the Forever client.
func (shaman *Shaman) applyElementalFocus() {
	if !shaman.Talents.ElementalFocus {
		return
	}

	// The TBC implementation, kept for the port:
	// if !shaman.Talents.ElementalFocus {
	// 	return
	// }
	// var triggeringSpell *core.Spell
	// var triggerTime time.Duration
	//
	// canConsumeSpells := SpellMaskLightningBolt | SpellMaskChainLightning | (SpellMaskShock & ^SpellMaskFlameShockDot)
	//
	// maxStacks := int32(2)
	//
	// clearcastingAura := shaman.RegisterAura(core.Aura{
	// 	Label:     "Clearcasting",
	// 	ActionID:  core.ActionID{SpellID: 16246},
	// 	Duration:  time.Second * 15,
	// 	MaxStacks: maxStacks,
	// 	OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
	// 		if !spell.Matches(canConsumeSpells) {
	// 			return
	// 		}
	// 		if spell == triggeringSpell && sim.CurrentTime == triggerTime {
	// 			return
	// 		}
	// 		aura.RemoveStack(sim)
	// 	},
	// }).AttachSpellMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_PowerCost_Pct_Add,
	// 	ClassMask:  canConsumeSpells,
	// 	FloatValue: -0.4,
	// })
	//
	// shaman.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:               "Elemental Focus",
	// 	Callback:           core.CallbackOnSpellHitDealt,
	// 	CanProcFromProcs:   true, // 16164 carries the bit: Lightning Overload crits count.
	// 	Outcome:            core.OutcomeCrit,
	// 	TriggerImmediately: true,
	//
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		if !spell.SpellSchool.Matches(core.SpellSchoolElemental) {
	// 			return
	// 		}
	// 		triggeringSpell = spell
	// 		triggerTime = sim.CurrentTime
	// 		clearcastingAura.Activate(sim)
	// 		clearcastingAura.SetStacks(sim, maxStacks)
	// 	},
	// })
}

// TODO: To be implemented. Port the TBC Elemental Fury implementation below; not yet verified against the Forever client.
func (shaman *Shaman) applyElementalFury() {
	if shaman.Talents.ElementalFury == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if shaman.Talents.ElementalFury == 0 {
	// 	return
	// }
	// // The talent's class mask (16089) also covers Flametongue Attack (bit 21) and Frostbrand
	// // Attack (bit 24): shamans' Flametongue Weapon hits crit for 2.0x in logs while the same
	// // attack granted by Flametongue Totem crits for 1.5x on other players.
	// //
	// // TODO: Forever expands Elemental Fury from 1 rank to 5; the per-rank crit-multiplier
	// // bonus is unconfirmed, so it is pinned to 0 until the Forever tooltip is known.
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_CritMultiplier_Flat,
	// 	FloatValue: 0,
	// 	ClassMask:  SpellMaskFireTotem | SpellMaskFire | SpellMaskNature | SpellMaskFrost | SpellMaskFlametongueWeapon | SpellMaskFrostbrandWeapon,
	// })
}
func (shaman *Shaman) applyLightningOverload() {
	if shaman.Talents.LightningOverload == 0 {
		return
	}
	// In shaman.go -> GetOverloadChance()
}

// TODO: To be implemented. Port the TBC Reverberation implementation below; not yet verified against the Forever client.
func (shaman *Shaman) applyReverberation() {
	if shaman.Talents.Reverberation == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if shaman.Talents.Reverberation == 0 {
	// 	return
	// }
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_Cooldown_Flat,
	// 	TimeValue: time.Duration(-200*shaman.Talents.Reverberation) * time.Millisecond,
	// 	ClassMask: SpellMaskShock,
	// })
}

// applyElementalAlacrity implements Elemental Alacrity, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyElementalAlacrity() {
	if shaman.Talents.ElementalAlacrity == 0 {
		return
	}
}

// applyElementalReach implements Elemental Reach, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyElementalReach() {
	if shaman.Talents.ElementalReach == 0 {
		return
	}
}

// applyElementalWarding implements Elemental Warding, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyElementalWarding() {
	if shaman.Talents.ElementalWarding == 0 {
		return
	}
}

// applyEyeOfTheStorm implements Eye of the Storm, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyEyeOfTheStorm() {
	if shaman.Talents.EyeOfTheStorm == 0 {
		return
	}
}

// applyImprovedFireNova implements Improved Fire Nova, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyImprovedFireNova() {
	if shaman.Talents.ImprovedFireNova == 0 {
		return
	}
}

// applyEarthbound implements Earthbound, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyEarthbound() {
	if !shaman.Talents.Earthbound {
		return
	}
}

// applyLavaBurst implements Lava Burst, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyLavaBurst() {
	if !shaman.Talents.LavaBurst {
		return
	}
}
