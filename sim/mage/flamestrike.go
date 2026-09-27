package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Every rank is registered: burns of different ranks run side by side, so an AoE rotation stacks
// the top two. A recast of the same rank replaces its own burn.
func (mage *Mage) registerFlamestrikeSpells() {
	for rank := int32(1); rank <= spellData.Flamestrike.Len(); rank++ {
		mage.registerFlamestrikeSpell(spellData.Flamestrike.Rank(rank), spellData.FlamestrikeTriggered.Rank(rank))
	}
}

// The burn is its own spell (MageSpellFlamestrikeDot) on the cast's row, tagged like Fireball's DoT so
// the two share a header: every 2 sec it deals the rank's tick to each enemy in the area, with the
// Arcane Charges the cast spent.
func (mage *Mage) registerFlamestrikeSpell(castRank, tickRank *spelldata.Spell) {
	chargeLock := arcaneChargesLock{}
	tickDamage := tickRank.DamageEffect()

	dotConfig := spelldata.SpellConfig(&mage.Unit, castRank, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Tag(1))
	dotConfig.ClassSpellMask = MageSpellFlamestrikeDot
	dotConfig.Dot = spelldata.DotConfig(castRank, castRank.Effect(dbcenums.A_PERIODIC_DUMMY, 0),
		spelldata.Label(castRank.Name+mage.Label+" "+castRank.Rank))
	dotConfig.Dot.IsAOE = true
	dotConfig.Dot.BonusCoefficient = tickDamage.Coeff()
	dotConfig.Dot.OnTick = func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
		chargeLock.deal(dot.Spell, &mage.Unit, func() {
			dot.Spell.CalcAndDealPeriodicAoeDamage(sim, tickDamage.Roll(sim, core.CharacterLevel), dot.OutcomeTickMagicHit)
		})
	}
	dotConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.AOEDot().Apply(sim)
	}
	dotSpell := mage.RegisterSpell(dotConfig)

	config := spelldata.SpellConfig(&mage.Unit, castRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellFlamestrike
	config.RelatedDotSpell = dotSpell
	config.Rank = castRank.RankNumber()

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, rollDamage(castRank.DamageEffect()))
		chargeLock.set(&mage.Unit, mage.arcaneChargesBonus())
		dotSpell.Cast(sim, target)
	}

	mage.Flamestrike = append(mage.Flamestrike, mage.RegisterSpell(config))
}
