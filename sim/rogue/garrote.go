package rogue

var garroteRank = spellData.Garrote.Highest()

// TODO: To be implemented. Garrote already resolves against Forever data
// (spellData.Garrote.Highest()); the TBC body needs review before it's uncommented.
func (rogue *Rogue) registerGarrote() {
	panic("To be implemented")

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
	// 		Cost:   garroteRank.Cost(),
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
