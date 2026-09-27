package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

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
	mage.registerColdSnapTalent()
	mage.registerFingersOfFrost()

	// Tier 6
	mage.registerWinterChill()

	// Tier 7
	mage.registerIceBarrier()
}

// registerColdSnapTalent implements the talent gate for Cold Snap (cold_snap.go).
func (mage *Mage) registerColdSnapTalent() {
	if !mage.Talents.ColdSnap {
		return
	}

	// TODO: call mage.registerColdSnapSpell() once it is implemented; it still panics.
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

// registerImprovedFrostbolt: Frostbolt's casting time is reduced by 0.1 sec per point.
func (mage *Mage) registerImprovedFrostbolt() {
	if mage.Talents.ImprovedFrostbolt == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ImprovedFrostbolt.Rank(mage.Talents.ImprovedFrostbolt))
}

// registerElementalPrecision: "Improves your chance to hit with Frost and Fire spells by 5%." The
// row's mask names the Fire and Frost spells, Frostfire Bolt included.
func (mage *Mage) registerElementalPrecision() {
	if mage.Talents.ElementalPrecision == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ElementalPrecision.Rank(mage.Talents.ElementalPrecision))
}

func (mage *Mage) registerIceShards() {
	if mage.Talents.IceShards == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.IceShards.Rank(mage.Talents.IceShards))
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

func (mage *Mage) registerImprovedFrostNova() {
	if mage.Talents.ImprovedFrostNova == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ImprovedFrostNova.Rank(mage.Talents.ImprovedFrostNova))
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

func (mage *Mage) registerPiercingIce() {
	if mage.Talents.PiercingIce == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.PiercingIce.Rank(mage.Talents.PiercingIce))
}

// registerFrostChanneling: "Reduces the mana cost of your Frost spells by 15% and reduces the
// threat caused by your Frost spells by 30%."
func (mage *Mage) registerFrostChanneling() {
	if mage.Talents.FrostChanneling == 0 {
		return
	}

	rank := spellData.FrostChanneling.Rank(mage.Talents.FrostChanneling)

	// The mana cost, on the row's mask of Frost spells.
	spelldata.ParseStatic(&mage.Character, rank, spelldata.Effects(1))

	// The threat, by hand: the parse reads A_MOD_THREAT as the unit's whole threat multiplier and
	// drops its misc 16, the Frost school. The talent curve also states the ranks as 10/20/30 where
	// the row itself states -30, so the amount is the curve's and the sign the row's and tooltip's.
	mage.AddStaticMod(core.SpellModConfig{
		School:     core.SpellSchoolFrost,
		FloatValue: -rank.EffectN(2).Percent(),
		Kind:       core.SpellMod_ThreatMultiplier_Pct,
	})
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

// registerArcticReach: "Increases the range of your Frostbolt and Blizzard spells and the radius of
// your Frost Nova and Cone of Cold spells by 20%."
func (mage *Mage) registerArcticReach() {
	if mage.Talents.ArcticReach == 0 {
		return
	}

	//Percentage is not handled in SpellMod so we can't use ParseStatic
	mage.AddStaticMod(core.SpellModConfig{
		ClassMask:  MageSpellFrostbolt | MageSpellBlizzard,
		FloatValue: 3 * float64(mage.Talents.ArcticReach),
		Kind:       core.SpellMod_Range_Flat,
	})
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

func (mage *Mage) registerImprovedConeOfCold() {
	if mage.Talents.ImprovedConeOfCold == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ImprovedConeOfCold.Rank(mage.Talents.ImprovedConeOfCold))
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

// registerWinterChill implements Winter's Chill: a Frost spell that lands has a 20% chance per point
// to put a stack of Winter's Chill (12579) on the target, 2% extra crit per stack for 15 sec, up to 5.
//
// The talent (11180) hears every harmful spell and names no class mask, so the Frost school is
// checked here; its chance is effect 2, which the resolver reads. Effect 1's base (1-5 by rank) is
// ignored: every proc adds one stack. The debuff is A_MOD_CRIT_CHANCE_FOR_CASTER_WITH_ABILITIES: it
// raises only this mage's crit, and only for the spells its class mask names (Frostbolt and Ice
// Lance), so it is applied by those spells rather than through the target's pseudo stats.
func (mage *Mage) registerWinterChill() {
	if mage.Talents.WintersChill == 0 {
		return
	}

	debuff := spellData.WintersChillTriggered.Highest()
	critEffect := debuff.EffectN(1)
	critPerStack := critEffect.BaseValue()

	config := spelldata.AuraConfig(debuff)
	wcAuras := mage.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(config)
	})

	// core has no per-target crit bonus scoped to a set of spells, so each affected spell adds the
	// target's stacks to its own crit for the duration of its effects.
	mage.Env.RegisterPreFinalizeEffect(func() {
		for _, spell := range mage.Spellbook {
			if !spell.MatchesFlags(critEffect.ClassFlags) {
				continue
			}

			spell.RelatedAuraArrays = wcAuras.ToMap()

			applyEffects := spell.ApplyEffects
			spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				bonus := critPerStack * float64(wcAuras.Get(target).GetStacks())
				spell.BonusCritPercent += bonus
				applyEffects(sim, target, spell)
				spell.BonusCritPercent -= bonus
			}
		}
	})

	trigger := spelldata.ProcTrigger(&mage.Character, spellData.WintersChill.Rank(mage.Talents.WintersChill),
		func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			aura := wcAuras.Get(result.Target)
			aura.Activate(sim)
			aura.AddStack(sim)
		})
	trigger.Name = "Winter's Chill - Trigger"
	// PROC_FLAG_DEAL_HARMFUL_SPELL decodes to the cast, which names no target; the debuff goes on
	// the one a Frost spell landed on.
	trigger.Callback = core.CallbackOnSpellHitDealt
	trigger.Outcome = core.OutcomeLanded
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spell.SpellSchool.Matches(core.SpellSchoolFrost)
	}

	mage.MakeProcTriggerAura(trigger)
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
