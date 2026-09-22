package hunter

func (hunter *Hunter) registerBeastMasteryTalents() {
	// Tier 1
	hunter.registerDeadlyAspects()
	hunter.registerEnduranceTraining()

	// Tier 2
	hunter.registerFocusedFire()
	hunter.registerImprovedAspectOfTheMonkey()
	hunter.registerPathfinding()
	hunter.registerImprovedRevivePet()

	// Tier 3
	hunter.registerBestialSwiftness()
	hunter.registerUnleashedFury()

	// Tier 4
	hunter.registerImprovedMendPet()
	hunter.registerFerocity()
	hunter.registerSummonHawk()

	// Tier 5
	hunter.registerSpiritBond()
	hunter.registerIntimidation()
	// Bestial Discipline implemented in pet.go

	// Tier 6
	hunter.registerFrenzy()

	// Tier 7
	hunter.registerBestialWrath()
}

// TODO: To be implemented.
func (hunter *Hunter) registerEnduranceTraining() {
	if hunter.Pet == nil || hunter.Talents.EnduranceTraining == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if hunter.Pet == nil || hunter.Talents.EnduranceTraining == 0 {
	// 	return
	// }
	//
	// hunter.Pet.StatDependencyManager.EnableDynamicStatDep(
	// 	hunter.Pet.NewDynamicMultiplyStat(stats.Health, spellData.EnduranceTraining.Effect(dbcenums.A_ADD_FLAT_MODIFIER, spelldata.SPELLMOD_ALL_EFFECTS).MultiplierAt(hunter.Talents.EnduranceTraining)),
	// )
	//
	// // TODO: Forever drops the hunter's own health bonus from Endurance Training; the spell
	// // carries only the pet modifier applied above (+3% per rank), so this is pinned to the
	// // untalented 1.0.
	// ownHealthMultiplier := 1.0
	// hunter.StatDependencyManager.EnableDynamicStatDep(
	// 	hunter.NewDynamicMultiplyStat(stats.Health, ownHealthMultiplier),
	// )
}

// TODO: To be implemented.
func (hunter *Hunter) registerFocusedFire() {
	if hunter.Pet == nil || hunter.Talents.FocusedFire == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if hunter.Pet == nil || hunter.Talents.FocusedFire == 0 {
	// 	return
	// }
	//
	// // The single dummy effect is the 1% per rank damage bonus.
	// hunter.PseudoStats.DamageDealtMultiplier *= spellData.FocusedFire.EffectAt(1).MultiplierAt(hunter.Talents.FocusedFire)
	//
	// // TODO: Forever drops Focused Fire's Kill Command crit bonus; the spell carries only the
	// // dummy used above, so the pet crit bonus is pinned to the untalented 0.
	// killCommandBonusCrit := 0.0
	// hunter.Pet.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_BonusCrit_Percent,
	// 	ClassMask:  HunterSpellKillCommandPet,
	// 	FloatValue: killCommandBonusCrit,
	// })
}

// TODO: To be implemented.
func (hunter *Hunter) registerUnleashedFury() {
	if hunter.Pet == nil || hunter.Talents.UnleashedFury == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if hunter.Pet == nil || hunter.Talents.UnleashedFury == 0 {
	// 	return
	// }
	//
	// hunter.Pet.PseudoStats.DamageDealtMultiplier *= spellData.UnleashedFury.MultiplierAt(hunter.Talents.UnleashedFury)
}

// TODO: To be implemented.
func (hunter *Hunter) registerFerocity() {
	if hunter.Pet == nil || hunter.Talents.Ferocity == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if hunter.Pet == nil || hunter.Talents.Ferocity == 0 {
	// 	return
	// }
	//
	// hunter.Pet.AddStats(stats.Stats{
	// 	stats.PhysicalCritPercent: spellData.Ferocity.ValueAt(hunter.Talents.Ferocity),
	// 	stats.SpellCritPercent:    spellData.Ferocity.ValueAt(hunter.Talents.Ferocity),
	// })
}

// TODO: To be implemented.
func (hunter *Hunter) registerFrenzy() {
	if hunter.Pet == nil || hunter.Talents.Frenzy == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if hunter.Pet == nil || hunter.Talents.Frenzy == 0 {
	// 	return
	// }
	//
	// frenzy := hunter.Pet.RegisterAura(core.Aura{
	// 	Label:    "Frenzy Effect",
	// 	ActionID: core.ActionID{SpellID: 19615},
	// 	Duration: time.Second * 8,
	// }).AttachMultiplyMeleeSpeed(1.3)
	//
	// hunter.Pet.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:       "Frenzy",
	// 	Callback:   core.CallbackOnSpellHitDealt,
	// 	Outcome:    core.OutcomeCrit,
	// 	ProcChance: spellData.Frenzy.FractionAt(hunter.Talents.Frenzy),
	//
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		frenzy.Activate(sim)
	// 	},
	// })
}

// TODO: To be implemented.
func (hunter *Hunter) registerBestialWrath() {
	if hunter.Pet == nil || !hunter.Talents.BestialWrath {
		return
	}

	// The TBC implementation, kept for the port:
	// if hunter.Pet == nil || !hunter.Talents.BestialWrath {
	// 	return
	// }
	//
	// actionID := core.ActionID{SpellID: 19574}
	//
	// hunter.Pet.BestialWrathAura = hunter.Pet.RegisterAura(core.Aura{
	// 	Label:    "Bestial Wrath",
	// 	ActionID: actionID,
	// 	Duration: time.Second * 18,
	// }).AttachMultiplicativePseudoStatBuff(
	// 	&hunter.Pet.PseudoStats.DamageDealtMultiplier, 1.5,
	// )
	//
	// hunter.BestialWrath = hunter.RegisterSpell(core.SpellConfig{
	// 	ActionID:       actionID,
	// 	SpellSchool:    core.SpellSchoolPhysical,
	// 	ClassSpellMask: HunterSpellBestialWrath,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		BaseCostPercent: 10,
	// 	},
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			NonEmpty: true,
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    hunter.NewTimer(),
	// 			Duration: time.Minute * 2,
	// 		},
	// 	},
	//
	// 	ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
	// 		return hunter.GCD.IsReady(sim)
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
	// 		hunter.Pet.BestialWrathAura.Activate(sim)
	// 	},
	// })
	//
	// hunter.AddMajorCooldown(core.MajorCooldown{
	// 	Spell: hunter.BestialWrath,
	// 	Type:  core.CooldownTypeDPS,
	// })
}

// registerDeadlyAspects implements Deadly Aspects, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerDeadlyAspects() {
	if hunter.Talents.DeadlyAspects == 0 {
		return
	}
}

// registerImprovedAspectOfTheMonkey implements Improved Aspect of the Monkey, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerImprovedAspectOfTheMonkey() {
	if hunter.Talents.ImprovedAspectOfTheMonkey == 0 {
		return
	}
}

// registerPathfinding implements Pathfinding, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerPathfinding() {
	if hunter.Talents.Pathfinding == 0 {
		return
	}
}

// registerImprovedRevivePet implements Improved Revive Pet, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerImprovedRevivePet() {
	if hunter.Talents.ImprovedRevivePet == 0 {
		return
	}
}

// registerBestialSwiftness implements Bestial Swiftness, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerBestialSwiftness() {
	if !hunter.Talents.BestialSwiftness {
		return
	}
}

// registerImprovedMendPet implements Improved Mend Pet, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerImprovedMendPet() {
	if hunter.Talents.ImprovedMendPet == 0 {
		return
	}
}

// registerSummonHawk implements Summon Hawk, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerSummonHawk() {
	if !hunter.Talents.SummonHawk {
		return
	}
}

// registerSpiritBond implements Spirit Bond, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerSpiritBond() {
	if hunter.Talents.SpiritBond == 0 {
		return
	}
}

// registerIntimidation implements Intimidation, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerIntimidation() {
	if !hunter.Talents.Intimidation {
		return
	}
}
