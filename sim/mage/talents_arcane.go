package mage

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (mage *Mage) registerArcaneTalents() {
	// Tier 1
	mage.registerWandSpecialization()
	mage.registerArcaneFocus()
	mage.registerImprovedChanneling()

	// Tier 2
	mage.registerArcaneSubtlety()
	mage.registerMagicAbsorption()
	mage.registerArcaneConcentration()
	mage.registerArcaneResilience()

	// Tier 3
	mage.registerArcaneGeometry()
	mage.registerArcaneImpact()
	mage.registerArcaneBlastTalent()

	// Tier 4
	mage.registerArcaneShielding()
	mage.registerImprovedCounterspell()
	mage.registerArcaneMeditation()
	mage.registerMissileBarrage()

	// Tier 5
	mage.registerPresenceOfMindTalent()
	mage.registerArcaneMind()

	// Tier 6
	mage.registerArcaneInstability()

	// Tier 7
	mage.registerArcanePowerTalent()
}

// registerWandSpecialization: "Increases your damage with Wands by 25%." The sim has no wand
// attack for it to raise. ParseStatic would read the A_MOD_DAMAGE_PERCENT_DONE without its wand
// requirement and raise every spell, so the row is not parsed.
func (mage *Mage) registerWandSpecialization() {
	if mage.Talents.WandSpecialization == 0 {
		return
	}
}

// registerArcaneFocus: "Improves your chance to hit with Arcane spells by 5%."
func (mage *Mage) registerArcaneFocus() {
	if mage.Talents.ArcaneFocus == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ArcaneFocus.Rank(mage.Talents.ArcaneFocus))
}

// registerImprovedChanneling: a chance to avoid interruption from damage while channeling Arcane
// Missiles or casting Arcane Blast. The sim models no pushback, so there is nothing to apply.
func (mage *Mage) registerImprovedChanneling() {
	if mage.Talents.ImprovedChanneling == 0 {
		return
	}
}

// registerArcaneSubtlety: "Reduces your target's resistance to all your spells by 15 and reduces
// the threat caused by your Arcane spells by 30%."
func (mage *Mage) registerArcaneSubtlety() {
	if mage.Talents.ArcaneSubtlety == 0 {
		return
	}

	rank := spellData.ArcaneSubtlety.Rank(mage.Talents.ArcaneSubtlety)
	spelldata.ParseStatic(&mage.Character, rank, spelldata.Effects(1))

	// A_MOD_TARGET_RESISTANCE, which the parse does not know: spell penetration, stated negative.
	mage.AddStat(stats.SpellPiercing, -rank.EffectN(2).BaseValue())
}

// registerMagicAbsorption: "Increases all your resistances by 10 and causes all spells you fully
// resist to restore 2% of your total mana." The mana half needs spells cast at the mage, which the
// sim does not model; only the resistances are applied.
func (mage *Mage) registerMagicAbsorption() {
	if mage.Talents.MagicAbsorption == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.MagicAbsorption.Rank(mage.Talents.MagicAbsorption), spelldata.Effects(2))
}

// registerArcaneConcentration: "Gives you a 10% chance of entering a Clearcasting state after any
// damage spell hits a target. The Clearcasting state reduces the mana cost of your next damage
// spell by 100%." It procs on the hit and is spent by the next cast.
func (mage *Mage) registerArcaneConcentration() {
	if mage.Talents.ArcaneConcentration == 0 {
		return
	}

	clearcastingRank := spellData.Clearcasting.Highest()
	costEffect := clearcastingRank.EffectN(1)

	costMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassFlags: costEffect.ClassFlags,
		FloatValue: costEffect.Percent() / 10,
		Kind:       core.SpellMod_PowerCost_Pct_Add,
	})

	var proccedAt time.Duration
	var proccedBy *core.Spell

	config := spelldata.AuraConfig(clearcastingRank)
	config.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		costMod.Activate()
	}
	config.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		costMod.Deactivate()
	}
	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if !spell.MatchesFlags(costEffect.ClassFlags) || spell.Cost == nil || spell.Cost.BaseCost == 0 {
			return
		}
		// An instant that hits before its own OnCastComplete must not spend what it just procced.
		if proccedAt == sim.CurrentTime && proccedBy == spell {
			return
		}
		aura.Deactivate(sim)
	}
	mage.ClearCasting = mage.RegisterAura(config)

	trigger := spelldata.ProcTrigger(&mage.Character, spellData.ArcaneConcentration.Rank(mage.Talents.ArcaneConcentration),
		func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			proccedAt = sim.CurrentTime
			proccedBy = spell
			mage.ClearCasting.Activate(sim)
		})
	trigger.Name = "Arcane Concentration - Trigger"
	trigger.ClassSpellMask = MageSpellsAllDamaging

	mage.MakeProcTriggerAura(trigger)
}

// registerArcaneResilience: "Increases your Armor by an amount equal to 50% of your Intellect."
// A_MOD_ARMOR_PCT_FROM_STAT, which the parse does not know.
func (mage *Mage) registerArcaneResilience() {
	if mage.Talents.ArcaneResilience == 0 {
		return
	}

	mage.AddStatDependency(stats.Intellect, stats.Armor, spellData.ArcaneResilience.FractionAt(mage.Talents.ArcaneResilience))
}

// registerArcaneGeometry: "Increases the range of your Arcane spells by 6 yards."
func (mage *Mage) registerArcaneGeometry() {
	if mage.Talents.ArcaneGeometry == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ArcaneGeometry.Rank(mage.Talents.ArcaneGeometry))
}

