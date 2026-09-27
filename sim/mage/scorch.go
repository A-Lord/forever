package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var scorchRank = spellData.Scorch.Highest()
var scorchDamage = scorchRank.DamageEffect()

func (mage *Mage) registerScorchSpell() {
	procChance := spellData.ImprovedScorch.EffectAt(1).FractionAt(mage.Talents.ImprovedScorch)

	if mage.Talents.ImprovedScorch > 0 {
		mage.ImprovedScorchAuras = mage.newFireVulnerabilityAuras()
	}

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

// Fire Vulnerability (22959), Improved Scorch's debuff: "Increases Fire damage taken from the Mage by
// 3% per stack", up to 5. It is A_MOD_SCHOOL_MASK_DAMAGE_FROM_CASTER with misc 4, so it scales only
// the Fire damage of the mage whose Scorch applied it, never another player's: each mage keeps its
// own stack on the target, and the bonus sits on this mage's attack table against it.
func (mage *Mage) newFireVulnerabilityAuras() core.AuraArray {
	debuff := spellData.ImprovedScorchTriggered.Highest()
	damageEffect := debuff.EffectN(1)
	school := core.SpellSchool(damageEffect.Misc)
	perStack := damageEffect.Percent()

	return mage.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		config := spelldata.AuraConfig(debuff, spelldata.Label(debuff.Name+mage.Label))

		aura := target.GetOrRegisterAura(config)
		aura.AttachDDBC(0, 1, &mage.AttackTables, func(_ *core.Simulation, spell *core.Spell, _ *core.AttackTable) float64 {
			if !spell.SpellSchool.Matches(school) {
				return 1
			}
			return 1 + perStack*float64(aura.GetStacks())
		})
		return aura
	})
}
