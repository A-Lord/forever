package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warlock *Warlock) registerAfflictionTalents() {
	// Tier 1
	// Improved Life Tap implemented in lifetap.go
	warlock.applySuppression()
	warlock.applyImprovedCorruption()

	// Tier 2
	warlock.applyMalediction()
	warlock.applySoulHarvesting()
	warlock.applyImprovedDrains()

	// Tier 3
	warlock.applyImprovedBaneOfAgony()
	warlock.applyFelConcentration()
	warlock.registerAmplifyCurse()
	warlock.applyPandemic()

	// Tier 4
	warlock.applyMalevolence()
	warlock.applyNightfall()
	warlock.registerCurseOfExhaustion()

	// Tier 5
	warlock.applySiphonLife()
	// Soul Siphon implemented via helper functions that calculate the multiplier, and are called by the affected spells.

	// Tier 6
	warlock.applyShadowMastery()

	// Tier 7
	warlock.applyWrack()
}

// --- Tier 1 ---

func (warlock *Warlock) applySuppression() {
	if warlock.Talents.Suppression == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.Suppression.Rank(warlock.Talents.Suppression))
}

func (warlock *Warlock) applyImprovedCorruption() {
	if warlock.Talents.ImprovedCorruption == 0 {
		return
	}
	spelldata.ParseStatic(&warlock.Character, spellData.ImprovedCorruption.Rank(warlock.Talents.ImprovedCorruption))
}

// --- Tier 2 ---

func (warlock *Warlock) applyMalediction() {
	if warlock.Talents.Malediction == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.Malediction.Rank(warlock.Talents.Malediction))
}

// TODO: Not implemented due to lack of a kill event. Currently adds "leave" the encounter
//
//	on a set timer, rather than "dying" to trigger the effect.
//
// Requires a drain soul kill to trigger.
func (warlock *Warlock) applySoulHarvesting() {
	if warlock.Talents.SoulHarvesting == 0 {
		return
	}
}

func (warlock *Warlock) applyImprovedDrains() {
	if warlock.Talents.ImprovedDrains == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.ImprovedDrains.Rank(warlock.Talents.ImprovedDrains))
}

// --- Tier 3 ---
func (warlock *Warlock) applyImprovedBaneOfAgony() {
	if warlock.Talents.ImprovedBaneOfAgony == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.ImprovedBaneOfAgony.Rank(warlock.Talents.ImprovedBaneOfAgony))
}

// TODO: Not implemented, as it does not affect sims for now.
// The row is SPELLMOD_NOT_LOSE_CASTING_TIME (pushback protection) on the drains and Wrack,
// which the parse skips; it matters only once the sim models pushback from damage taken.
func (warlock *Warlock) applyFelConcentration() {
	if warlock.Talents.FelConcentration == 0 {
		return
	}
}

func (warlock *Warlock) registerAmplifyCurse() {
	if !warlock.Talents.AmplifyCurse {
		return
	}

	rankedSpell := spellData.AmplifyCurse.ByID(18288)
	warlock.AmplifyCurseMultiplier = 1 + rankedSpell.EffectN(1).Percent()

	// Both effects name the curses they amplify, which are the ones that spend it.
	// Effect 1 is Bane of Agony and Curse of Weakness
	// Effect 2 is Curse of the Elements.
	amplifiedCurses, amplifiedElements := rankedSpell.EffectN(1).ClassFlags, rankedSpell.EffectN(2).ClassFlags

	// TODO: Confirm that the amp is consumed on cast, regardless of hit.
	auraConfig := spelldata.AuraConfig(rankedSpell)
	auraConfig.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.MatchesFlags(amplifiedCurses) || spell.MatchesFlags(amplifiedElements) {
			aura.Deactivate(sim)
		}
	}
	warlock.AmplifyCurseAura = warlock.RegisterAura(auraConfig)

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Flags(core.SpellFlagAPL))
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		warlock.AmplifyCurseAura.Activate(sim)
	}
	warlock.AmplifyCurse = warlock.RegisterSpell(config)

	warlock.AddMajorCooldown(core.MajorCooldown{
		Spell: warlock.AmplifyCurse,
		Type:  core.CooldownTypeDPS,
	})
}

