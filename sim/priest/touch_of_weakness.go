package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerTouchOfWeakness(rank *spelldata.Spell) {
	// The buff's trigger is a dummy; the tooltip names the rank's strike, which deals the damage and
	// weakens the attacker.
	strikeRank := rank.Refs()[0]
	strikeDamage := strikeRank.DamageEffect()
	label := rank.Name + " " + rank.Rank

	weakened := priest.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		aura := target.RegisterAura(spelldata.AuraConfig(strikeRank, spelldata.Label(label)))
		spelldata.ParseEffects(nil, aura, strikeRank, spelldata.Level(priest.Level))
		return aura
	})

	strikeConfig := spelldata.SpellConfig(&priest.Unit, strikeRank,
		spelldata.Magic(core.ProcMaskSpellDamageProc), spelldata.Proc())
	strikeConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, strikeDamage.Roll(sim), spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			weakened.Get(target).Activate(sim)
		}
	}
	strikeConfig.RelatedAuraArrays = weakened.ToMap()
	strike := priest.RegisterSpell(strikeConfig)

	auraConfig := spelldata.AuraConfig(rank, spelldata.Label(label))
	auraConfig.Tag = rank.Name
	aura := priest.RegisterAura(auraConfig)
	aura.AttachProcTrigger(spelldata.ProcTrigger(&priest.Character, rank,
		func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			strike.Cast(sim, spell.Unit)
			aura.Deactivate(sim)
		}))

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		activateRank(sim, aura)
	}
	config.RelatedSelfBuff = aura

	priest.RegisterSpell(config)
}
