package druid

var wrathRank = spellData.Wrath.Highest()

// TODO: To be implemented.
func (druid *Druid) registerWrathSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// druid.Wrath = druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: wrathRank.ID},
	// 	SpellSchool:    wrathRank.SpellSchool(),
	// 	DefenseType:    wrathRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	ClassSpellMask: DruidSpellWrath,
	// 	Flags:          core.SpellFlagAPL,
	// 	MissileSpeed:   float64(wrathRank.Speed),
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: wrathRank.Cost(),
	// 	},
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD:      wrathRank.GCD(),
	// 			CastTime: wrathRank.CastTime(),
	// 		},
	// 	},
	//
	// 	BonusCoefficient: wrathRank.DamageEffect().Coeff(),
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	// 	MaxRange:         float64(wrathRank.MaxRange),
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := wrathRank.DamageEffect().Average(core.CharacterLevel)
	// 		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	//
	// 		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
	// 			spell.DealDamage(sim, result)
	// 		})
	// 	},
	// })
}
