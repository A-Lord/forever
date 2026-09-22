package mage

var coneOfColdRank = spellData.ConeOfCold.Highest()

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerConeOfColdSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	//
	// coneOfColdCoefficient := 0.1930000037 // Per https://wago.tools/db2/SpellEffect?build=2.5.5.65295&filter%5BSpellID%5D=exact%253A120 Field "EffetBonusCoefficient"
	//
	// mage.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: coneOfColdRank.ID},
	// 	SpellSchool:    coneOfColdRank.SpellSchool(),
	// 	DefenseType:    coneOfColdRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL | core.SpellFlagBinary,
	// 	ClassSpellMask: MageSpellConeOfCold,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: coneOfColdRank.Cost(),
	// 	},
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: coneOfColdRank.GCD(),
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    mage.NewTimer(),
	// 			Duration: max(coneOfColdRank.Cooldown(), coneOfColdRank.CategoryCooldown()),
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	BonusCoefficient: coneOfColdCoefficient,
	// 	ThreatMultiplier: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
	// 		baseDamage := coneOfColdRank.DamageEffect().Average(core.CharacterLevel)
	// 		spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
	// 	},
	// })
}