func (warlock *Warlock) applyPandemic() {
	if warlock.Talents.Pandemic == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.Pandemic.Rank(warlock.Talents.Pandemic))
}

// --- Tier 4 ---
func (warlock *Warlock) applyMalevolence() {
	if warlock.Talents.Malevolence == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.Malevolence.Rank(warlock.Talents.Malevolence))
}

func (warlock *Warlock) applyNightfall() {
	if warlock.Talents.Nightfall == 0 {
		return
	}

	shadowTrance := spellData.NightfallTriggered.ByID(17941)
	// Effect 1 names the spells it makes instant, which are the ones that spend it.
	instant := shadowTrance.EffectN(1).ClassFlags

	auraConfig := spelldata.AuraConfig(shadowTrance)
	auraConfig.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.MatchesFlags(instant) && spell.CurCast.CastTime == 0 {
			aura.Deactivate(sim)
		}
	}

	warlock.NightfallProcAura = warlock.RegisterAura(auraConfig)

	spelldata.ParseEffects(&warlock.Character, warlock.NightfallProcAura, shadowTrance)

	trigger := spelldata.ProcTrigger(&warlock.Character, spellData.Nightfall.Rank(warlock.Talents.Nightfall),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			warlock.NightfallProcAura.Activate(sim)
		})
	// The talent row names no spells, so the triggering ones are gathered from their own rows.
	trigger.ClassFlags = combineClassFlags(
		spellData.Corruption.Highest().ClassFlags,
		spellData.DrainSoul.Highest().ClassFlags,
		spellData.DrainLife.Highest().ClassFlags,
		spellData.Wrack.Highest().ClassFlags,
	)

	trigger.RequireDamageDealt = true
	warlock.MakeProcTriggerAura(trigger)
}

// TODO: Not implemented, as it does not affect DPS: the curse only slows the target's movement.
func (warlock *Warlock) registerCurseOfExhaustion() {
	if !warlock.Talents.CurseOfExhaustion {
		return
	}
}

// --- Tier 5 ---

func (warlock *Warlock) applySiphonLife() {
	if !warlock.Talents.SiphonLife {
		return
	}

	warlock.registerSiphonLifeSpell()
}

func (warlock *Warlock) soulSiphonPerEffect() float64 {
	return spellData.SoulSiphon.FractionAt(warlock.Talents.SoulSiphon)
}

// Soul Siphon implementation, counts affliction effects to be used as a mutliplier in other spells.
// Caps at a maximum of 3 effects (12% per, max 36%)
// TODO: Appears to count any "affliction effect", likely includes curses in addition to standard affliction dots.
func (warlock *Warlock) soulSiphonMultiplier(target *core.Unit, multiplierPerEffect float64) float64 {
	if multiplierPerEffect == 0 {
		return 1
	}
	return 1 + multiplierPerEffect*min(warlock.AfflictionCount(target), 3)
}

// --- Tier 6 ---
func (warlock *Warlock) applyShadowMastery() {
	if warlock.Talents.ShadowMastery == 0 {
		return
	}

	rank := spellData.ShadowMastery.Rank(warlock.Talents.ShadowMastery)

	// Effect 1 applies to direct damage dealt.
	// Spell flags include corruption, but should not affect sims as it has no direct damage component.
	hit := rank.EffectN(1)
	warlock.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DirectDamageDone_Flat,
		ClassFlags: hit.ClassFlags,
		FloatValue: hit.Percent(),
	})

	// Effect 2 modifies periodic damage dealt.
	spelldata.ParseStatic(&warlock.Character, rank, spelldata.Effects(2))
}

// --- Tier 7 ---
func (warlock *Warlock) applyWrack() {
	if !warlock.Talents.Wrack {
		return
	}

	warlock.registerWrackSpell()
}
