package warlock

import (
	"github.com/wowsims/forever/sim/core"
)

var corruptionRank = spellData.Corruption.Highest()
var corruptionTick = corruptionRank.PeriodicEffect()
var corruptionCoeff = corruptionTick.Coeff()

// TODO: To be implemented. Port the TBC Corruption implementation below; not yet verified against the Forever client.
func (warlock *Warlock) registerCorruption() *core.Spell {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// tickLength := corruptionTick.Period()
	// tickCount := int32(corruptionRank.Duration() / tickLength)
	// warlock.CorruptionTickBaseDamage = corruptionTick.Average(core.CharacterLevel)
	//
	// warlock.Corruption = warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: corruptionRank.ID},
	// 	SpellSchool:    corruptionRank.SpellSchool(),
	// 	DefenseType:    corruptionRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: WarlockSpellCorruption,
	//
	// 	DamageMultiplier: 1,
	// 	ManaCost:         core.ManaCostOptions{FlatCost: int32(corruptionRank.Cost())},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD:      corruptionRank.GCD(),
	// 			CastTime: corruptionRank.CastTime(),
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
	//
	// 		if result.Landed() {
	// 			spell.Dot(target).Apply(sim)
	// 		}
	// 		spell.DealOutcome(sim, result)
	// 	},
	// 	BonusCoefficient: corruptionCoeff,
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label: "Corruption",
	// 			Tag:   "Affliction",
	// 		},
	// 		NumberOfTicks:    tickCount,
	// 		TickLength:       tickLength,
	// 		BonusCoefficient: corruptionCoeff,
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, warlock.CorruptionTickBaseDamage, dot.OutcomeTick)
	// 		},
	// 	},
	// 	ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
	// 		dot := spell.Dot(target)
	// 		result := spell.CalcPeriodicDamage(sim, target, warlock.CorruptionTickBaseDamage*float64(tickCount), spell.OutcomeExpectedMagicHit)
	// 		result.Damage /= dot.CalcTickPeriod().Round(time.Millisecond).Seconds()
	// 		return result
	// 	},
	// })
	//
	// return warlock.Corruption
}
