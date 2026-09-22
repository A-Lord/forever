package rogue

var sinisterStrikeRank = spellData.SinisterStrike.Highest()

// TODO: To be implemented. Sinister Strike already resolves against Forever data
// (spellData.SinisterStrike.Highest()); the TBC body needs review before it's uncommented.
func (rogue *Rogue) registerSinisterStrikeSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// baseDamage := sinisterStrikeRank.DamageEffect().Average(core.CharacterLevel)
	//
	// rogue.SinisterStrike = rogue.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: sinisterStrikeRank.ID},
	// 	SpellSchool:    sinisterStrikeRank.SpellSchool(),
	// 	DefenseType:    sinisterStrikeRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	Flags:          core.SpellFlagMeleeMetrics | SpellFlagBuilder | core.SpellFlagAPL,
	// 	ClassSpellMask: RogueSpellSinisterStrike,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost:   sinisterStrikeRank.Cost(),
	// 		Refund: 0.8,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: sinisterStrikeRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	//
	// 	DamageMultiplier:         1,
	// 	DamageMultiplierAdditive: 1,
	// 	ThreatMultiplier:         1,
	//
	// 	BonusCoefficient: sinisterStrikeRank.DamageEffect().Coeff(),
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		rogue.BreakStealth(sim)
	// 		baseDamage := baseDamage +
	// 			spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
	//
	// 		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
	//
	// 		if result.Landed() {
	// 			rogue.AddComboPoints(sim, 1, spell.ComboPointMetrics())
	// 		} else {
	// 			spell.IssueRefund(sim)
	// 		}
	// 	},
	// })
}
