package warlock

var deathCoilRank = spellData.DeathCoil.Highest()

// TODO: To be implemented. Port the TBC Death Coil implementation below; not yet verified against the Forever client.
func (warlock *Warlock) registerDeathCoil() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	//
	// warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: deathCoilRank.ID},
	// 	SpellSchool:    deathCoilRank.SpellSchool(),
	// 	DefenseType:    deathCoilRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: WarlockSpellDeathCoil,
	// 	MissileSpeed:   float64(deathCoilRank.Speed),
	// 	MaxRange:       float64(deathCoilRank.MaxRange),
	//
	// 	ManaCost: core.ManaCostOptions{FlatCost: deathCoilRank.Cost()},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: deathCoilRank.GCD(),
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    warlock.NewTimer(),
	// 			Duration: max(deathCoilRank.Cooldown(), deathCoilRank.CategoryCooldown()),
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	// 	BonusCoefficient: 0.214,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		spell.CalcAndDealDamage(sim, target, 526, spell.OutcomeMagicHit)
	// 	},
	// })
}
