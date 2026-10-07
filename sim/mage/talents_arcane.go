package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
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

// Wand damage, +13% / +25%: the wand's shots are the mage's only ranged autos. The row's school mask
// (126) leaves out physical wands.
func (mage *Mage) registerWandSpecialization() {
	if mage.Talents.WandSpecialization == 0 {
		return
	}

	const schools = 126
	mage.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		ProcMask:   core.ProcMaskRangedAuto,
		School:     core.SpellSchool(schools),
		FloatValue: spellData.WandSpecialization.Effect(dbcenums.A_MOD_DAMAGE_PERCENT_DONE, schools).FractionAt(mage.Talents.WandSpecialization),
	})
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

	spelldata.ParseStatic(&mage.Character, spellData.ArcaneSubtlety.Rank(mage.Talents.ArcaneSubtlety))
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

	config := spelldata.AuraConfig(clearcastingRank)
	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if !spell.MatchesFlags(costEffect.ClassFlags) || spell.Cost == nil || spell.Cost.BaseCost == 0 {
			return
		}
		aura.Deactivate(sim)
	}
	mage.ClearCasting = mage.RegisterAura(config)
	// The row's -1000% cost is what lets Clearcasting beat Arcane Charges and Arcane Power, which
	// raise the cost in the same additive bucket.
	spelldata.ParseEffects(&mage.Character, mage.ClearCasting, clearcastingRank)

	trigger := spelldata.ProcTrigger(&mage.Character, spellData.ArcaneConcentration.Rank(mage.Talents.ArcaneConcentration),
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			mage.ClearCasting.Activate(sim)
		})
	trigger.Name = "Arcane Concentration - Trigger"
	trigger.ProcMask = core.ProcMaskSpellDamage
	// The Arcane Missiles channel lands an outcome with no damage; its missiles are the hits.
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return !spellData.ArcaneMissiles.Contains(spell)
	}

	mage.MakeProcTriggerAura(trigger)
}

// registerArcaneResilience: "Increases your Armor by an amount equal to 50% of your Intellect."
func (mage *Mage) registerArcaneResilience() {
	if mage.Talents.ArcaneResilience == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ArcaneResilience.Rank(mage.Talents.ArcaneResilience))
}

// registerArcaneGeometry: "Increases the range of your Arcane spells by 3/6 yards."
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

// registerArcaneBlastTalent implements the talent gate for Arcane Blast (arcane_blast.go).
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
// Both are applied where the spells are registered, in defensives.go and armors.go.
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
func (mage *Mage) registerArcaneMeditation() {
	if mage.Talents.ArcaneMeditation == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ArcaneMeditation.Rank(mage.Talents.ArcaneMeditation))
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

	config := spelldata.AuraConfig(buffRank)
	// The channel took the tick length and the cost when it was applied, before this runs.
	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.Matches(MageSpellArcaneMissilesCast) {
			aura.Deactivate(sim)
		}
	}
	buffAura := mage.RegisterAura(config)

	// Missiles every 0.5 sec (effect 2) and the mana cost (effect 3). The sim counts a DoT's
	// duration as its ticks times their length, so the halved tick alone halves the channel and
	// keeps its 5 missiles: the row's -50% duration (effect 1) is already in it and is not applied
	// again.
	spelldata.ParseEffects(&mage.Character, buffAura, buffRank, spelldata.Effects(2, 3))

	trigger := spelldata.ProcTrigger(&mage.Character, talentRank,
		func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			chance := arcaneBlastChance
			if !spellData.ArcaneBlast.Contains(spell) {
				chance /= 2
			}
			if sim.Proc(chance, "Missile Barrage") {
				buffAura.Activate(sim)
			}
		}, spelldata.Chance(1))
	trigger.Name = "Missile Barrage - Trigger"
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spellData.ArcaneBlast.Contains(spell) || spellData.Fireball.Contains(spell) ||
			spellData.Frostbolt.Contains(spell) || spellData.FrostfireBolt.Contains(spell)
	}

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

	// Effect 1's SPELLMOD_DAMAGE raises the direct hits and effect 2's SPELLMOD_DOT the ticks.
	spelldata.ParseStatic(&mage.Character, spellData.ArcaneInstability.Rank(mage.Talents.ArcaneInstability))
}
