package mage

func (mage *Mage) registerFrostTalents() {
	// Tier 1
	mage.registerFrostWarding()
	mage.registerImprovedFrostbolt()
	mage.registerElementalPrecision()

	// Tier 2
	mage.registerIceShards()
	mage.registerPermafrost()
	mage.registerImprovedFrostNova()
	mage.registerFrostbite()

	// Tier 3
	mage.registerPiercingIce()
	mage.registerFrostChanneling()
	mage.registerIceLance()
	mage.registerImprovedBlizzard()

	// Tier 4
	mage.registerArcticReach()
	mage.registerIceBlock()
	mage.registerShatter()

	// Tier 5
	mage.registerImprovedConeOfCold()
	// Cold Snap implemented in cold_snap.go; registered unconditionally
	// from registerSpells (its own Talents.ColdSnap guard is inside that file).
	mage.registerFingersOfFrost()

	// Tier 6
	mage.registerWinterChill()

	// Tier 7
	mage.registerIceBarrier()
}

// registerFrostWarding implements Frost Warding, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerFrostWarding() {
	if mage.Talents.FrostWarding == 0 {
		return
	}
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerImprovedFrostbolt() {
	if mage.Talents.ImprovedFrostbolt == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.ImprovedFrostbolt == 0 {
	// 	return
	// }
	//
	// mage.AddStaticMod(core.SpellModConfig{
	// 	ClassMask: MageSpellFrostbolt,
	// 	TimeValue: time.Millisecond * time.Duration(spellData.ImprovedFrostbolt.Effect(dbcenums.A_ADD_FLAT_MODIFIER, spelldata.SPELLMOD_CASTING_TIME).ValueAt(mage.Talents.ImprovedFrostbolt)),
	// 	Kind:      core.SpellMod_CastTime_Flat,
	// })
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerElementalPrecision() {
	if mage.Talents.ElementalPrecision == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.ElementalPrecision == 0 {
	// 	return
	// }
	// percent := spellData.ElementalPrecision.Effect(dbcenums.A_ADD_FLAT_MODIFIER, spelldata.SPELLMOD_RESIST_MISS_CHANCE).ValueAt(mage.Talents.ElementalPrecision)
	// mage.AddStaticMod(core.SpellModConfig{
	// 	School:     core.SpellSchoolFrostfire,
	// 	FloatValue: -percent / 100,
	// 	Kind:       core.SpellMod_PowerCost_Pct,
	// })
	//
	// // Bug: Gives 2% hit per point instead of 1% to frost spells.
	// // https://www.warcraftlogs.com/reports/kwd3V8MA9FgrRYhf/?boss=-3&difficulty=0&type=damage-done&source=1&target=2
	// mage.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexFrost] += (percent * 2)
	// mage.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexFire] += percent
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerIceShards() {
	if mage.Talents.IceShards == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.IceShards == 0 {
	// 	return
	// }
	//
	// mage.AddStaticMod(core.SpellModConfig{
	// 	School:     core.SpellSchoolFrost,
	// 	FloatValue: spellData.IceShards.FractionAt(mage.Talents.IceShards),
	// 	Kind:       core.SpellMod_CritMultiplier_Flat,
	// })
}

// registerPermafrost implements Permafrost, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerPermafrost() {
	if mage.Talents.Permafrost == 0 {
		return
	}
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerImprovedFrostNova() {
	if mage.Talents.ImprovedFrostNova == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.ImprovedFrostNova == 0 {
	// 	return
	// }
	//
	// mage.AddStaticMod(core.SpellModConfig{
	// 	ClassMask: MageSpellFrostNova,
	// 	TimeValue: time.Second * time.Duration(-2*mage.Talents.ImprovedFrostNova),
	// 	Kind:      core.SpellMod_CastTime_Flat,
	// })
}

// registerFrostbite implements Frostbite, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerFrostbite() {
	if mage.Talents.Frostbite == 0 {
		return
	}
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerPiercingIce() {
	if mage.Talents.PiercingIce == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.PiercingIce == 0 {
	// 	return
	// }
	//
	// mage.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  MageSpellFrost,
	// 	FloatValue: spellData.PiercingIce.Effect(dbcenums.A_ADD_PCT_MODIFIER, spelldata.SPELLMOD_DAMAGE).FractionAt(mage.Talents.PiercingIce),
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// })
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerFrostChanneling() {
	if mage.Talents.FrostChanneling == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.FrostChanneling == 0 {
	// 	return
	// }
	//
	// mage.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  MageSpellFrost,
	// 	FloatValue: -.05 * float64(mage.Talents.FrostChanneling),
	// 	Kind:       core.SpellMod_PowerCost_Pct_Add,
	// })
	//
	// threatMod := []float64{.04, .07, .1}
	// mage.AddStaticMod(core.SpellModConfig{
	// 	School:     core.SpellSchoolFrost,
	// 	FloatValue: -threatMod[mage.Talents.FrostChanneling-1],
	// 	Kind:       core.SpellMod_ThreatMultiplier_Pct,
	// })
}

// registerIceLance implements Ice Lance, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerIceLance() {
	if !mage.Talents.IceLance {
		return
	}
}

// registerImprovedBlizzard implements Improved Blizzard, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerImprovedBlizzard() {
	if mage.Talents.ImprovedBlizzard == 0 {
		return
	}
}

// registerArcticReach implements Arctic Reach, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerArcticReach() {
	if mage.Talents.ArcticReach == 0 {
		return
	}
}

// registerIceBlock implements Ice Block, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerIceBlock() {
	if !mage.Talents.IceBlock {
		return
	}
}

// registerShatter implements Shatter, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerShatter() {
	if mage.Talents.Shatter == 0 {
		return
	}
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerImprovedConeOfCold() {
	if mage.Talents.ImprovedConeOfCold == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.ImprovedConeOfCold == 0 {
	// 	return
	// }
	//
	// mage.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  MageSpellConeOfCold,
	// 	FloatValue: .15 + (.10 * (float64(mage.Talents.ImprovedConeOfCold) - 1)),
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// })
}

// registerFingersOfFrost implements Fingers of Frost, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerFingersOfFrost() {
	if mage.Talents.FingersOfFrost == 0 {
		return
	}
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerWinterChill() {
	if mage.Talents.WintersChill == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.WintersChill == 0 {
	// 	return
	// }
	//
	// // Forever states a flat SpellAuraOptions.ProcChance of 100 on the talent spell and puts the
	// // real per-rank chance on the effect, so ProcChanceAt would read 100% at every rank.
	// // Effect 1 is the stack count (1..5); effect 2 is the chance (20..100).
	// procChance := spellData.WintersChill.EffectAt(2).FractionAt(mage.Talents.WintersChill)
	//
	// wcAuras := mage.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
	// 	return core.WintersChillAura(target, 0)
	// })
	//
	// mage.Env.RegisterPreFinalizeEffect(func() {
	// 	for _, spell := range mage.GetSpellsMatchingSchool(core.SpellSchoolFrost) {
	// 		spell.RelatedAuraArrays.Append(wcAuras)
	// 	}
	// })
	//
	// mage.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:           "Winters Chill Talent",
	// 	Callback:       core.CallbackOnSpellHitDealt,
	// 	Outcome:        core.OutcomeLanded,
	// 	ClassSpellMask: MageSpellFrost,
	// 	ProcChance:     procChance,
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		aura := wcAuras.Get(result.Target)
	// 		aura.Activate(sim)
	// 		aura.AddStack(sim)
	// 	},
	// })
}

// registerIceBarrier implements Ice Barrier, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerIceBarrier() {
	if !mage.Talents.IceBarrier {
		return
	}
}
