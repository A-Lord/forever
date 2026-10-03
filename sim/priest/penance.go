package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerPenance(rank *spelldata.Spell, penanceTimer *core.Timer) {
	// panic("to be implemented")

	penanceTick := rank.PeriodicEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))

	config.ActionID = core.ActionID{SpellID: rank.ID}
	config.ClassSpellMask = PriestSpellSmite
	config.Flags = core.SpellFlagChanneled | core.SpellFlagAPL

	config.Cast.CD = core.Cooldown{
		Timer:    penanceTimer,
		Duration: max(rank.Cooldown(), rank.CategoryCooldown()),
	}

	config.Dot = spelldata.DotConfig(rank, penanceTick)
	config.Dot.Aura.Tag = "Penance"
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		damageMod := 1.0
		if target.HasAuraWithTag("Holy Fire") {
			damageMod *= 1 + 0.1*float64(priest.Talents.PowerInLight)
		}
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, penanceTick.Roll(sim)*damageMod, rank.TickOutcomeHitRolled(dot))
	}
	config.Dot.OnTick = spelldata.PeriodicDamageTick(penanceTick, rank.TickOutcomeHitRolled)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
	}

	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
		return spell.CalcPeriodicDamage(sim, target, penanceTick.Average(), spell.OutcomeExpectedMagicAlwaysHit)
	}

	priest.RegisterSpell(config)
}
