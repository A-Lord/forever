package shaman

func (shaman *Shaman) registerEnhancementTalents() {
	// Tier 1
	shaman.applyEarthsGrasp()
	shaman.applyThunderingStrikes()
	shaman.applyAncestralKnowledge()

	// Tier 2
	shaman.applyGuardianTotems()
	shaman.applyMentalDexterity()
	shaman.applyImprovedGhostWolf()
	shaman.applyImprovedLightningShield()

	// Tier 3
	shaman.applyElementalWeapons()
	shaman.applyShamanisticFocus()
	shaman.applyAnticipation()

	// Tier 4
	shaman.applyToughness()
	shaman.applyFlurry()
	shaman.applyStormstrike()

	// Tier 5
	shaman.applySpiritWeapons()
	shaman.applyMentalQuickness()
	shaman.applyImprovedStormstrike()

	// Tier 6
	shaman.applyMaelstromWeapon()

	// Tier 7
	shaman.applyRageOfTheFarseer()
}

// TODO: To be implemented. Port the TBC Ancestral Knowledge implementation below; not yet verified against the Forever game files.
func (shaman *Shaman) applyAncestralKnowledge() {
	if shaman.Talents.AncestralKnowledge == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if shaman.Talents.AncestralKnowledge == 0 {
	// 	return
	// }
	// shaman.MultiplyStat(stats.Mana, spellData.AncestralKnowledge.MultiplierAt(shaman.Talents.AncestralKnowledge))
}

// TODO: To be implemented. Port the TBC Elemental Weapons implementation below; not yet verified against the Forever game files.
func (shaman *Shaman) applyElementalWeapons() {
	if shaman.Talents.ElementalWeapons == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if shaman.Talents.ElementalWeapons == 0 {
	// 	return
	// }
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: [4]float64{0.0, 0.07, 0.14, 0.2}[shaman.Talents.ElementalWeapons],
	// 	ClassMask:  SpellMaskRockbiterWeapon,
	// })
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: [4]float64{0.0, 0.13, 0.27, 0.4}[shaman.Talents.ElementalWeapons],
	// 	ClassMask:  SpellMaskWindfuryWeapon,
	// })
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: 0.05 * float64(shaman.Talents.ElementalWeapons),
	// 	ClassMask:  SpellMaskFlametongueWeapon | SpellMaskFrostbrandWeapon,
	// })
}

// TODO: To be implemented. Port the TBC Flurry implementation below; not yet verified against the Forever game files.
func (shaman *Shaman) applyFlurry() {
	if shaman.Talents.Flurry == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if shaman.Talents.Flurry == 0 {
	// 	return
	// }
	//
	// flurryICD := &core.Cooldown{
	// 	Timer:    shaman.NewTimer(),
	// 	Duration: 500 * time.Millisecond,
	// }
	//
	// attackSpeed := 1.05 +
	// 	0.05*float64(shaman.Talents.Flurry)
	//
	// flurryAura := shaman.RegisterAura(core.Aura{
	// 	ActionID:  core.ActionID{SpellID: 16284},
	// 	Label:     "Flurry",
	// 	Duration:  time.Second * 15,
	// 	MaxStacks: 3,
	// }).AttachMultiplyMeleeSpeed(attackSpeed)
	//
	// shaman.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:             "Flurry Trigger",
	// 	Callback:         core.CallbackOnSpellHitDealt,
	// 	ProcMask:         core.ProcMaskMelee,
	// 	CanProcFromProcs: true, // 16256, 16281-16284 carry the bit.
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		if result.Outcome.Matches(core.OutcomeCrit) {
	// 			flurryAura.Activate(sim)
	// 			flurryAura.SetStacks(sim, 3)
	// 			return
	// 		}
	//
	// 		// Remove a stack.
	// 		if flurryAura.IsActive() && spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) && flurryICD.IsReady(sim) {
	// 			flurryICD.Use(sim)
	// 			flurryAura.RemoveStack(sim)
	// 		}
	// 	},
	// })
	//
}

// TODO: To be implemented. Port the TBC Improved Lightning Shield implementation below; not yet verified against the Forever game files.
func (shaman *Shaman) applyImprovedLightningShield() {
	if shaman.Talents.ImprovedLightningShield == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if shaman.Talents.ImprovedLightningShield == 0 {
	// 	return
	// }
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: spellData.ImprovedLightningShield.FractionAt(shaman.Talents.ImprovedLightningShield),
	// 	ClassMask:  SpellMaskLightningShield,
	// })
}

// TODO: To be implemented. Port the TBC Mental Quickness implementation below; not yet verified against the Forever game files.
func (shaman *Shaman) applyMentalQuickness() {
	if shaman.Talents.MentalQuickness == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if shaman.Talents.MentalQuickness == 0 {
	// 	return
	// }
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_PowerCost_Pct_Add,
	// 	FloatValue: -0.02 * float64(shaman.Talents.MentalQuickness),
	// 	SpellFlag:  SpellFlagInstant,
	// })
	//
	// // TODO: To be implemented. Forever's regenerated aura enum no longer carries the
	// // attack-power-to-spell-damage aura this talent's rank data used
	// // (A_MOD_SPELL_DAMAGE_OF_ATTACK_POWER is gone from the auto-generated table), so the
	// // spell damage conversion below is not applied.
}

// TODO: To be implemented. Port the TBC Shamanistic Focus implementation below; not yet verified against the Forever game files.
func (shaman *Shaman) applyShamanisticFocus() {
	if !shaman.Talents.ShamanisticFocus {
		return
	}

	// The TBC implementation, kept for the port:
	// if !shaman.Talents.ShamanisticFocus {
	// 	return
	// }
	// sfAura := shaman.RegisterAura(core.Aura{
	// 	Label:    "Focused",
	// 	ActionID: core.ActionID{SpellID: 43339},
	// 	Duration: time.Second * 15,
	// }).AttachSpellMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_PowerCost_Pct_Add,
	// 	FloatValue: -0.6,
	// 	ClassMask:  SpellMaskShock,
	// })
	//
	// shaman.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:             "Shamanistic Focus Trigger",
	// 	Callback:         core.CallbackOnSpellHitDealt,
	// 	ProcMask:         core.ProcMaskMelee,
	// 	CanProcFromProcs: true, // 43338 carries the bit.
	// 	Outcome:          core.OutcomeCrit,
	// 	Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
	// 		sfAura.Activate(sim)
	// 	},
	// })
	//
	// shaman.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:           "Shamanistic Focus Untrigger",
	// 	Callback:       core.CallbackOnCastComplete,
	// 	ClassSpellMask: SpellMaskShock,
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
	// 		sfAura.Deactivate(sim)
	// 	},
	// })
}

func (shaman *Shaman) applySpiritWeapons() {
	if !shaman.Talents.SpiritWeapons {
		return
	}
	//TODO ? Threat related talent
}

func (shaman *Shaman) applyStormstrike() {
	if !shaman.Talents.Stormstrike {
		return
	}
	shaman.registerStormstrikeSpell()
}

// TODO: To be implemented. Port the TBC Thundering Strikes implementation below; not yet verified against the Forever game files.
func (shaman *Shaman) applyThunderingStrikes() {
	if shaman.Talents.ThunderingStrikes == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if shaman.Talents.ThunderingStrikes == 0 {
	// 	return
	// }
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_BonusCrit_Percent,
	// 	FloatValue: spellData.ThunderingStrikes.ValueAt(shaman.Talents.ThunderingStrikes),
	// 	ProcMask:   core.ProcMaskMelee,
	// })
}

// applyEarthsGrasp implements Earth's Grasp, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyEarthsGrasp() {
	if shaman.Talents.EarthsGrasp == 0 {
		return
	}
}

// applyGuardianTotems implements Guardian Totems, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyGuardianTotems() {
	if shaman.Talents.GuardianTotems == 0 {
		return
	}
}

// applyMentalDexterity implements Mental Dexterity, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyMentalDexterity() {
	if shaman.Talents.MentalDexterity == 0 {
		return
	}
}

// applyImprovedGhostWolf implements Improved Ghost Wolf, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyImprovedGhostWolf() {
	if shaman.Talents.ImprovedGhostWolf == 0 {
		return
	}
}

// applyAnticipation implements Anticipation, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyAnticipation() {
	if shaman.Talents.Anticipation == 0 {
		return
	}
}

// applyToughness implements Toughness, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyToughness() {
	if shaman.Talents.Toughness == 0 {
		return
	}
}

// applyImprovedStormstrike implements Improved Stormstrike, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyImprovedStormstrike() {
	if shaman.Talents.ImprovedStormstrike == 0 {
		return
	}
}

// applyMaelstromWeapon implements Maelstrom Weapon, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyMaelstromWeapon() {
	if shaman.Talents.MaelstromWeapon == 0 {
		return
	}
}

// applyRageOfTheFarseer implements Rage of the Farseer, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyRageOfTheFarseer() {
	if !shaman.Talents.RageOfTheFarseer {
		return
	}
}
