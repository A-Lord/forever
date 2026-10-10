package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerPenance(_ int32, rank *spelldata.Spell) {
	// The tooltip names the damage bolt first and the channel that times the bolts third. Each bolt is
	// its own spell, because Holy Precision, Holy Specialization and Inner Focus's crit name the bolt
	// row's class flags, not the cast's.
	refs := rank.Refs()
	boltRow := refs[0]
	healBoltRow := refs[1]
	bolt := boltRow.DamageEffect()
	healBolt := healBoltRow.HealEffect()
	channel := refs[2]

	boltConfig := spelldata.SpellConfig(&priest.Unit, boltRow, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc())
	boltConfig.ClassSpellMask = PriestSpellPenance
	boltConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, bolt.Average(), spell.OutcomeMagicHitAndCrit)
	}
	boltSpell := priest.RegisterSpell(boltConfig)

	healBoltConfig := spelldata.SpellConfig(&priest.Unit, healBoltRow, spelldata.Magic(core.ProcMaskSpellHealing), spelldata.Proc())
	healBoltConfig.ClassSpellMask = PriestSpellPenance
	healBoltConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		target = priest.healTarget(target)
		priest.withRenewedHope(sim, target, spell, func() {
			spell.CalcAndDealHealing(sim, target, healBolt.Average(), spell.OutcomeHealingCrit)
		})
	}
	healBoltSpell := priest.RegisterSpell(healBoltConfig)

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))

	config.Dot = spelldata.DotConfig(channel, channel.PeriodicEffect(), spelldata.RankLabel(rank, &priest.Unit))
	config.Dot.Aura.ActionID = config.ActionID
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, _ *core.Dot) {
		boltSpell.Cast(sim, target)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		boltSpell.Cast(sim, target)
		spell.Dot(target).Apply(sim)
	}

	bolts := float64(1 + config.Dot.NumberOfTicks)
	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, _ *core.Spell, _ bool) *core.SpellResult {
		result := boltSpell.CalcDamage(sim, target, bolt.Average(), boltSpell.OutcomeExpectedMagicHitAndCrit)
		result.Damage *= bolts
		return result
	}

	healConfig := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellHealing))
	// Keep the heal bolt's action ID so existing APL entries still resolve to this cast.
	healConfig.ActionID = (core.ActionID{SpellID: healBoltRow.ID}).WithTag(1)
	healConfig.ClassSpellMask = PriestSpellPenance

	healConfig.Hot = spelldata.DotConfig(channel, channel.PeriodicEffect(), spelldata.RankLabel(rank, &priest.Unit))
	healConfig.Hot.Aura.ActionID = healConfig.ActionID
	healConfig.Hot.OnTick = func(sim *core.Simulation, target *core.Unit, _ *core.Dot) {
		healBoltSpell.Cast(sim, target)
	}

	healConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		target = priest.healTarget(target)
		priest.withRenewedHope(sim, target, spell, func() {
			healBoltSpell.Cast(sim, target)
			spell.Hot(target).Apply(sim)
		})
	}

	priest.RegisterSpell(config)
	priest.RegisterSpell(healConfig)
}
