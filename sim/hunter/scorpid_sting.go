package hunter

// TODO: To be implemented.
func (hunter *Hunter) registerScorpidStingSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// auraArray := hunter.NewEnemyAuraArray(func(unit *core.Unit) *core.Aura {
	// 	aura := buffs.ScorpidStingAura(unit, true, 0)
	// 	aura.Tag = "Sting"
	// 	return aura
	// })
	//
	// hunter.ScorpidSting = hunter.RegisterRangedSpell(core.SpellConfig{
	// 	ActionID:    core.ActionID{SpellID: 3043},
	// 	SpellSchool: core.SpellSchoolNature,
	// 	DefenseType: core.DefenseTypeRanged,
	// 	// A cast, not a proc, but one that must not read as a ranged hit to on-hit listeners; what
	// 	// the sting's application should count as is a separate question. Matches only listeners
	// 	// that state no mask.
	// 	ProcMask:       core.ProcMaskEmpty,
	// 	ClassSpellMask: HunterSpellScorpidSting,
	// 	Flags:          core.SpellFlagAPL,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		BaseCostPercent: 9,
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeRangedHit)
	//
	// 		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
	// 			if result.Landed() {
	// 				aura := auraArray.Get(target)
	// 				activeSting := target.GetActiveAuraWithTag("Sting")
	// 				if activeSting != nil && activeSting != aura {
	// 					activeSting.Deactivate(sim)
	// 				}
	// 				aura.Activate(sim)
	// 			}
	// 			spell.DealOutcome(sim, result)
	// 		})
	// 	},
	// })
}
