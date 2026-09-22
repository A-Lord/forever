package rogue

import (
	"github.com/wowsims/forever/sim/core"
)

var exposeArmorRank = spellData.ExposeArmor.Highest()

// TODO: To be implemented. Expose Armor already resolves against Forever data
// (spellData.ExposeArmor.Highest()); the TBC body needs review before it's uncommented.
func (rogue *Rogue) registerExposeArmorSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// rogue.ExposeArmorAuras = rogue.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
	// 	return core.ExposeArmorAura(target, rogue.ComboPoints, rogue.Talents.ImprovedExposeArmor)
	// })
	//
	// rogue.ExposeArmor = rogue.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: exposeArmorRank.ID},
	// 	SpellSchool:    exposeArmorRank.SpellSchool(),
	// 	DefenseType:    exposeArmorRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
	// 	MetricSplits:   6,
	// 	ClassSpellMask: RogueSpellExposeArmor,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost: exposeArmorRank.Cost(),
	// 		// TODO: Forever drops Quick Recovery; no energy refund until we know whether the
	// 		// effect moved onto another talent.
	// 		Refund:        0,
	// 		RefundMetrics: rogue.EnergyRefundMetrics,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: exposeArmorRank.GCD(),
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
	// 	ThreatMultiplier: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		if rogue.CanApplyExposeArmorAura(target) {
	// 			rogue.BreakStealth(sim)
	// 			result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
	// 			if result.Landed() {
	// 				rogue.ExposeArmorAuras.Get(target).Activate(sim)
	// 				rogue.ApplyFinisher(sim, spell)
	// 			} else {
	// 				spell.IssueRefund(sim)
	// 			}
	// 			spell.DealOutcome(sim, result)
	// 		}
	// 	},
	//
	// 	RelatedAuraArrays: rogue.ExposeArmorAuras.ToMap(),
	// })
}

func (rogue *Rogue) GetExposeArmorValue() float64 {
	// TODO: Forever repurposes Improved Expose Armor: the spell now carries an energy cost
	// reduction (SPELLMOD_COST -5/-10) and a dummy of 1/2, neither of which is the 25/50%
	// armor bonus this call wants, so the armor value is pinned to the untalented one.
	improvedExposeArmorMultiplier := 1.0
	return 410.0 * float64(rogue.ComboPoints()) * improvedExposeArmorMultiplier
}

func (rogue *Rogue) CanApplyExposeArmorAura(target *core.Unit) bool {
	return !rogue.ExposeArmorAuras.Get(target).IsActive() || rogue.ExposeArmorAuras.Get(target).ExclusiveEffects[0].Priority <= rogue.GetExposeArmorValue()
}
