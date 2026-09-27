package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcaneMissilesRank = spellData.ArcaneMissiles.Highest()
var triggered = spellData.ArcaneMissilesTriggered.Highest()
var arcaneMissilesBaseDamage = triggered.DamageEffect().Average(core.CharacterLevel)

func (mage *Mage) registerArcaneMissilesSpell() {
	tickConfig := spelldata.SpellConfig(&mage.Unit, triggered, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.RemoveFlags(core.SpellFlagAPL))
	tickConfig.ClassSpellMask = MageSpellArcaneMissilesTick
	var channel *core.Spell
	tickConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		// The Arcane Charges the channel locked in when it was cast, on every missile of it.
		chargeMultiplier := channel.Dot(target).PeriodicDamageMultiplier
		spell.DamageMultiplier *= chargeMultiplier
		result := spell.CalcAndDealDamage(sim, target, arcaneMissilesBaseDamage, spell.OutcomeTickMagicHitAndCrit)
		spell.DamageMultiplier /= chargeMultiplier
		spell.SpellMetrics[result.Target.UnitIndex].Casts--
	}

	tickSpell := mage.RegisterSpell(tickConfig)
	chargeBonuses := arcaneChargesDotBonuses{}

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

		// Locked in now: the charges are consumed by this cast's OnCastComplete, and the
		// missiles keep the bonus for the whole channel.
		dot := spell.Dot(target)
		chargeBonuses.lockIn(dot, mage.arcaneChargesDamageBonus())
		dot.Apply(sim)

		tickSpell.SpellMetrics[target.UnitIndex].Casts++
		tickSpell.SpellMetrics[target.UnitIndex].Hits++
	}

	channel = mage.RegisterSpell(spellConfig)
}
