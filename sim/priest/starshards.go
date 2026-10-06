package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Starshards: Night Elf only. A 6 s Arcane channel on a 30 s cooldown.
var StarshardsRankMap = spellData.Starshards

func (priest *Priest) registerStarshardsSpell(rank *spelldata.Spell) {
	starShardTick := rank.PeriodicEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = PriestSpellStarshards

	config.Dot = spelldata.DotConfig(rank, starShardTick, spelldata.Label(rank.Name+" "+rank.Rank))
	config.Dot.OnTick = spelldata.PeriodicDamageTick(starShardTick, rank.TickOutcomeHitRolled)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}

	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
		return spell.CalcPeriodicDamage(sim, target, starShardTick.Average(), spell.OutcomeExpectedMagicAlwaysHit)
	}

	priest.RegisterSpell(config)
}
