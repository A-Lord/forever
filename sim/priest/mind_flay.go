package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerMindFlaySpell(rank *spelldata.Spell) {
	mindFlayTick := rank.PeriodicEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))

	config.ActionID = core.ActionID{SpellID: rank.ID}
	config.ClassSpellMask = PriestSpellMindFlay
	config.Flags = core.SpellFlagChanneled | core.SpellFlagAPL

	config.Dot = spelldata.DotConfig(rank, mindFlayTick)
	config.Dot.Aura.Tag = "Mind Flay"
	config.Dot.OnTick = spelldata.PeriodicDamageTick(mindFlayTick, rank.TickOutcomeHitRolled)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
	}

	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
		return spell.CalcPeriodicDamage(sim, target, mindFlayTick.Average(), spell.OutcomeExpectedMagicAlwaysHit)
	}

	priest.RegisterSpell(config)
}
