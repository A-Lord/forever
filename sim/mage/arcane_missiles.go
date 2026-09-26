package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcaneMissilesRank = spellData.ArcaneMissiles.Highest()

// The channel's ticks each cast the same rank of the ArcaneMissilesTriggered family.
var arcaneMissilesTickRank = spellData.ArcaneMissilesTriggered.Highest()
var arcaneMissilesTickDamage = arcaneMissilesTickRank.DamageEffect()

func (mage *Mage) registerArcaneMissilesSpell() {
	tickConfig := spelldata.SpellConfig(&mage.Unit, arcaneMissilesTickRank,
		spelldata.Magic(core.ProcMaskSpellDamage))
	tickConfig.ActionID = core.ActionID{SpellID: arcaneMissilesRank.ID}.WithTag(1)
	tickConfig.ClassSpellMask = MageSpellArcaneMissilesTick
	// The missile row carries no speed of its own.
	tickConfig.MissileSpeed = 20

	tickConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, arcaneMissilesTickDamage.Average(core.CharacterLevel), spell.OutcomeTickMagicHitAndCrit)
		spell.SpellMetrics[result.Target.UnitIndex].Casts--
		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}
	arcaneMissilesTickSpell := mage.RegisterSpell(tickConfig)

	tickPeriod := arcaneMissilesRank.Effect(dbcenums.A_PERIODIC_TRIGGER_SPELL, 0).Period()

	config := spelldata.SpellConfig(&mage.Unit, arcaneMissilesRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellArcaneMissilesCast
	config.DamageMultiplier = 0
	config.BonusCoefficient = 0

	config.Dot = core.DotConfig{
		Aura: core.Aura{
			Label: "ArcaneMissiles",
		},
		NumberOfTicks:        int32(arcaneMissilesRank.Duration() / tickPeriod),
		TickLength:           tickPeriod,
		HasteReducesDuration: true,
		AffectedByCastSpeed:  true,
		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			arcaneMissilesTickSpell.Cast(sim, target)
		},
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
		// Casts is out here and not in Landed() to refer to the number of attempted casts.
		arcaneMissilesTickSpell.SpellMetrics[target.UnitIndex].Casts++
		if result.Landed() {
			spell.Dot(target).Apply(sim)
			arcaneMissilesTickSpell.SpellMetrics[target.UnitIndex].Hits++
		}
	}

	mage.RegisterSpell(config)
}
