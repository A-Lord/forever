package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var blastWaveRank = spellData.BlastWave.Highest()

// The daze (effect 2) does nothing to a boss.
func (mage *Mage) registerBlastWaveSpell() {
	config := spelldata.SpellConfig(&mage.Unit, blastWaveRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellBlastWave

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, rollDamage(blastWaveRank.DamageEffect()))
	}

	mage.RegisterSpell(config)
}
