package rogue

import (
	"github.com/wowsims/forever/sim/core"
)

// Was the TBC rank-6 id 26867, which the level squish removed. Derived from the table so
// it follows the data instead of naming a rank that may not exist.
var RuptureSpellID = spellData.Rupture.Highest().ID

var ruptureRank = spellData.Rupture.ByID(RuptureSpellID)

// TODO: To be implemented. Rupture already resolves against Forever data
// (spellData.Rupture.ByID(RuptureSpellID)); the TBC body needs review before it's
// uncommented.
func (rogue *Rogue) registerRupture() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// tick := ruptureRank.PeriodicEffect()
	// tickLength := tick.Period()
	// baseTickCount := int32(ruptureRank.Duration() / tickLength)
	// var cp int32
	//
	// rogue.Rupture = rogue.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: ruptureRank.ID},
	// 	SpellSchool:    core.SpellSchoolPhysical,
	// 	DefenseType:    core.DefenseTypeMelee,
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	Flags:          core.SpellFlagMeleeMetrics | SpellFlagFinisher | core.SpellFlagAPL,
	// 	MetricSplits:   6,
	// 	ClassSpellMask: RogueSpellRupture,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost: ruptureRank.Cost(),
	// 		// TODO: Forever drops Quick Recovery; no energy refund until we know whether the
	// 		// effect moved onto another talent.
	// 		Refund:        0,
	// 		RefundMetrics: rogue.EnergyRefundMetrics,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: ruptureRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 		ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
	// 			spell.SetMetricsSplit(rogue.ComboPoints())
	// 		},
	// 	},
	// 	ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
	// 		return rogue.ComboPoints() > 0
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label: "Rupture",
	// 			Tag:   RogueBleedTag,
	// 		},
	// 		NumberOfTicks: 0, // Set dynamically
	// 		TickLength:    tickLength,
	//
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, rogue.ruptureDamage(target, cp, tick.Average(core.CharacterLevel), 11), dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		rogue.BreakStealth(sim)
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
	// 		if result.Landed() {
	// 			cp = rogue.ComboPoints()
	// 			dot := spell.Dot(target)
	// 			dot.BaseTickCount = baseTickCount + cp
	// 			dot.Apply(sim)
	// 			rogue.ApplyFinisher(sim, spell)
	// 			spell.DealOutcome(sim, result)
	// 		} else {
	// 			spell.DealOutcome(sim, result)
	// 			spell.IssueRefund(sim)
	// 		}
	//
	// 	},
	// })
}

func (rogue *Rogue) ruptureDamage(target *core.Unit, comboPoints int32, baseDamage float64, damagePerComboPoint float64) float64 {
	return baseDamage +
		damagePerComboPoint*float64(comboPoints) +
		[]float64{0, 0.01, 0.02, 0.03, 0.03, 0.03}[comboPoints]*rogue.Rupture.MeleeAttackPower(target)
}
