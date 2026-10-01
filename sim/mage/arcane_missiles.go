package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcaneMissilesRank = spellData.ArcaneMissiles.Highest()
var triggered = spellData.ArcaneMissilesTriggered.Highest()
var arcaneMissilesDamage = triggered.DamageEffect()

func (mage *Mage) registerArcaneMissilesSpell() {
	tickConfig := spelldata.SpellConfig(&mage.Unit, triggered, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.RemoveFlags(core.SpellFlagAPL))
	tickConfig.ClassSpellMask = MageSpellArcaneMissilesTick
	// Every missile of a channel carries the Arcane Charges its cast spent.
	chargeLock := arcaneChargesLock{}
	tickConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		chargeLock.deal(spell, target, func() {
			spell.CalcAndDealDamage(sim, target, arcaneMissilesDamage.Roll(sim), spell.OutcomeTickMagicHitAndCrit)
		})
		spell.SpellMetrics[target.UnitIndex].Casts--
	}

	tickSpell := mage.RegisterSpell(tickConfig)

	spellConfig := spelldata.SpellConfig(&mage.Unit, arcaneMissilesRank, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.RemoveFlags(core.SpellFlagHelpful))
	spellConfig.ClassSpellMask = MageSpellArcaneMissilesCast

	dotConfig := spelldata.DotConfig(arcaneMissilesRank, arcaneMissilesRank.PeriodicEffect())
	dotConfig.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		tickSpell.Cast(sim, target)
	}

	spellConfig.Dot = dotConfig

	spellConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		// The channel never misses; each missile rolls its own hit.
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)

		// Locked in now, before this cast's OnCastComplete spends the charges.
		chargeLock.set(target, mage.arcaneChargesBonus())
		spell.Dot(target).Apply(sim)

		tickSpell.SpellMetrics[target.UnitIndex].Casts++
		tickSpell.SpellMetrics[target.UnitIndex].Hits++
	}

	mage.RegisterSpell(spellConfig)
}
