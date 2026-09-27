package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var blizzardRank = spellData.Blizzard.Highest()
var blizzardTickRank = spellData.BlizzardTriggered.Highest()

// Each tick casts 1279949 on every enemy in the area; the ticks do not crit.
func (mage *Mage) registerBlizzardSpell() {
	chargeLock := arcaneChargesLock{}

	tickConfig := spelldata.SpellConfig(&mage.Unit, blizzardTickRank, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc())
	tickConfig.ClassSpellMask = MageSpellBlizzard
	tickConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		chargeLock.deal(spell, &mage.Unit, func() {
			spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHit, rollDamage(blizzardTickRank.DamageEffect()))
		})
	}
	tickSpell := mage.RegisterSpell(tickConfig)

	config := spelldata.SpellConfig(&mage.Unit, blizzardRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellBlizzard
	config.Dot = spelldata.DotConfig(blizzardRank, blizzardRank.Effect(dbcenums.A_PERIODIC_DUMMY, 0))
	config.Dot.IsAOE = true
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, _ *core.Dot) {
		tickSpell.Cast(sim, target)
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		chargeLock.set(&mage.Unit, mage.arcaneChargesBonus())
		spell.AOEDot().Apply(sim)
	}

	mage.RegisterSpell(config)
}
