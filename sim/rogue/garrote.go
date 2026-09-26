package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var garroteRank = spellData.Garrote.Highest()

func (rogue *Rogue) registerGarrote() {
	tick := garroteRank.PeriodicEffect()
	apScaling := 0.03 // Estimated from in-game tests.
	comboPointGen := garroteRank.EnergizeEffect().Average(core.CharacterLevel)

	config := spelldata.SpellConfig(&rogue.Unit, garroteRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.Dot = spelldata.DotConfig(garroteRank, tick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, tick.Average(core.CharacterLevel)+dot.Spell.MeleeAttackPower(target)*apScaling, garroteRank.TickOutcome(dot))
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
			rogue.AddComboPoints(sim, int32(comboPointGen), spell.ComboPointMetrics())
		} else {
			spell.IssueRefund(sim)
		}
	}

	rogue.Garrote = rogue.RegisterSpell(config)

	// The TBC implementation, kept for the port:
	// tick := garroteRank.PeriodicEffect()
	// tickLength := tick.Period()
	//
	// rogue.Garrote = rogue.GetOrRegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: garroteRank.ID},
	// 	SpellSchool:    garroteRank.SpellSchool(),
	// 	DefenseType:    garroteRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	Flags:          core.SpellFlagMeleeMetrics | SpellFlagBuilder | core.SpellFlagAPL,
	// 	ClassSpellMask: RogueSpellGarrote,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost:   int32(garroteRank.Cost()),
	// 		Refund: 0.8,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: garroteRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	// 	ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
	// 		return !rogue.PseudoStats.InFrontOfTarget && rogue.IsStealthed()
	// 	},
	//
	// 	DamageMultiplierAdditive: 1,
	// 	DamageMultiplier:         1,
	// 	ThreatMultiplier:         1,
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label: "Garrote",
	// 			Tag:   RogueBleedTag,
	// 		},
	// 		NumberOfTicks: int32(garroteRank.Duration() / tickLength),
	// 		TickLength:    tickLength,
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, tick.Average(core.CharacterLevel)+dot.Spell.MeleeAttackPower(target)*0.03, dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		rogue.BreakStealth(sim)
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialNoBlockDodgeParryNoCrit)
	// 		if result.Landed() {
	// 			rogue.AddComboPoints(sim, 1, spell.ComboPointMetrics())
	// 			spell.Dot(target).Apply(sim)
	// 		} else {
	// 			spell.IssueRefund(sim)
	// 		}
	// 		spell.DealOutcome(sim, result)
	// 	},
	// })
}
