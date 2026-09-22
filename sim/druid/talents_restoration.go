package druid

func (druid *Druid) registerRestorationTalents() {
	// Tier 1
	druid.applyNaturesFocus()
	druid.applyFuror()

	// Tier 2
	druid.applyNaturalist()
	druid.applySubtlety()
	druid.applyNaturalShapeshifter()

	// Tier 3
	druid.applyReflection()
	druid.applyGiftOfNature()
	druid.applyGiftOfTheEarthmother()

	// Tier 4
	druid.applyTranquilSpirit()
	druid.applyImprovedRejuvenation()
	druid.applySwiftmend()

	// Tier 5
	druid.applyNaturesSwiftness()
	druid.applyLivingSpirit()
	druid.applyImprovedTranquility()

	// Tier 6
	druid.applyImprovedRegrowth()

	// Tier 7
	druid.applyWildGrowth()
}

// TODO: To be implemented.
func (druid *Druid) applyNaturalShapeshifter() {
	if druid.Talents.NaturalShapeshifter == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.NaturalShapeshifter == 0 {
	// 	return
	// }
	//
	// druid.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  DruidSpellCatForm | DruidSpellBearForm,
	// 	Kind:       core.SpellMod_PowerCost_Pct_Add,
	// 	FloatValue: -0.1 * float64(druid.Talents.NaturalShapeshifter),
	// })
}

// TODO: To be implemented.
func (druid *Druid) applyNaturalist() {
	if druid.Talents.Naturalist == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.Naturalist == 0 {
	// 	return
	// }
	//
	// // Forever states the damage bonus against every school (mask 127) instead of physical
	// // only; the sim keeps it on physical, which is all a feral druid deals.
	// druid.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= spellData.Naturalist.Effect(dbcenums.A_MOD_DAMAGE_PERCENT_DONE, 127).MultiplierAt(druid.Talents.Naturalist)
}

// TODO: To be implemented.
func (druid *Druid) applySubtlety() {
	if druid.Talents.Subtlety == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.Subtlety == 0 {
	// 	return
	// }
	//
	// // Reduces the threat caused by healing/damage spells by 4/8/12/16/20% per rank.
	// druid.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  DruidHealingSpells | DruidDamagingSpells,
	// 	Kind:       core.SpellMod_ThreatMultiplier_Pct,
	// 	FloatValue: -0.04 * float64(druid.Talents.Subtlety),
	// })
}

// TODO: To be implemented.
func (druid *Druid) applyLivingSpirit() {
	if druid.Talents.LivingSpirit == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.LivingSpirit == 0 {
	// 	return
	// }
	//
	// // Increases total Spirit by 5/10/15% per rank.
	// druid.MultiplyStat(stats.Spirit, spellData.LivingSpirit.MultiplierAt(druid.Talents.LivingSpirit))
}

// applyNaturesFocus implements Nature's Focus, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyNaturesFocus() {
	if druid.Talents.NaturesFocus == 0 {
		return
	}
}

// TODO: To be implemented.
func (druid *Druid) applyFuror() {
	if druid.Talents.Furor == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.Furor == 0 {
	// 	return
	// }
	//
	// // Both dummy effects carry the same ladder, one per form, so either answers the chance.
	// druid.FurorProcChance = spellData.Furor.EffectAt(1).FractionAt(druid.Talents.Furor)
}

// applyReflection implements Reflection, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyReflection() {
	if druid.Talents.Reflection == 0 {
		return
	}
}

// applyGiftOfNature implements Gift of Nature, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyGiftOfNature() {
	if druid.Talents.GiftOfNature == 0 {
		return
	}
}

// applyGiftOfTheEarthmother implements Gift of the Earthmother, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyGiftOfTheEarthmother() {
	if !druid.Talents.GiftOfTheEarthmother {
		return
	}
}

// applyTranquilSpirit implements Tranquil Spirit, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyTranquilSpirit() {
	if druid.Talents.TranquilSpirit == 0 {
		return
	}
}

// applyImprovedRejuvenation implements Improved Rejuvenation, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyImprovedRejuvenation() {
	if druid.Talents.ImprovedRejuvenation == 0 {
		return
	}
}

// applySwiftmend implements Swiftmend, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applySwiftmend() {
	if !druid.Talents.Swiftmend {
		return
	}
}

// applyNaturesSwiftness implements Nature's Swiftness, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyNaturesSwiftness() {
	if !druid.Talents.NaturesSwiftness {
		return
	}
}

// applyImprovedTranquility implements Improved Tranquility, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyImprovedTranquility() {
	if druid.Talents.ImprovedTranquility == 0 {
		return
	}
}

// applyImprovedRegrowth implements Improved Regrowth, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyImprovedRegrowth() {
	if druid.Talents.ImprovedRegrowth == 0 {
		return
	}
}

// applyWildGrowth implements Wild Growth, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyWildGrowth() {
	if !druid.Talents.WildGrowth {
		return
	}
}
