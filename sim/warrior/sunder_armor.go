package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var sunderArmorRank = spellData.SunderArmor.Highest()

func (warrior *Warrior) registerSunderArmor() {
	warrior.SunderArmorAuras = warrior.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataSunderArmorAura(target, true, 0)
	})

	config := spelldata.SpellConfig(&warrior.Unit, sunderArmorRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.FlatThreatBonus = sunderArmorRank.FindEffect(dbcenums.E_THREAT, 0, 0).Average(core.CharacterLevel)

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return warrior.CanApplySunderAura(target)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHit)

		if result.Landed() {
			aura := warrior.SunderArmorAuras.Get(target)
			aura.Activate(sim)
			aura.AddStack(sim)
		} else {
			spell.IssueRefund(sim)
		}

		spell.DealOutcome(sim, result)
	}

	config.RelatedAuraArrays = warrior.SunderArmorAuras.ToMap()

	warrior.RegisterSpell(config)
}

func (warrior *Warrior) CanApplySunderAura(target *core.Unit) bool {
	return warrior.SunderArmorAuras.Get(target).IsActive() || !warrior.SunderArmorAuras.Get(target).ExclusiveEffects[0].Category.AnyActive()
}
