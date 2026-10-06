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
	bolt := boltRow.DamageEffect()
	channel := refs[2]

	boltConfig := spelldata.SpellConfig(&priest.Unit, boltRow, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc())
	boltConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		priest.calcAndDealPowerInLightDamage(sim, target, spell, bolt.Roll(sim))
	}
	boltSpell := priest.RegisterSpell(boltConfig)

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))

	config.Dot = spelldata.DotConfig(channel, channel.PeriodicEffect(), spelldata.RankLabel(rank))
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

	priest.RegisterSpell(config)
}
