package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerPenance(rank *spelldata.Spell) {
	// The tooltip names the damage bolt first and the channel that times the bolts third. Each bolt is
	// its own spell, because Holy Precision, Holy Specialization and Inner Focus's crit name the bolt
	// row's class flags, not the cast's.
	refs := rank.Refs()
	boltRow := refs[0]
	bolt := boltRow.DamageEffect()
	channel := refs[2]

	boltConfig := spelldata.SpellConfig(&priest.Unit, boltRow, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc())
	boltConfig.ClassSpellMask = PriestSpellPenance
	boltConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		multiplier := priest.powerInLightMultiplier(target)
		spell.DamageMultiplier *= multiplier
		spell.CalcAndDealDamage(sim, target, bolt.Roll(sim), spell.OutcomeMagicHitAndCrit)
		spell.DamageMultiplier /= multiplier
	}
	boltSpell := priest.RegisterSpell(boltConfig)

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = PriestSpellPenance

	config.Dot = spelldata.DotConfig(channel, channel.PeriodicEffect(), spelldata.Label(rank.Name+" "+rank.Rank))
	config.Dot.Aura.ActionID = config.ActionID
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, _ *core.Dot) {
		boltSpell.Cast(sim, target)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		boltSpell.Cast(sim, target)
		spell.Dot(target).Apply(sim)
	}

	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, _ *core.Spell, _ bool) *core.SpellResult {
		return boltSpell.CalcDamage(sim, target, bolt.Average(), boltSpell.OutcomeExpectedMagicHitAndCrit)
	}

	priest.RegisterSpell(config)
}
