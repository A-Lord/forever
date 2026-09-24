package warlock

func (warlock *Warlock) registerDestructionTalents() {
	// Tier 1
	warlock.applyDestructiveReach()
	warlock.applyImprovedShadowBolt()
	warlock.applyBane()

	// Tier 2
	warlock.applyMoltenSkin()
	warlock.applyCataclysm()
	warlock.applyAftermath()

	// Tier 3
	warlock.applyRuin()
	warlock.applyShadowburn()

	// Tier 4
	warlock.applyIntensity()
	warlock.applyAgonizingFlames()
	warlock.applyConflagrate()

	// Tier 5
	warlock.applyPyroclasm()
	warlock.applyBaneOfHavoc()
	warlock.applyFireAndBrimstone()

	// Tier 6
	warlock.applyShadowAndFlame()

	// Tier 7
	warlock.applyIncinerate()
}

// TODO: To be implemented. Improved Shadow Bolt (17793) is one dummy effect
// with no trigger spell and its debuff 17800 has no SpellName row, so there is
// nothing in the client to build the aura this talent applies from.
func (warlock *Warlock) applyImprovedShadowBolt() {
	if warlock.Talents.ImprovedShadowBolt == 0 {
		return
	}
}

// TODO: To be implemented. Port the TBC Cataclysm implementation below; not yet verified against the Forever client.
func (warlock *Warlock) applyCataclysm() {
	if warlock.Talents.Cataclysm == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.Cataclysm == 0 {
	// 	return
	// }
	//
	// warlock.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_PowerCost_Pct_Add,
	// 	FloatValue: -0.01 * float64(warlock.Talents.Cataclysm),
	// 	ClassMask:  WarlockDestructionSpells,
	// })
}

// TODO: To be implemented. Port the TBC Bane implementation below; not yet verified against the Forever client.
func (warlock *Warlock) applyBane() {
	if warlock.Talents.Bane == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.Bane == 0 {
	// 	return
	// }
	//
	// warlock.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_CastTime_Flat,
	// 	TimeValue: time.Millisecond * time.Duration(-100*warlock.Talents.Bane),
	// 	ClassMask: WarlockSpellShadowBolt | WarlockSpellImmolate,
	// })
	//
	// warlock.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_CastTime_Flat,
	// 	TimeValue: time.Millisecond * time.Duration(-400*warlock.Talents.Bane),
	// 	ClassMask: WarlockSpellSoulFire,
	// })
}

func (warlock *Warlock) applyShadowburn() {
	if !warlock.Talents.Shadowburn {
		return
	}

	warlock.registerShadowBurn()
}

// TODO: To be implemented. Port the TBC Destructive Reach implementation below; not yet verified against the Forever client.
func (warlock *Warlock) applyDestructiveReach() {
	if warlock.Talents.DestructiveReach == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.DestructiveReach == 0 {
	// 	return
	// }
	//
	// warlock.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_ThreatMultiplier_Pct,
	// 	FloatValue: -0.05 * float64(warlock.Talents.DestructiveReach),
	// 	ClassMask:  WarlockDestructionSpells,
	// })
	//
}

// TODO: To be implemented. Port the TBC Ruin implementation below; not yet verified against the Forever client.
func (warlock *Warlock) applyRuin() {
	if warlock.Talents.Ruin == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.Ruin == 0 {
	// 	return
	// }
	//
	// // TODO: Forever expands Ruin from 1 rank to 5; the per-rank crit-multiplier bonus is
	// // unconfirmed, so it is pinned to 0 until the Forever tooltip is known.
	// warlock.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_CritMultiplier_Flat,
	// 	FloatValue: 0,
	// 	ClassMask:  WarlockDestructionSpells,
	// })
}

func (warlock *Warlock) applyConflagrate() {
	if !warlock.Talents.Conflagrate {
		return
	}

	warlock.registerConflagrate()
}

// TODO: To be implemented. Port the TBC Shadow And Flame implementation below; not yet verified against the Forever client.
func (warlock *Warlock) applyShadowAndFlame() {
	if warlock.Talents.ShadowAndFlame == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.ShadowAndFlame == 0 {
	// 	return
	// }
	//
	// warlock.AddStaticMod(core.SpellModConfig{
	// 	Kind: core.SpellMod_BonusCoeffecient_Flat,
	// 	// Four dummy effects, of which only the first (4% per rank) matches the coefficient
	// 	// bonus this talent has always granted; the other three (20/2/2 per rank) are
	// 	// unidentified.
	// 	FloatValue: spellData.ShadowAndFlame.EffectAt(1).FractionAt(warlock.Talents.ShadowAndFlame),
	// 	ClassMask:  WarlockSpellShadowBolt | WarlockSpellIncinerate,
	// })
}

// applyMoltenSkin implements Molten Skin, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyMoltenSkin() {
	if warlock.Talents.MoltenSkin == 0 {
		return
	}
}

// applyAftermath implements Aftermath, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyAftermath() {
	if warlock.Talents.Aftermath == 0 {
		return
	}
}

// applyIntensity implements Intensity, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyIntensity() {
	if warlock.Talents.Intensity == 0 {
		return
	}
}

// applyAgonizingFlames implements Agonizing Flames, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyAgonizingFlames() {
	if warlock.Talents.AgonizingFlames == 0 {
		return
	}
}

// applyPyroclasm implements Pyroclasm, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyPyroclasm() {
	if warlock.Talents.Pyroclasm == 0 {
		return
	}
}

// applyBaneOfHavoc implements Bane of Havoc, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyBaneOfHavoc() {
	if !warlock.Talents.BaneOfHavoc {
		return
	}
}

// applyFireAndBrimstone implements Fire and Brimstone, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyFireAndBrimstone() {
	if warlock.Talents.FireAndBrimstone == 0 {
		return
	}
}

// applyIncinerate implements Incinerate, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyIncinerate() {
	if !warlock.Talents.Incinerate {
		return
	}
}
