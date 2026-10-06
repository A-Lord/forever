package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerPenance(rank *spelldata.Spell) {
	// The tooltip names the damage bolt first and the channel that times the bolts third.
	refs := rank.Refs()
	boltRow := refs[0]
	bolt := boltRow.DamageEffect()
	channel := refs[2]

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = PriestSpellPenance
	config.DefenseType = boltRow.DefenseTypeCore()
	config.BonusCoefficient = bolt.Coeff()

	dealBolt := func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		multiplier := priest.powerInLightMultiplier(target)
		spell.DamageMultiplier *= multiplier
		spell.CalcAndDealDamage(sim, target, bolt.Roll(sim), spell.OutcomeMagicHitAndCrit)
		spell.DamageMultiplier /= multiplier
	}

	config.Dot = spelldata.DotConfig(channel, channel.PeriodicEffect(), spelldata.Label(rank.Name+" "+rank.Rank))
	config.Dot.Aura.ActionID = config.ActionID
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dealBolt(sim, target, dot.Spell)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		dealBolt(sim, target, spell)
		spell.Dot(target).Apply(sim)
	}

	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, bolt.Average(), spell.OutcomeExpectedMagicHitAndCrit)
	}

	priest.RegisterSpell(config)
}
