package hunter

var multiShotRank = spellData.MultiShot.Highest()

// TODO: To be implemented.
func (hunter *Hunter) registerMultiShotSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// hunter.MultiShot = hunter.RegisterRangedSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: multiShotRank.ID},
	// 	SpellSchool:    multiShotRank.SpellSchool(),
	// 	DefenseType:    multiShotRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskRangedSpecial,
	// 	ClassSpellMask: HunterSpellMultiShot,
	// 	Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
	//
	// 	MissileSpeed: float64(multiShotRank.Speed),
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(multiShotRank.Cost()),
	// 	},
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			CastTime: time.Millisecond * 500,
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    hunter.NewTimer(),
	// 			Duration: max(multiShotRank.Cooldown(), multiShotRank.CategoryCooldown()),
	// 		},
	// 	},
	//
	// 	BonusCoefficient: multiShotRank.DamageEffect().Coeff(),
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := spell.RangedAttackPower(target)*0.2 +
	// 			hunter.AutoAttacks.Ranged().BaseDamage(sim) +
	// 			hunter.talonOfAlarBonus() +
	// 			multiShotRank.DamageEffect().Average(core.CharacterLevel)
	//
	// 		spell.CalcAoeDamage(sim, baseDamage, spell.OutcomeRangedHitAndCrit)
	//
	// 		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
	// 			spell.DealBatchedAoeDamage(sim)
	// 		})
	// 	},
	// })
}
