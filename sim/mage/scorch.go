package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var scorchRank = spellData.Scorch.Highest()
var scorchDamage = scorchRank.DamageEffect()

var fireVulnerabilityRank = spellData.FireVulnerability.Highest()

// Fire Vulnerability's per-stack bonus, a percent on the row (3).
var fireVulnerabilityBonus = fireVulnerabilityRank.EffectN(1).Percent()

func (mage *Mage) registerScorchSpell() {
	// The client's Fire Vulnerability (22959) is A_MOD_SCHOOL_MASK_DAMAGE_FROM_CASTER, so it
	// raises the fire damage this mage deals rather than the damage the target takes, and
	// sim/core has no aura to build it from; each stack is worth 3% here.
	if mage.Talents.ImprovedScorch > 0 {
		mage.ImprovedScorchAuras = mage.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			aura := target.GetOrRegisterAura(spelldata.AuraConfig(fireVulnerabilityRank,
				spelldata.Label("Fire Vulnerability"+mage.Label)))
			aura.AttachDDBC(0, 1, &mage.AttackTables, func(_ *core.Simulation, spell *core.Spell, _ *core.AttackTable) float64 {
				if spell.SpellSchool.Matches(core.SpellSchoolFire) {
					return 1 + fireVulnerabilityBonus*float64(aura.GetStacks())
				}
				return 1
			})
			return aura
		})
	}

	// The chance is Improved Scorch's own ladder: A_ADD_TARGET_TRIGGER at 33/67/100.
	procChance := spellData.ImprovedScorch.FractionAt(mage.Talents.ImprovedScorch)

	config := spelldata.SpellConfig(&mage.Unit, scorchRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellScorch

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := scorchDamage.Roll(sim, core.CharacterLevel)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		if result.Landed() && mage.Talents.ImprovedScorch > 0 {
			if sim.Proc(procChance, "Improved Scorch") {
				aura := mage.ImprovedScorchAuras.Get(target)
				aura.Activate(sim)
				aura.AddStack(sim)
			}
		}
	}

	config.RelatedAuraArrays = mage.ImprovedScorchAuras.ToMap()

	mage.RegisterSpell(config)
}
