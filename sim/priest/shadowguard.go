package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerShadowguard(rank *spelldata.Spell) {
	// The buff's trigger is a dummy and its own effect states a larger amount (116 at rank 6) than the
	// tooltip, which prices the strike off the rank's triggered row (96). The tooltip's row is dealt.
	strikeRank := rank.Refs()[0]
	strikeDamage := strikeRank.DamageEffect()

	strikeConfig := spelldata.SpellConfig(&priest.Unit, strikeRank,
		spelldata.Magic(core.ProcMaskSpellDamageProc), spelldata.Proc())
	strikeConfig.ThreatMultiplier = 0 // The tooltip: this damage causes no threat.
	strikeConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, strikeDamage.Roll(sim), spell.OutcomeMagicHit)
	}
	strike := priest.RegisterSpell(strikeConfig)

	auraConfig := spelldata.AuraConfig(rank, spelldata.Label(rank.Name+" "+rank.Rank))
	auraConfig.Tag = rank.Name
	aura := priest.RegisterAura(auraConfig)
	aura.AttachProcTrigger(spelldata.ProcTrigger(&priest.Character, rank,
		func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			strike.Cast(sim, spell.Unit)
			aura.RemoveStack(sim)
		}))

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		activateRank(sim, aura)
		aura.SetStacks(sim, aura.MaxStacks)
	}
	config.RelatedSelfBuff = aura

	priest.RegisterSpell(config)
}
