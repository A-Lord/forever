package hunter

var arcaneShotRank = spellData.ArcaneShot.Highest()

// TODO: To be implemented.
func (hunter *Hunter) registerArcaneShotSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// hunter.ArcaneShot = hunter.RegisterRangedSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: arcaneShotRank.ID},
	// 	SpellSchool:    arcaneShotRank.SpellSchool(),
	// 	DefenseType:    arcaneShotRank.DefenseTypeCore(),
	// 	ClassSpellMask: HunterSpellArcaneShot,
	// 	ProcMask:       core.ProcMaskRangedSpecial,
	// 	Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(arcaneShotRank.Cost()),
	// 	},
	//
	// 	Cast: core.CastConfig{
	// 		CD: core.Cooldown{
	// 			Timer:    hunter.NewTimer(),
	// 			Duration: max(arcaneShotRank.Cooldown(), arcaneShotRank.CategoryCooldown()),
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := spell.RangedAttackPower(target)*0.15 +
	// 			hunter.talonOfAlarBonus() +
	// 			arcaneShotRank.DamageEffect().Average(core.CharacterLevel)
	//
	// 		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)
	//
	// 		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
	// 			spell.DealDamage(sim, result)
	// 		})
	// 	},
	// })
}
