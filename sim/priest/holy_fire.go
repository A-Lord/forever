package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var HolyFireRankMap = spellData.HolyFire

func (priest *Priest) registerHolyFireSpell(_ int32, rank *spelldata.Spell) {
	holyFireHit := rank.DamageEffect()
	// TODO: Verify rank 5 (15265) in game. Its tick states 13 where the ranks around it state 7 and 10
	// (4/5/6/7/13/10/13/15 by rank, the same 0.05 spell power share on each).
	holyFireTick := rank.PeriodicEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = PriestSpellHolyFire

	config.Dot = spelldata.DotConfig(rank, holyFireTick, spelldata.RankLabel(rank))
	config.Dot.OnTick = spelldata.PeriodicDamageTick(holyFireTick, rank.TickOutcomeHitRolled)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, holyFireHit.Roll(sim), spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
	}

	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, holyFireHit.Average(), spell.OutcomeExpectedMagicHitAndCrit)
	}

	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
		return spell.CalcPeriodicDamage(sim, target, holyFireTick.Average(), spell.OutcomeExpectedMagicAlwaysHit)
	}

	priest.holyFires = append(priest.holyFires, priest.RegisterSpell(config))
}
