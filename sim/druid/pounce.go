package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerPounceSpell() {
	rank := spellData.Pounce.Highest()
	bleedRank := spellData.PounceBleed.Highest()

	bleedConfig := spelldata.SpellConfig(&druid.Unit, bleedRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Proc())
	// Genesis names the parent Pounce in its periodic-damage mask. The triggered bleed row carries no
	// class flags of its own, so it inherits the parent's identity for that modifier.
	bleedConfig.ClassFlags = rank.ClassFlags
	bleedConfig.Dot = spelldata.DotConfig(bleedRank, bleedRank.PeriodicDamageEffect(), bleedAuraOpt())
	druid.PounceBleed = druid.RegisterSpell(bleedConfig)

	stuns := druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(spelldata.AuraConfig(rank, spelldata.Label(rank.Name+" Stun")))
	})
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	// "Must be prowling and behind the target" is the tooltip's; the row states Cat Form only.
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return druid.ProwlAura != nil && druid.ProwlAura.IsActive() && druid.BehindTarget()
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if result.Landed() {
			stuns.Get(target).Activate(sim)
			druid.PounceBleed.Dot(target).Apply(sim)
			druid.awardComboPoints(sim, spell, rank)
		}
		druid.refundOnMiss(sim, spell, result)
	}
	druid.Pounce = druid.RegisterSpell(config)
}
