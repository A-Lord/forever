package warlock

// TODO: To be implemented. spellData.DrainLife holds the six trainer ranks, 689 to 11700.
// The client also carries 403677 to 403689, the copies the Season of Discovery rune passive
// Master Channeler (403668) swaps onto the action bar. That is an Engrave grant with no place
// in Forever, and the generator drops its stand-ins.
//
// Imports and rank pins the implementation needs, kept with it:
// import (
// 	"math"
//
// 	"github.com/wowsims/forever/sim/core"
// 	"github.com/wowsims/forever/sim/core/dbcenums"
// )
//
// var drainLifeRank = spellData.DrainLife.Highest()
// var drainLifeTick = drainLifeRank.PeriodicEffect()
// var drainLifeCoeff = drainLifeTick.Coeff()

func (warlock *Warlock) registerDrainLife() {
	panic("To be implemented")

	// tickLength := drainLifeTick.Period()
	// healthMetric := warlock.NewHealthMetrics(core.ActionID{SpellID: drainLifeRank.ID})
	// resultSlice := make(core.SpellResultSlice, 1)
	//
	// // TODO: cappedDmgBonus/the == 2 check are TBC's 2-rank Soul Siphon (2%/4% per Affliction effect,
	// // capped at 1.24/1.60 total). spellData.SoulSiphon now ships 3 ranks at 4/8/12% with no cap effect
	// // in the table, so rank 3 currently falls through to the rank-2 cap of 1.60 rather than a sourced
	// // rank-3 number - neither the client nor spellData states what the rank-3 cap should be.
	// cappedDmgBonus := 1.24
	// if warlock.Talents.SoulSiphon == 2 {
	// 	cappedDmgBonus = 1.60
	// }
	//
	// // Read once rather than per tick: the ladder lookup scans the table and boxes the rank.
	// soulSiphonPerAffliction := spellData.SoulSiphon.Effect(dbcenums.A_DUMMY, 0).FractionAt(warlock.Talents.SoulSiphon)
	//
	// warlock.DrainLife = warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: drainLifeRank.ID},
	// 	SpellSchool:    drainLifeRank.SpellSchool(),
	// 	DefenseType:    drainLifeRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagChanneled | core.SpellFlagAPL,
	// 	ClassSpellMask: WarlockSpellDrainLife,
	//
	// 	ManaCost: core.ManaCostOptions{FlatCost: drainLifeRank.Cost()},
	// 	Cast:     core.CastConfig{DefaultCast: core.Cast{GCD: drainLifeRank.GCD()}},
	//
	// 	DamageMultiplierAdditive: 1,
	// 	ThreatMultiplier:         1,
	// 	BonusCoefficient:         drainLifeCoeff,
	//
	// 	Dot: core.DotConfig{
	// 		Aura:                 core.Aura{Label: "Drain Life"},
	// 		NumberOfTicks:        int32(drainLifeRank.Duration() / tickLength),
	// 		TickLength:           tickLength,
	// 		AffectedByCastSpeed:  true,
	// 		HasteReducesDuration: true,
	// 		BonusCoefficient:     drainLifeCoeff,
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			dot.PeriodicDamageMultiplier = math.Max(1, math.Min(1+(soulSiphonPerAffliction*warlock.AfflictionCount(target)), cappedDmgBonus))
	// 			resultSlice[0] = dot.Spell.CalcAndDealPeriodicDamage(sim, target, drainLifeTick.Average(core.CharacterLevel), dot.OutcomeTick)
	// 			warlock.GainHealth(sim, resultSlice[0].Damage*warlock.PseudoStats.SelfHealingMultiplier, healthMetric)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
	// 		if result.Landed() {
	// 			spell.Dot(target).Apply(sim)
	// 			spell.DealOutcome(sim, result)
	// 		}
	// 	},
	// })
}
