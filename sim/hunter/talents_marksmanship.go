package hunter

func (hunter *Hunter) registerMarksmanshipTalents() {
	// Tier 1
	hunter.registerHawkEye()
	hunter.registerImprovedConcussiveShot()
	hunter.registerLethalAttacks()

	// Tier 2
	hunter.registerImprovedStings()
	hunter.registerEfficiency()
	hunter.registerCarefulAim()

	// Tier 3
	hunter.registerRapidKilling()
	hunter.registerImprovedArcaneShot()
	hunter.registerLoneWolf()

	// Tier 4
	// Trueshot Aura handled as a group buff in hunter.go
	hunter.registerMortalShots()
	hunter.registerImprovedSerpentSting()

	// Tier 5
	hunter.registerRapidRecuperation()
	hunter.registerBarrage()
	hunter.registerScatterShot()

	// Tier 6
	hunter.registerRangedWeaponSpecialization()

	// Tier 7
	hunter.registerSniperShot()
}

// TODO: To be implemented.
func (hunter *Hunter) registerEfficiency() {
	if hunter.Talents.Efficiency == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if hunter.Talents.Efficiency == 0 {
	// 	return
	// }
	//
	// hunter.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_PowerCost_Pct_Add,
	// 	ClassMask:  HunterSpellsShotsAndStings,
	// 	FloatValue: -0.02 * float64(hunter.Talents.Efficiency),
	// })
}

// TODO: To be implemented.
func (hunter *Hunter) registerImprovedArcaneShot() {
	if hunter.Talents.ImprovedArcaneShot == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if hunter.Talents.ImprovedArcaneShot == 0 {
	// 	return
	// }
	//
	// hunter.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_Cooldown_Flat,
	// 	ClassMask: HunterSpellArcaneShot,
	// 	TimeValue: -core.DurationFromSeconds(0.2 * float64(hunter.Talents.ImprovedArcaneShot)),
	// })
}

// TODO: To be implemented.
func (hunter *Hunter) registerRapidKilling() {
	if hunter.Talents.RapidKilling == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if hunter.Talents.RapidKilling == 0 {
	// 	return
	// }
	//
	// hunter.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_Cooldown_Flat,
	// 	ClassMask: HunterSpellRapidFire,
	// 	TimeValue: -core.DurationFromSeconds(60 * float64(hunter.Talents.RapidKilling)),
	// })
}

// TODO: To be implemented.
func (hunter *Hunter) registerImprovedStings() {
	if hunter.Talents.ImprovedStings == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if hunter.Talents.ImprovedStings == 0 {
	// 	return
	// }
	//
	// hunter.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	ClassMask:  HunterSpellSerpentSting,
	// 	FloatValue: spellData.ImprovedStings.Effect(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_DOT)).FractionAt(hunter.Talents.ImprovedStings),
	// })
}

// TODO: To be implemented.
func (hunter *Hunter) registerMortalShots() {
	if hunter.Talents.MortalShots == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if hunter.Talents.MortalShots == 0 {
	// 	return
	// }
	//
	// hunter.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_CritMultiplier_Flat,
	// 	ProcMask:   core.ProcMaskRanged,
	// 	FloatValue: spellData.MortalShots.FractionAt(hunter.Talents.MortalShots),
	// })
}

// TODO: To be implemented.
func (hunter *Hunter) registerBarrage() {
	if hunter.Talents.Barrage == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if hunter.Talents.Barrage == 0 {
	// 	return
	// }
	//
	// hunter.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	ClassMask:  HunterSpellMultiShot | HunterSpellVolley,
	// 	FloatValue: spellData.Barrage.Effect(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_DAMAGE)).FractionAt(hunter.Talents.Barrage),
	// })
}

// TODO: To be implemented.
func (hunter *Hunter) registerRangedWeaponSpecialization() {
	if hunter.Talents.RangedWeaponSpecialization == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if hunter.Talents.RangedWeaponSpecialization == 0 {
	// 	return
	// }
	//
	// hunter.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Pct,
	// 	ProcMask:   core.ProcMaskRanged,
	// 	FloatValue: spellData.RangedWeaponSpecialization.FractionAt(hunter.Talents.RangedWeaponSpecialization),
	// })
}

// TODO: To be implemented.
func (hunter *Hunter) registerCarefulAim() {
	if hunter.Talents.CarefulAim == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if hunter.Talents.CarefulAim == 0 {
	// 	return
	// }
	//
	// hunter.AddStatDependency(stats.Intellect, stats.RangedAttackPower, 0.15*float64(hunter.Talents.CarefulAim))
}

// TODO: To be implemented.
func (hunter *Hunter) registerHawkEye() {
	if hunter.Talents.HawkEye == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if hunter.Talents.HawkEye == 0 {
	// 	return
	// }
	//
	// bonusRange := float64(hunter.Talents.HawkEye) * 2
	// ranged := hunter.AutoAttacks.Ranged()
	//
	// if ranged != nil {
	// 	ranged.MaxRange += bonusRange
	// }
	//
	// hunter.AddStaticMod(core.SpellModConfig{
	// 	Kind:     core.SpellMod_Custom,
	// 	ProcMask: core.ProcMaskRanged,
	// 	ApplyCustom: func(mod *core.SpellMod, spell *core.Spell) {
	// 		if spell.MaxRange > 0 {
	// 			spell.MaxRange += bonusRange
	// 		}
	// 	},
	// 	RemoveCustom: func(mod *core.SpellMod, spell *core.Spell) {
	// 		if spell.MaxRange > 0 {
	// 			spell.MaxRange -= bonusRange
	// 		}
	// 	},
	// })
}

// registerImprovedConcussiveShot implements Improved Concussive Shot, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerImprovedConcussiveShot() {
	if hunter.Talents.ImprovedConcussiveShot == 0 {
		return
	}
}

// registerLethalAttacks implements Lethal Attacks, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerLethalAttacks() {
	if hunter.Talents.LethalAttacks == 0 {
		return
	}
}

// registerLoneWolf implements Lone Wolf, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerLoneWolf() {
	if !hunter.Talents.LoneWolf {
		return
	}
}

// registerImprovedSerpentSting implements Improved Serpent Sting, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerImprovedSerpentSting() {
	if hunter.Talents.ImprovedSerpentSting == 0 {
		return
	}
}

// registerRapidRecuperation implements Rapid Recuperation, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerRapidRecuperation() {
	if hunter.Talents.RapidRecuperation == 0 {
		return
	}
}

// registerScatterShot implements Scatter Shot, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerScatterShot() {
	if !hunter.Talents.ScatterShot {
		return
	}
}

// registerSniperShot implements Sniper Shot, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (hunter *Hunter) registerSniperShot() {
	if !hunter.Talents.SniperShot {
		return
	}
}
