package rogue

var backstabRank = spellData.Backstab.Highest()

// TODO: To be implemented. Backstab already resolves against Forever data
// (spellData.Backstab.Highest()); the TBC body needs review before it's uncommented.
func (rogue *Rogue) registerBackstabSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// baseDamage := backstabRank.DamageEffect().Average(core.CharacterLevel)
	// weaponDamage := 1.5
	//
	// rogue.Backstab = rogue.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: backstabRank.ID},
	// 	SpellSchool:    backstabRank.SpellSchool(),
	// 	DefenseType:    backstabRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	Flags:          core.SpellFlagMeleeMetrics | SpellFlagBuilder | core.SpellFlagAPL,
	// 	ClassSpellMask: RogueSpellBackstab,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost:   int32(backstabRank.Cost()),
	// 		Refund: 0.8,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: backstabRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	// 	ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
	// 		return !rogue.PseudoStats.InFrontOfTarget && rogue.HasDagger(core.MainHand)
	// 	},
	//
	// 	DamageMultiplierAdditive: weaponDamage,
	// 	DamageMultiplier:         1,
	// 	ThreatMultiplier:         1,
	//
	// 	BonusCoefficient: backstabRank.DamageEffect().Coeff(),
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
