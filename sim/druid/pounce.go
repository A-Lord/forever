package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerPounceSpell() {
	rank := spellData.Pounce.Highest()
	bleedRank := spellData.PounceBleed.Highest()

	bleedConfig := spelldata.SpellConfig(&druid.Unit, bleedRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Proc())
	// TODO: Ingame research needed -- Genesis names Pounce 9827, while bleed 9826 has no class flags.
	// Its tooltip raises periodic damage, so only Genesis is applied to the bleed.
	if druid.Talents.Genesis > 0 {
		bleedConfig.DamageMultiplier *= 1 + spellData.Genesis.Rank(druid.Talents.Genesis).EffectN(1).Percent()
	}
	bleedConfig.Dot = spelldata.DotConfig(bleedRank, bleedRank.PeriodicDamageEffect(), bleedAuraOpt())
	druid.PounceBleed = druid.RegisterSpell(bleedConfig)

	stuns := druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(spelldata.AuraConfig(rank, spelldata.Label(rank.Name+" Stun")))
	})
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.RelatedAuraArrays = stuns.ToMap()
	// "Must be prowling and behind the target" is the tooltip's; the row states Cat Form only.
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return druid.ProwlAura != nil && druid.ProwlAura.IsActive() && druid.BehindTarget()
	}
	// No Active Defense (Attributes 0x200000): it misses or lands, never dodged, parried or blocked.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialNoBlockDodgeParryNoCrit)
		if result.Landed() {
			stuns.Get(target).Activate(sim)
			druid.PounceBleed.Dot(target).Apply(sim)
			druid.awardComboPoints(sim, spell, rank)
		}
		druid.refundOnMiss(sim, spell, result)
	}
	druid.Pounce = druid.RegisterSpell(config)
}
