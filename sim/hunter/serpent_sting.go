package hunter

var serpentStingRank = spellData.SerpentSting.Highest()

// TODO: To be implemented.
func (hunter *Hunter) registerSerpentStingSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// serpentStingTick := serpentStingRank.PeriodicEffect()
	//
	// hunter.SerpentSting = hunter.RegisterRangedSpell(core.SpellConfig{
	// 	ActionID:    core.ActionID{SpellID: serpentStingRank.ID},
	// 	SpellSchool: serpentStingRank.SpellSchool(),
	// 	DefenseType: serpentStingRank.DefenseTypeCore(),
	// 	// A cast, not a proc, but one that must not read as a ranged hit to on-hit listeners; what
	// 	// the sting's application should count as is a separate question. Matches only listeners
	// 	// that state no mask.
	// 	ProcMask:       core.ProcMaskEmpty,
	// 	ClassSpellMask: HunterSpellSerpentSting,
	// 	Flags:          core.SpellFlagAPL,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(serpentStingRank.Cost()),
	// 	},
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label: "Serpent Sting",
	// 			Tag:   "Sting",
	// 		},
	//
	// 		NumberOfTicks: int32(serpentStingRank.Duration() / serpentStingTick.Period()),
	// 		TickLength:    serpentStingTick.Period(),
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			baseDmg := dot.Spell.RangedAttackPower(target)*0.02 + serpentStingTick.Average(core.CharacterLevel)
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, baseDmg, dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeRangedHit)
	//
	// 		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
	// 			if result.Landed() {
	// 				dot := spell.Dot(target)
	// 				activeSting := target.GetActiveAuraWithTag("Sting")
	// 				if activeSting != nil && activeSting != dot.Aura {
	// 					activeSting.Deactivate(sim)
	// 				}
	// 				dot.Apply(sim)
	// 			}
	// 			spell.DealOutcome(sim, result)
	// 		})
	// 	},
	// })
}
