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

	mage.registerColdSnapSpell()
}

// registerFrostWarding: "Increases the Armor and resistance given by your Frost Armor and Ice Armor
// spells by 30%. In addition, gives your Frost Ward a 20% chance to reflect Frost spells and effects
// while active."
//
// Not modelled: neither armor (armors.go is unported) nor Frost Ward is registered, and the reflect
// needs an enemy casting Frost at the mage.
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

// registerPermafrost: "Increases the duration of your Chill effects by 33% and reduces the target's
// speed by an additional 10%."
//
// Not modelled: both effects act on the slow alone, which does nothing to a boss the sim fights. The
// chance a Chill effect rolls for Frostbite and Fingers of Frost is per hit, not per second chilled.
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

// registerFrostbite: "Gives your Chill effects a 15% chance to Freeze the target for 5 sec."
//
// Not modelled: the freeze (12494) is a root, which a boss the sim fights is immune to, the same
// reason Frost Nova's root is not modelled. A target is only ever Frozen through Fingers of Frost.
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

// registerIceLance implements the talent gate for Ice Lance (ice_lance.go).
func (mage *Mage) registerIceLance() {
	if !mage.Talents.IceLance {
		return
	}

	mage.registerIceLanceSpell()
}

// registerImprovedBlizzard: "Adds a Chill effect to your Blizzard spell. This effect lowers the
// target's movement speed by 40% for 1.5 sec."
//
// The slow (12484) does nothing to a boss; what the talent changes is that Blizzard becomes a Chill
// effect, which isChill reads for Fingers of Frost. Blizzard itself is not registered yet.
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

// registerIceBlock implements the talent gate for Ice Block (ice_block.go).
func (mage *Mage) registerIceBlock() {
	if !mage.Talents.IceBlock {
		return
	}

	mage.registerIceBlockSpell()
}

// registerShatter: "Increases the critical strike chance of all your spells against Frozen targets
// by 50%." The amount is the talent's dummy effect, 17/33/50.
func (mage *Mage) registerShatter() {
	if mage.Talents.Shatter == 0 {
		return
	}

	critBonus := spellData.Shatter.ValueAt(mage.Talents.Shatter)

	mage.addTargetCritBonus(
		func(spell *core.Spell) bool {
			return spell.ProcMask.Matches(core.ProcMaskSpellDamage)
		},
		func(target *core.Unit) float64 {
			if mage.isFrozen(target) {
				return critBonus
			}
			return 0
		})
}

func (mage *Mage) registerImprovedConeOfCold() {
	if mage.Talents.ImprovedConeOfCold == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ImprovedConeOfCold.Rank(mage.Talents.ImprovedConeOfCold))
}

// registerFingersOfFrost: "Gives your Chill effects a 15% chance to grant you the Fingers of Frost
// effect, which treats your next 2 spells cast as if the target were Frozen. Lasts 15 sec."
//
// The chance is effect 2 (15 at both ranks), which the resolver reads; the charges are effect 1 (1
// and 2). The buff's own row (400669) names only Deep Freeze, which is not in Forever, so what
// "as if Frozen" reaches is the tooltip's: every spell reading isFrozen, Shatter and Ice Lance.
func (mage *Mage) registerFingersOfFrost() {
	if mage.Talents.FingersOfFrost == 0 {
		return
	}

	talentRank := spellData.FingersOfFrost.Rank(mage.Talents.FingersOfFrost)
	charges := int32(talentRank.EffectN(1).BaseValue())

	config := spelldata.AuraConfig(spellData.FingersOfFrostTriggered.Highest())
	config.MaxStacks = charges

	// A charge is spent by each damage spell cast. OnCastComplete fires after the cast's effects,
	// so the spell spending the last charge still reads the target as Frozen. The proc itself lands
	// a batch window after the hit, so it can never be spent by the cast that granted it.
	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.Flags.Matches(core.SpellFlagAPL) && spell.ProcMask.Matches(core.ProcMaskSpellDamage) {
			aura.RemoveStack(sim)
		}
	}

	mage.FingersOfFrostAura = mage.RegisterAura(config)

	trigger := spelldata.ProcTrigger(&mage.Character, talentRank,
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			mage.FingersOfFrostAura.Activate(sim)
			mage.FingersOfFrostAura.SetStacks(sim, charges)
		})
	trigger.Name = "Fingers of Frost - Trigger"
	// PROC_FLAG_DEAL_HARMFUL_SPELL decodes to the cast; a Chill effect is applied by the hit.
	trigger.Callback = core.CallbackOnSpellHitDealt
	trigger.Outcome = core.OutcomeLanded
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return mage.isChill(spell)
	}

	mage.MakeProcTriggerAura(trigger)
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

	mage.addTargetCritBonus(
		func(spell *core.Spell) bool {
			return spell.MatchesFlags(critEffect.ClassFlags)
		},
		func(target *core.Unit) float64 {
			return critPerStack * float64(wcAuras.Get(target).GetStacks())
		})

	mage.Env.RegisterPreFinalizeEffect(func() {
		for _, spell := range mage.Spellbook {
			if spell.MatchesFlags(critEffect.ClassFlags) {
				spell.RelatedAuraArrays = wcAuras.ToMap()
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

// registerIceBarrier implements the talent gate for Ice Barrier (ice_barrier.go).
func (mage *Mage) registerIceBarrier() {
	if !mage.Talents.IceBarrier {
		return
	}

	mage.registerIceBarrierSpell()
}

// The Chill effects Frostbite and Permafrost name (family 3, mask 0x100000): Frostbolt, Cone of Cold,
// Frostfire Bolt and the armors' Chilled.
var chillFlags = spellData.Frostbite.Highest().EffectN(1).ClassFlags

// Whether the spell applies a Chill effect: one the Chill mask names, or Blizzard with Improved
// Blizzard, which adds a Chill the mask does not carry.
func (mage *Mage) isChill(spell *core.Spell) bool {
	if spell.MatchesFlags(chillFlags) {
		return true
	}
	return mage.Talents.ImprovedBlizzard > 0 && spell.Matches(MageSpellBlizzard)
}

// Whether this mage's spells treat the target as Frozen. A boss the sim fights is immune to Frost
// Nova's and Frostbite's roots, so the only way there is Fingers of Frost, which does not look at
// the target.
func (mage *Mage) isFrozen(_ *core.Unit) bool {
	return mage.FingersOfFrostAura.IsActive()
}

// addTargetCritBonus raises the crit chance of the spells matches accepts by what bonus reads for
// the target. core has no per-target crit bonus scoped to a set of spells, so each matching spell
// carries the bonus for the duration of its own effects: the crit is rolled there, travel time or
// not. Spells are matched once every one is registered.
func (mage *Mage) addTargetCritBonus(matches func(*core.Spell) bool, bonus func(target *core.Unit) float64) {
	mage.Env.RegisterPreFinalizeEffect(func() {
		for _, spell := range mage.Spellbook {
			if spell.ApplyEffects == nil || !matches(spell) {
				continue
			}

			applyEffects := spell.ApplyEffects
			spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				critBonus := bonus(target)
				spell.BonusCritPercent += critBonus
				applyEffects(sim, target, spell)
				spell.BonusCritPercent -= critBonus
			}
		}
	})
}
