package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var ruptureRank = spellData.Rupture.Highest()
var ruptureApScaling = [6]float64{0, 0.01, 0.02, 0.03, 0.03, 0.03}

func (rogue *Rogue) registerRupture() {
	tick := ruptureRank.PeriodicEffect()
	tickBase := tick.Average()
	tickPerPoint := float64(tick.PointsPerResource)
	baseTicks := int32(ruptureRank.Duration() / tick.Period())
	var comboPoints int32

	config := spelldata.SpellConfig(&rogue.Unit, ruptureRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = rogue.FinisherExtraCastCondition
	config.ClassSpellMask = RogueSpellRupture

	config.Dot = spelldata.DotConfig(ruptureRank, tick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		damage := tickBase + tickPerPoint*float64(comboPoints) + dot.Spell.MeleeAttackPower(target)*ruptureApScaling[comboPoints]
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, ruptureRank.TickOutcome(dot))
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if result.Landed() {
			comboPoints = rogue.ComboPoints()
			dot := spell.Dot(target)
			dot.BaseTickCount = baseTicks + comboPoints
			dot.Apply(sim)
			rogue.ApplyFinisher(sim, spell)
		} else {
			spell.IssueRefund(sim)
		}
		spell.DealOutcome(sim, result)
	}

	rogue.Rupture = rogue.RegisterSpell(config)

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
	// 	ClassFlags:     ruptureRank.ClassFlags,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost: int32(ruptureRank.Cost()),
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
