package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var iceBarrierRank = spellData.IceBarrier.Highest()
var iceBarrierAbsorb = iceBarrierRank.EffectN(1).Average(core.CharacterLevel)

// Ice Barrier
// https://www.wowhead.com/forever/spell=13033
//
// "Instantly shields you, absorbing 811 damage. Lasts 1 min. While the shield holds, your spellcasts
// will not be interrupted or delayed from taking damage." The row states no spell power coefficient,
// and the sim models no pushback, so only the absorb is.
func (mage *Mage) registerIceBarrierSpell() {
	shield := mage.NewDamageAbsorptionAura(core.AbsorptionAuraConfig{
		Aura: spelldata.AuraConfig(iceBarrierRank),
		ShieldStrengthCalculator: func(_ *core.Unit) float64 {
			return iceBarrierAbsorb
		},
	})

	config := spelldata.SpellConfig(&mage.Unit, iceBarrierRank, spelldata.Flags(core.SpellFlagAPL))
	config.ClassSpellMask = MageSpellIceBarrier
	config.RelatedSelfBuff = shield.Aura

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		shield.Activate(sim)
	}

	mage.RegisterSpell(config)
}
