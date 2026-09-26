package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var evocationRank = spellData.Evocation.Highest()

// The regen multiplier the row states: A_MOD_POWER_REGEN_PERCENT +1500%.
var evocationRegenBonus = evocationRank.EffectN(1).Percent()

func (mage *Mage) registerEvocation() {
	manaMetrics := mage.NewManaMetrics(core.ActionID{SpellID: evocationRank.ID})

	tickPeriod := evocationRank.Effect(dbcenums.A_PERIODIC_DUMMY, 0).Period()

	config := spelldata.SpellConfig(&mage.Unit, evocationRank)
	config.Flags |= core.SpellFlagHelpful | core.SpellFlagAPL
	config.ClassSpellMask = MageSpellEvocation

	config.Hot = core.DotConfig{
		SelfOnly: true,
		Aura: core.Aura{
			Label: "Evocation",
		},
		NumberOfTicks:        int32(evocationRank.Duration() / tickPeriod),
		TickLength:           tickPeriod,
		AffectedByCastSpeed:  true,
		HasteReducesDuration: true,

		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			// +1500% regen: each second grants fifteen extra seconds' worth of spirit
			// regen on top of the regen the sim already runs.
			mage.AddMana(sim, evocationRegenBonus*mage.SpiritManaRegenPerSecond(), manaMetrics)
		},
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.SelfHot().Apply(sim)
	}

	evocation := mage.RegisterSpell(config)

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: evocation,
		Type:  core.CooldownTypeMana,

		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			return character.CurrentManaPercent() <= .1
		},
	})
}
