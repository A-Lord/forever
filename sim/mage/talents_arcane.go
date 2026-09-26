package mage

import (
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
	// Arcane Blast implemented in arcane_blast.go and arcane_charge.go, gated on the
	// talent there.

	// Tier 4
	mage.registerArcaneShielding()
	mage.registerImprovedCounterspell()
	mage.registerArcaneMeditation()
	mage.registerMissileBarrage()

	// Tier 5
	// Presence of Mind implemented in presence_of_mind.go; registered unconditionally
	// from registerSpells (its own Talents.PresenceOfMind guard is inside that file).
	mage.registerArcaneMind()

	// Tier 6
	mage.registerArcaneInstability()

	// Tier 7
	// Arcane Power implemented in arcane_power.go; registered unconditionally
	// from registerSpells (its own Talents.ArcanePower guard is inside that file).
}

// Wand Specialization (6057) raises damage done with a wand equipped. The sim gives a mage
// no wand attack to raise, so the talent does nothing here; parsing its
// A_MOD_DAMAGE_PERCENT_DONE row statically would wrongly raise every spell school instead.
func (mage *Mage) registerWandSpecialization() {
	if mage.Talents.WandSpecialization == 0 {
		return
	}
}

func (mage *Mage) registerArcaneFocus() {
	if mage.Talents.ArcaneFocus == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ArcaneFocus.Rank(mage.Talents.ArcaneFocus))
}

// Improved Channeling (11237) reduces pushback while channeling, which the sim does not model.
func (mage *Mage) registerImprovedChanneling() {
	if mage.Talents.ImprovedChanneling == 0 {
		return
	}
}

func (mage *Mage) registerArcaneSubtlety() {
	if mage.Talents.ArcaneSubtlety == 0 {
		return
	}

	rank := spellData.ArcaneSubtlety.Rank(mage.Talents.ArcaneSubtlety)

	// Effect 2 is A_MOD_TARGET_RESISTANCE, spell penetration the parse table does not carry.
	spelldata.ParseStatic(&mage.Character, rank, spelldata.SkipEffects(2))
	mage.AddStat(stats.SpellPiercing, -rank.EffectN(2).BaseValue())
}

// Magic Absorption's first effect is a dummy restoring mana on a full resist, which the sim's
// targets never cause; only the resistances apply.
func (mage *Mage) registerMagicAbsorption() {
	if mage.Talents.MagicAbsorption == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.MagicAbsorption.Rank(mage.Talents.MagicAbsorption),
		spelldata.SkipEffects(1))
}

var clearcastingBuff = spellData.Clearcasting.Highest()

func (mage *Mage) registerArcaneConcentration() {
	if mage.Talents.ArcaneConcentration == 0 {
		return
	}

	// Clearcasting (12536): the row's one charge and its -100% cost modifier, masked to
	// the spells the client names.
	mage.ClearCasting = mage.RegisterAura(spelldata.AuraConfig(clearcastingBuff))
	spelldata.ParseEffects(&mage.Character, mage.ClearCasting, clearcastingBuff)

	mage.ClearCasting.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.ClassSpellMask&MageSpellsAllDamaging == 0 {
			return
		}
		if spell.DefaultCast.Cost == 0 {
			return
		}
		aura.Deactivate(sim)
	}

	trigger := spelldata.ProcTrigger(&mage.Character,
		spellData.ArcaneConcentration.Rank(mage.Talents.ArcaneConcentration),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			mage.ClearCasting.Activate(sim)
		})
	trigger.Name = "Arcane Concentration"
	// The row's proc flags hear every ability; the tooltip means damage spells.
	trigger.ClassSpellMask = MageSpellsAllDamaging
	mage.MakeProcTriggerAura(trigger)
}

// Arcane Resilience (28574) is A_MOD_ARMOR_PCT_FROM_STAT: armor raised by a share of
// intellect, which the parse table does not carry.
func (mage *Mage) registerArcaneResilience() {
	if mage.Talents.ArcaneResilience == 0 {
		return
	}

	mage.AddStatDependency(stats.Intellect, stats.Armor,
		spellData.ArcaneResilience.FractionAt(mage.Talents.ArcaneResilience))
}