// registerArcaneImpact: "Increases the critical strike chance of your Arcane spells by 6%."
func (mage *Mage) registerArcaneImpact() {
	if mage.Talents.ArcaneImpact == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ArcaneImpact.Rank(mage.Talents.ArcaneImpact))
}

// registerArcaneBlastTalent: Arcane Blast is a talent in Forever.
func (mage *Mage) registerArcaneBlastTalent() {
	if !mage.Talents.ArcaneBlast {
		return
	}

	mage.registerArcaneBlastSpell()
}

// registerPresenceOfMindTalent implements the talent gate for Presence of Mind (presence_of_mind.go).
func (mage *Mage) registerPresenceOfMindTalent() {
	if !mage.Talents.PresenceOfMind {
		return
	}

	mage.registerPresenceOfMindSpell()
}

// registerArcanePowerTalent implements the talent gate for Arcane Power (arcane_power.go).
func (mage *Mage) registerArcanePowerTalent() {
	if !mage.Talents.ArcanePower {
		return
	}

	mage.registerArcanePowerSpell()
}

// registerArcaneShielding: Mana Shield drains less mana and Mage Armor grants more resistance.
// Both are defensive, and neither spell is in the sim.
func (mage *Mage) registerArcaneShielding() {
	if mage.Talents.ArcaneShielding == 0 {
		return
	}
}

// registerImprovedCounterspell: "Your Counterspell also Silences the target for 4 sec." Nothing
// for a damage sim to apply.
func (mage *Mage) registerImprovedCounterspell() {
	if mage.Talents.ImprovedCounterspell == 0 {
		return
	}
}

// registerArcaneMeditation: "Allows 50% of your Mana regeneration to continue while casting."
// A_MOD_MANA_REGEN_INTERRUPT, which the parse does not know.
func (mage *Mage) registerArcaneMeditation() {
	if mage.Talents.ArcaneMeditation == 0 {
		return
	}

	mage.PseudoStats.SpiritRegenRateCasting += spellData.ArcaneMeditation.FractionAt(mage.Talents.ArcaneMeditation)
}

// registerMissileBarrage: "Gives your Arcane Blast spell a 40% chance, and your Fireball,
// Frostbolt, and Frostfire Bolt spells a 20% chance to reduce the channeled duration of your next
// Arcane Missiles spell by 50%, reduce the Mana cost by 100%, and missiles fire every 0.5 sec."
func (mage *Mage) registerMissileBarrage() {
	if !mage.Talents.MissileBarrage {
		return
	}

	talentRank := spellData.MissileBarrage.Highest()
	buffRank := spellData.MissileBarrageTriggered.Highest()
	arcaneBlastChance := talentRank.EffectN(1).Percent()

	// Missiles every 0.5 sec: SPELLMOD_ACTIVATION_TIME (effect 2), which the parse does not know.
	// The sim counts a DoT's duration as its ticks times their length, so the halved tick alone
	// halves the channel and keeps its 5 missiles: the row's -50% duration (effect 1) is already
	// in it and is not applied again.
	tickMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask: MageSpellArcaneMissilesCast,
		TimeValue: buffRank.EffectN(2).TimeValue(),
		Kind:      core.SpellMod_DotTickLength_Flat,
	})

	config := spelldata.AuraConfig(buffRank)
	config.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		tickMod.Activate()
	}
	config.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		tickMod.Deactivate()
	}
	// The channel took the tick length and the cost when it was applied, before this runs.
	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.Matches(MageSpellArcaneMissilesCast) {
			aura.Deactivate(sim)
		}
	}
	buffAura := mage.RegisterAura(config)

	// The mana cost (effect 3).
	spelldata.ParseEffects(&mage.Character, buffAura, buffRank, spelldata.Effects(3))

	trigger := spelldata.ProcTrigger(&mage.Character, talentRank,
		func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			chance := arcaneBlastChance
			if !spell.Matches(MageSpellArcaneBlast) {
				chance /= 2
			}
			if sim.Proc(chance, "Missile Barrage") {
				buffAura.Activate(sim)
			}
		}, spelldata.Chance(1))
	trigger.Name = "Missile Barrage - Trigger"
	trigger.ClassSpellMask = MageSpellArcaneBlast | MageSpellFireball | MageSpellFrostbolt | MageSpellFrostfireBolt

	mage.MakeProcTriggerAura(trigger)
}

// registerArcaneMind: "Increases your Intellect by 10% and increases the critical strike damage
// bonus of your Arcane spells by 100%."
func (mage *Mage) registerArcaneMind() {
	if mage.Talents.ArcaneMind == 0 {
		return
	}

	rank := spellData.ArcaneMind.Rank(mage.Talents.ArcaneMind)

	// Effect 1 names its stat in misc2 (8, Intellect's bit) and leaves misc at 0, which the parse
	// reads as Strength; the tooltip says Intellect.
	mage.MultiplyStat(stats.Intellect, 1+rank.EffectN(1).Percent())

	spelldata.ParseStatic(&mage.Character, rank, spelldata.Effects(2))
}

// registerArcaneInstability: "Increases the damage done by your spells by 3% and your critical
// strike chance by 3%."
func (mage *Mage) registerArcaneInstability() {
	if mage.Talents.ArcaneInstability == 0 {
		return
	}

	// Effect 2 is the SPELLMOD_DOT twin of effect 1's SPELLMOD_DAMAGE, on a mask effect 1 already
	// covers. The damage mod reaches the ticks as well as the hits, so the twin would count twice;
	// the parse folds such a twin only when both masks match, and these do not.
	spelldata.ParseStatic(&mage.Character, spellData.ArcaneInstability.Rank(mage.Talents.ArcaneInstability), spelldata.SkipEffects(2))
}
