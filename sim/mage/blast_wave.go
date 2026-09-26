package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var blastWaveRank = spellData.BlastWave.Highest()
var blastWaveDamage = blastWaveRank.DamageEffect()

func (mage *Mage) registerBlastWaveSpell() {
	if !mage.Talents.BlastWave {
		return
	}

	config := spelldata.SpellConfig(&mage.Unit, blastWaveRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellBlastWave

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		baseDamage := blastWaveDamage.Roll(sim, core.CharacterLevel)
		spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
		// The row's second effect dazes the targets hit, which does nothing to a raid boss.
	}

	mage.RegisterSpell(config)
}