func (mage *Mage) registerArcaneGeometry() {
	if mage.Talents.ArcaneGeometry == 0 {
		return
	}

	// A range modifier; the sim ignores range but the mod is the row's own.
	spelldata.ParseStatic(&mage.Character, spellData.ArcaneGeometry.Rank(mage.Talents.ArcaneGeometry))
}

func (mage *Mage) registerArcaneImpact() {
	if mage.Talents.ArcaneImpact == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ArcaneImpact.Rank(mage.Talents.ArcaneImpact))
}

func (mage *Mage) registerArcaneShielding() {
	if mage.Talents.ArcaneShielding == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ArcaneShielding.Rank(mage.Talents.ArcaneShielding))
}

// Improved Counterspell (11255) is a dummy the client scripts (silence on interrupt), which
// does nothing to a raid boss.
func (mage *Mage) registerImprovedCounterspell() {
	if mage.Talents.ImprovedCounterspell == 0 {
		return
	}
}

func (mage *Mage) registerArcaneMeditation() {
	if mage.Talents.ArcaneMeditation == 0 {
		return
	}

	// A_MOD_MANA_REGEN_INTERRUPT, which the parse table does not carry.
	mage.PseudoStats.SpiritRegenRateCasting += spellData.ArcaneMeditation.FractionAt(mage.Talents.ArcaneMeditation)
	mage.UpdateManaRegenRates()
}

var missileBarrageRank = spellData.MissileBarrage.Highest()
var missileBarrageBuff = spellData.MissileBarrageTriggered.Highest()

func (mage *Mage) registerMissileBarrage() {
	if !mage.Talents.MissileBarrage {
		return
	}

	// 400589 halves the next Arcane Missiles: the tick period drops by its effect 2
	// (-500ms, so the same five missiles fire twice as fast) and effect 3 empties the cost.
	// Effect 1's -50% duration is what the shorter channel adds up to, so only the period
	// and the cost are attached.
	tickTimeMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask: MageSpellArcaneMissilesCast,
		Kind:      core.SpellMod_DotTickLength_Flat,
		TimeValue: missileBarrageBuff.EffectN(2).TimeValue(),
	})
	costMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask:  MageSpellArcaneMissilesCast,
		Kind:       core.SpellMod_PowerCost_Pct_Add,
		FloatValue: missileBarrageBuff.EffectN(3).Percent(),
	})

	aura := mage.RegisterAura(spelldata.AuraConfig(missileBarrageBuff))
	aura.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
		tickTimeMod.Activate()
		costMod.Activate()
	})
	aura.ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
		tickTimeMod.Deactivate()
		costMod.Deactivate()
	})
	aura.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.ClassSpellMask == MageSpellArcaneMissilesCast {
			aura.Deactivate(sim)
		}
	}

	trigger := spelldata.ProcTrigger(&mage.Character, missileBarrageRank,
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			aura.Activate(sim)
		})
	trigger.Name = "Missile Barrage - Trigger"
	// The row's proc hint names one ability; the tooltip names Arcane Blast.
	trigger.ClassSpellMask = MageSpellArcaneBlast
	mage.MakeProcTriggerAura(trigger)
}

func (mage *Mage) registerArcaneMind() {
	if mage.Talents.ArcaneMind == 0 {
		return
	}

	// Effect 1 is +intellect% (A_MOD_TOTAL_STAT_PERCENTAGE, stat mask 8), effect 2 the
	// crit damage bonus on the spells the mask names.
	spelldata.ParseStatic(&mage.Character, spellData.ArcaneMind.Rank(mage.Talents.ArcaneMind))
}

func (mage *Mage) registerArcaneInstability() {
	if mage.Talents.ArcaneInstability == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ArcaneInstability.Rank(mage.Talents.ArcaneInstability))
}

// ------ FIRE TALENTS ------
