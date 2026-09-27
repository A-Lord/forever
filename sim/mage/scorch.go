package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var scorchRank = spellData.Scorch.Highest()
var scorchDamage = scorchRank.DamageEffect()

func (mage *Mage) registerScorchSpell() {

	procChance := spellData.ImprovedScorch.EffectAt(1).FractionAt(mage.Talents.ImprovedScorch)

	config := spelldata.SpellConfig(&mage.Unit, scorchRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellScorch

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, scorchDamage.Roll(sim, core.CharacterLevel), spell.OutcomeMagicHitAndCrit)
		if result.Landed() && sim.Proc(procChance, "Improved Scorch") {
			aura := mage.ImprovedScorchAuras.Get(target)
			aura.Activate(sim)
			aura.AddStack(sim)
		}
	}

	config.RelatedAuraArrays = mage.ImprovedScorchAuras.ToMap()

	mage.RegisterSpell(config)
}
