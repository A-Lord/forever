package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (priest *Priest) registerDarkSacrifice(rank *spelldata.Spell) {
	healthTick := rank.Effect(dbcenums.A_PERIODIC_DUMMY, 0)
	manaTick := rank.FirstAura(dbcenums.A_PERIODIC_ENERGIZE)
	manaMetrics := priest.NewManaMetrics(core.ActionID{SpellID: rank.ID})

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ClassSpellMask = PriestSpellDarkSacrifice

	config.Hot = spelldata.DotConfig(rank, manaTick, spelldata.Label(rank.Name+" "+rank.Rank))
	config.Hot.SelfOnly = true
	ticks := float64(config.Hot.NumberOfTicks)
	config.Hot.OnTick = func(sim *core.Simulation, _ *core.Unit, _ *core.Dot) {
		// The tooltip's ${$o2+$SPI}: the row's mana over the duration, and Spirit once, shared over the ticks.
		priest.RemoveHealth(sim, healthTick.Average())
		priest.AddMana(sim, manaTick.Average()+priest.GetStat(stats.Spirit)/ticks, manaMetrics)
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.SelfHot().Apply(sim)
	}

	darkSacrifice := priest.RegisterSpell(config)
	if rank.ID != spellData.DarkSacrifice.Highest().ID {
		return
	}

	priest.AddMajorCooldown(core.MajorCooldown{
		Spell: darkSacrifice,
		Type:  core.CooldownTypeMana,
		ShouldActivate: func(_ *core.Simulation, character *core.Character) bool {
			return character.MaxMana()-character.CurrentMana() >= manaTick.Average()*ticks+character.GetStat(stats.Spirit)
		},
	})
}
