package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Every rank is registered. The burns of all ranks share one rank group, so a cast of any rank ends
// the burn already down.
func (mage *Mage) registerFlamestrikeSpells() {
	spellData.Flamestrike.Each(mage.registerFlamestrikeSpell)
}

// The burn is its own spell (MageSpellFlamestrikeDot) on the cast's row, tagged like Fireball's DoT so
// the two share a header: every 2 sec it deals the same rank of spellData.FlamestrikeTriggered to each
// enemy in the area, with the Arcane Charges the cast spent. The cast's row names no tick spell, only
// an area trigger, so the two pair by rank.
func (mage *Mage) registerFlamestrikeSpell(rank int32, castRank *spelldata.Spell) {
	tickRank := spellData.FlamestrikeTriggered.Rank(rank)
	chargeLock := arcaneChargesLock{}
	tickDamage := tickRank.DamageEffect()

	dotConfig := spelldata.SpellConfig(&mage.Unit, castRank, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Tag(1))
	dotConfig.ClassSpellMask = MageSpellFlamestrikeDot
	dotConfig.Dot = spelldata.DotConfig(castRank, castRank.Effect(dbcenums.A_PERIODIC_DUMMY, 0),
		spelldata.RankLabel(castRank, &mage.Unit))
	dotConfig.Dot.IsAOE = true
	dotConfig.Dot.BonusCoefficient = tickDamage.Coeff()
	// An area DoT ticks on its caster, which is the unit the cast locks the charges on.
	dotConfig.Dot.OnTick = chargeLock.onTick(func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicAoeDamage(sim, tickDamage.Average(), dot.OutcomeTickMagicHit)
	})
	dotConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.AOEDot().Apply(sim)
	}
	dotSpell := mage.RegisterSpell(dotConfig)

	config := spelldata.SpellConfig(&mage.Unit, castRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellFlamestrike
	config.RelatedDotSpell = dotSpell

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, rollDamage(castRank.DamageEffect()))
		chargeLock.set(&mage.Unit, mage.arcaneChargesBonus())
		dotSpell.Cast(sim, target)
	}

	mage.RegisterSpell(config)
}
