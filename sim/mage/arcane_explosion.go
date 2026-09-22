package mage

var arcaneExplosionRank = spellData.ArcaneExplosion.Highest()

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerArcaneExplosionSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// arcaneExplosionCoefficient := 0.21400000155
	//
	// mage.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: arcaneExplosionRank.ID},
	// 	SpellSchool:    arcaneExplosionRank.SpellSchool(),
	// 	DefenseType:    arcaneExplosionRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: MageSpellArcaneExplosion,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: arcaneExplosionRank.Cost(),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: arcaneExplosionRank.GCD(),
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	BonusCoefficient: arcaneExplosionCoefficient,
	// 	ThreatMultiplier: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := arcaneExplosionRank.DamageEffect().Average(core.CharacterLevel)
	// 		spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
	// 	},
	// })
}
