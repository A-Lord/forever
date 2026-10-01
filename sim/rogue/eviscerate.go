package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var eviscerateRank = spellData.Eviscerate.Rank(8)

func (rogue *Rogue) registerEviscerate() {
	eviscerateBaseDamage := eviscerateRank.DamageEffect().Average()
	eviscerateDamagePerPoint := float64(eviscerateRank.DamageEffect().PointsPerResource)
	eviscerateApScalingPerPoint := 0.04 // Estimated from in-game tests.
	config := spelldata.SpellConfig(&rogue.Unit, eviscerateRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = rogue.FinisherExtraCastCondition
	config.Flags |= SpellFlagFinisher | SpellFlagColdBlooded

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		comboPoints := float64(rogue.ComboPoints())
		flatBaseDamage := eviscerateBaseDamage + eviscerateDamagePerPoint*comboPoints

		baseDamage := flatBaseDamage + eviscerateApScalingPerPoint*comboPoints*spell.MeleeAttackPower(target)

		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		if result.Landed() {
			rogue.ApplyFinisher(sim, spell)
		} else {
			spell.IssueRefund(sim)
		}

		spell.DealDamage(sim, result)
	}

	rogue.Eviscerate = rogue.RegisterSpell(config)

	// The TBC implementation, kept for the port:
	// flatDamage := eviscerateRank.DamageEffect().Average(core.CharacterLevel)
	// comboDamageBonus := 185.0 + rogue.DeathmantleBonus
	//
	// rogue.Eviscerate = rogue.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: eviscerateRank.ID},
	// 	SpellSchool:    eviscerateRank.SpellSchool(),
	// 	DefenseType:    eviscerateRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	Flags:          core.SpellFlagMeleeMetrics | SpellFlagFinisher | core.SpellFlagAPL,
	// 	MetricSplits:   6,
	// 	ClassSpellMask: RogueSpellEviscerate,
	// 	ClassFlags:     eviscerateRank.ClassFlags,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost: int32(eviscerateRank.Cost()),
	// 		// TODO: Forever drops Quick Recovery; no energy refund until we know whether the
	// 		// effect moved onto another talent.
	// 		Refund:        0,
	// 		RefundMetrics: rogue.EnergyRefundMetrics,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: eviscerateRank.GCD(),
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
	// 	DamageMultiplier:         1,
	// 	DamageMultiplierAdditive: 1,
	// 	ThreatMultiplier:         1,
	//
	// 	BonusCoefficient: eviscerateRank.DamageEffect().Coeff(),
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		rogue.BreakStealth(sim)
	//
	// 		comboPoints := float64(rogue.ComboPoints())
	// 		flatBaseDamage := flatDamage + comboDamageBonus*float64(comboPoints)
	//
	// 		baseDamage := flatBaseDamage + 0.03*float64(comboPoints)*spell.MeleeAttackPower(target)
	//
	// 		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
	//
	// 		if result.Landed() {
	// 			rogue.ApplyFinisher(sim, spell)
	// 		} else {
	// 			spell.IssueRefund(sim)
	// 		}
	//
	// 		spell.DealDamage(sim, result)
	// 	},
	// })
}
