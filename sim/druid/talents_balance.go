package druid

func (druid *Druid) registerBalanceTalents() {
	// Tier 1
	druid.applyImprovedWrath()
	druid.applyGenesis()

	// Tier 2
	druid.applyMoonglow()
	druid.applyImprovedMoonfire()
	druid.applyNaturesMajesty()
	druid.applyNaturesReach()

	// Tier 3
	druid.applyImprovedEntanglingRoots()
	druid.applyNaturesSplendor()

	// Tier 4
	druid.applyInsectSwarm()
	druid.applyVengeance()
	druid.applyImprovedStarfire()

	// Tier 5
	druid.applyOvergrowth()
	druid.applyNaturesGrace()
	druid.applyEclipse()

	// Tier 6
	druid.applyMoonfury()

	// Tier 7
	// Moonkin Form implemented in forms.go
}

// TODO: To be implemented.
func (druid *Druid) applyMoonfury() {
	if druid.Talents.Moonfury == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.Moonfury == 0 {
	// 	return
	// }
	//
	// // Forever states Moonfury as +2% damage per rank to the Arcane|Nature schools (mask 72)
	// // rather than as a spell modifier; the class mask keeps it on the TBC spell list.
	// druid.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  DruidSpellWrath | DruidSpellStarfire | DruidSpellMoonfire,
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: spellData.Moonfury.Effect(dbcenums.A_MOD_DAMAGE_PERCENT_DONE, 72).FractionAt(druid.Talents.Moonfury),
	// })
}

// TODO: To be implemented.
func (druid *Druid) applyMoonglow() {
	if druid.Talents.Moonglow == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.Moonglow == 0 {
	// 	return
	// }
	//
	// druid.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  DruidSpellMoonfire | DruidSpellStarfire | DruidSpellWrath | DruidSpellHealingTouch | DruidSpellRegrowth | DruidSpellRejuvenation,
	// 	FloatValue: -0.03 * float64(druid.Talents.Moonglow),
	// 	Kind:       core.SpellMod_PowerCost_Pct_Add,
	// })
}

// TODO: To be implemented.
func (druid *Druid) applyNaturesGrace() {
	if !druid.Talents.NaturesGrace {
		return
	}

	// The TBC implementation, kept for the port:
	// if !druid.Talents.NaturesGrace {
	// 	return
	// }
	//
	// lastProcAt := time.Duration(-1)
	//
	// aura := druid.RegisterAura(core.Aura{
	// 	Label:    "Nature's Grace",
	// 	ActionID: core.ActionID{SpellID: 16886},
	// 	Duration: time.Second * 15,
	// 	OnReset: func(aura *core.Aura, sim *core.Simulation) {
	// 		lastProcAt = -1
	// 	},
	// 	OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
	// 		if spell.CurCast.CastTime == 0 {
	// 			return
	// 		}
	// 		// Only consume if the aura was already active when this cast started;
	// 		// a cast in flight when the proc landed did not benefit from it.
	// 		if aura.TimeActive(sim) < spell.CurCast.CastTime {
	// 			return
	// 		}
	// 		// A proc that landed during this cast re-arms the buff for the next
	// 		// cast instead of being consumed by this one.
	// 		if lastProcAt > sim.CurrentTime-spell.CurCast.CastTime {
	// 			return
	// 		}
	//
	// 		aura.Deactivate(sim)
	// 	},
	// }).AttachSpellMod(core.SpellModConfig{
	// 	ClassMask: DruidSpellStarfire | DruidSpellWrath,
	// 	Kind:      core.SpellMod_CastTime_Flat,
	// 	TimeValue: time.Millisecond * -500,
	// })
	//
	// druid.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:               "Nature's Grace Trigger",
	// 	Callback:           core.CallbackOnSpellHitDealt,
	// 	ClassSpellMask:     DruidSpellWrath | DruidSpellStarfire,
	// 	Outcome:            core.OutcomeCrit,
	// 	TriggerImmediately: true,
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		lastProcAt = sim.CurrentTime
	// 		aura.Activate(sim)
	// 	},
	// })
}

// TODO: To be implemented.
func (druid *Druid) applyVengeance() {
	if druid.Talents.Vengeance == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.Vengeance == 0 {
	// 	return
	// }
	//
	// druid.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  DruidSpellWrath | DruidSpellStarfire | DruidSpellMoonfire,
	// 	Kind:       core.SpellMod_CritMultiplier_Flat,
	// 	FloatValue: spellData.Vengeance.FractionAt(druid.Talents.Vengeance),
	// })
}

func (druid *Druid) applyNaturesReach() {
	if druid.Talents.NaturesReach == 0 {
		return
	}

	// druid.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  DruidSpellsBalance | DruidSpellFearieFireFeral,
	// 	Kind:       ****BONUS RANGE**** most likely irrelevant for sim
	// 	FloatValue: 10.0 * float64(druid.Talents.NaturesReach),
	// })
}

func (druid *Druid) applyInsectSwarm() {
	if !druid.Talents.InsectSwarm {
		return
	}

	druid.registerInsectSwarmSpell()
}

// TODO: To be implemented.
func (druid *Druid) applyImprovedMoonfire() {
	if druid.Talents.ImprovedMoonfire == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.ImprovedMoonfire == 0 {
	// 	return
	// }
	//
	// // 5% per point damage increase to Moonfire and its DoT
	// druid.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  DruidSpellMoonfire,
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: spellData.ImprovedMoonfire.Effect(dbcenums.A_ADD_PCT_MODIFIER, spelldata.SPELLMOD_DAMAGE).FractionAt(druid.Talents.ImprovedMoonfire),
	// })
	//
	// // 5% per point chance to crit with Moonfire
	// druid.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  DruidSpellMoonfire,
	// 	Kind:       core.SpellMod_BonusCrit_Percent,
	// 	FloatValue: spellData.ImprovedMoonfire.Effect(dbcenums.A_ADD_FLAT_MODIFIER, spelldata.SPELLMOD_CRITICAL_CHANCE).ValueAt(druid.Talents.ImprovedMoonfire),
	// })
}

// applyImprovedWrath implements Improved Wrath, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyImprovedWrath() {
	if druid.Talents.ImprovedWrath == 0 {
		return
	}
}

// applyGenesis implements Genesis, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyGenesis() {
	if druid.Talents.Genesis == 0 {
		return
	}
}

// applyNaturesMajesty implements Nature's Majesty, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyNaturesMajesty() {
	if druid.Talents.NaturesMajesty == 0 {
		return
	}
}

// applyImprovedEntanglingRoots implements Improved Entangling Roots, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyImprovedEntanglingRoots() {
	if druid.Talents.ImprovedEntanglingRoots == 0 {
		return
	}
}

// applyNaturesSplendor implements Nature's Splendor, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyNaturesSplendor() {
	if !druid.Talents.NaturesSplendor {
		return
	}
}

// applyImprovedStarfire implements Improved Starfire, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyImprovedStarfire() {
	if druid.Talents.ImprovedStarfire == 0 {
		return
	}
}

// applyOvergrowth implements Overgrowth, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyOvergrowth() {
	if druid.Talents.Overgrowth == 0 {
		return
	}
}

// applyEclipse implements Eclipse, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyEclipse() {
	if druid.Talents.Eclipse == 0 {
		return
	}
}
