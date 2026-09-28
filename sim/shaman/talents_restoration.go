package shaman

func (shaman *Shaman) registerRestorationTalents() {
	// Tier 1
	shaman.applyImprovedHealingWave()
	shaman.applyTotemicFocus()

	// Tier 2
	shaman.applyMindfulness()
	shaman.applyNaturalGrace()
	// Tidal Focus not implemented
	shaman.applyImprovedReincarnation()

	// Tier 3
	shaman.applyAncestralHealing()
	shaman.applyHealingFocus()
	shaman.applyWaterShield()

	// Tier 4
	shaman.applyTidalMastery()
	shaman.applyRestorativeTotems()
	shaman.applyManaTideTotem()

	// Tier 5
	shaman.applyHealingWay()
	shaman.applyNaturesSwiftness()

	// Tier 6
	shaman.applyPurification()

	// Tier 7
	shaman.applyRiptide()
}

// TODO: To be implemented. Port the TBC Natures Swiftness implementation below; not yet verified against the Forever game files.
func (shaman *Shaman) applyNaturesSwiftness() {
	if !shaman.Talents.NaturesSwiftness {
		return
	}

	// The TBC implementation, kept for the port:
	// if !shaman.Talents.NaturesSwiftness {
	// 	return
	// }
	// nsAura := shaman.RegisterAura(core.Aura{
	// 	ActionID: core.ActionID{SpellID: 16188},
	// 	Label:    "Nature's Swiftness",
	// 	OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
	// 		if !spell.Matches(SpellMaskChainLightning | SpellMaskLightningBolt) {
	// 			return
	// 		}
	// 		aura.Deactivate(sim)
	// 	},
	// }).AttachSpellMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_CastTime_Pct,
	// 	FloatValue: -100,
	// 	ClassMask:  SpellMaskChainLightning | SpellMaskLightningBolt,
	// })
	//
	// shaman.RegisterSpell(core.SpellConfig{
	// 	ActionID:    core.ActionID{SpellID: 16188},
	// 	SpellSchool: core.SpellSchoolPhysical,
	// 	DefenseType: core.DefenseTypeMagic,
	// 	Flags:       core.SpellFlagAPL | core.SpellFlagNoOnCastComplete | SpellFlagInstant,
	// 	Cast: core.CastConfig{
	// 		CD: core.Cooldown{
	// 			Timer:    shaman.NewTimer(),
	// 			Duration: time.Second * 180,
	// 		},
	// 	},
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		nsAura.Activate(sim)
	// 	},
	// })
}

func (shaman *Shaman) applyRestorativeTotems() {
	if shaman.Talents.RestorativeTotems == 0 {
		return
	}
	// In totems.go
}

// TODO: To be implemented. Port the TBC Tidal Mastery implementation below; not yet verified against the Forever game files.
func (shaman *Shaman) applyTidalMastery() {
	if shaman.Talents.TidalMastery == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if shaman.Talents.TidalMastery == 0 {
	// 	return
	// }
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_BonusCrit_Percent,
	// 	FloatValue: spellData.TidalMastery.ValueAt(shaman.Talents.TidalMastery),
	// 	ClassMask:  SpellMaskChainLightning | SpellMaskLightningBolt | SpellMaskLightningShield | SpellMaskOverload,
	// })
}

// TODO: To be implemented. Port the TBC Totemic Focus implementation below; not yet verified against the Forever game files.
func (shaman *Shaman) applyTotemicFocus() {
	if shaman.Talents.TotemicFocus == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if shaman.Talents.TotemicFocus == 0 {
	// 	return
	// }
	// shaman.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_PowerCost_Pct_Add,
	// 	FloatValue: -0.05 * float64(shaman.Talents.TotemicFocus),
	// 	ClassMask:  SpellMaskTotem,
	// })
}

// applyImprovedHealingWave implements Improved Healing Wave, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyImprovedHealingWave() {
	if shaman.Talents.ImprovedHealingWave == 0 {
		return
	}
}

// applyMindfulness implements Mindfulness, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyMindfulness() {
	if shaman.Talents.Mindfulness == 0 {
		return
	}
}

// applyNaturalGrace implements Natural Grace, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyNaturalGrace() {
	if shaman.Talents.NaturalGrace == 0 {
		return
	}
}

// applyImprovedReincarnation implements Improved Reincarnation, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyImprovedReincarnation() {
	if shaman.Talents.ImprovedReincarnation == 0 {
		return
	}
}

// applyAncestralHealing implements Ancestral Healing, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyAncestralHealing() {
	if shaman.Talents.AncestralHealing == 0 {
		return
	}
}

// applyHealingFocus implements Healing Focus, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyHealingFocus() {
	if shaman.Talents.HealingFocus == 0 {
		return
	}
}

// applyWaterShield implements Water Shield, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyWaterShield() {
	if !shaman.Talents.WaterShield {
		return
	}
}

// applyManaTideTotem implements Mana Tide Totem, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyManaTideTotem() {
	if !shaman.Talents.ManaTideTotem {
		return
	}
}

// applyHealingWay implements Healing Way, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyHealingWay() {
	if shaman.Talents.HealingWay == 0 {
		return
	}
}

// applyPurification implements Purification, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyPurification() {
	if shaman.Talents.Purification == 0 {
		return
	}
}

// applyRiptide implements Riptide, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (shaman *Shaman) applyRiptide() {
	if !shaman.Talents.Riptide {
		return
	}
}
