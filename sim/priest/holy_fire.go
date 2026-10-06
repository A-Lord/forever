package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var HolyFireRankMap = spellData.HolyFire

func (priest *Priest) registerHolyFireSpell(rank *spelldata.Spell) {
	holyFireHit := rank.DamageEffect()
	holyFireTick := rank.PeriodicEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))

	config.ActionID = core.ActionID{SpellID: rank.ID}
	config.ClassSpellMask = PriestSpellHolyFire

	config.Dot = spelldata.DotConfig(rank, holyFireTick, spelldata.Label(rank.Name+" "+rank.Rank))
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

	priest.RegisterSpell(config)
}
