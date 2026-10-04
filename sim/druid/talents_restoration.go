package druid

import "github.com/wowsims/forever/sim/core"

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

func (druid *Druid) applyNaturalShapeshifter() {
	druid.parseTalent(druid.Talents.NaturalShapeshifter, spellData.NaturalShapeshifter)
}

func (druid *Druid) applyNaturalist() {
	druid.parseTalent(druid.Talents.Naturalist, spellData.Naturalist)
}

func (druid *Druid) applySubtlety() {
	if druid.Talents.Subtlety == 0 {
		return
	}

	effect := spellData.Subtlety.Rank(druid.Talents.Subtlety).EffectN(1)
	// A_MOD_THREAT's school mask limits both damage threat and flat threat to Nature and Arcane;
	// ParseStatic's global threat multiplier does not preserve that mask.
	for _, kind := range []core.SpellModType{core.SpellMod_ThreatMultiplier_Pct, core.SpellMod_FlatThreatBonus_Pct} {
		druid.AddStaticMod(core.SpellModConfig{
			School:     core.SpellSchool(effect.Misc),
			Kind:       kind,
			FloatValue: effect.Percent(),
		})
	}
}

func (druid *Druid) applyLivingSpirit() {
	druid.parseTalent(druid.Talents.LivingSpirit, spellData.LivingSpirit)
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

func (druid *Druid) applyFuror() {
	// Entering Cat Form and the bear forms reads the rank (forms.go). There is no aura to attach.
}

func (druid *Druid) applyReflection() {
	druid.parseTalent(druid.Talents.Reflection, spellData.Reflection)
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
