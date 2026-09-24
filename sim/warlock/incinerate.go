package warlock

var incinerateRank = spellData.Incinerate.Highest()
var incinerateCoeff = incinerateRank.DamageEffect().Coeff()

// TODO: To be implemented. Port the TBC Incinerate implementation below; not yet verified against the Forever client.
func (warlock *Warlock) registerIncinerate() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: incinerateRank.ID},
	// 	SpellSchool:    incinerateRank.SpellSchool(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	MissileSpeed:   float64(incinerateRank.Speed),
	// 	ClassSpellMask: WarlockSpellIncinerate,
	//
	// 	ManaCost: core.ManaCostOptions{FlatCost: int32(incinerateRank.Cost())},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD:      incinerateRank.GCD(),
	// 			CastTime: incinerateRank.CastTime(),
	// 		},
	// 	},
	//
	// 	DamageMultiplierAdditive: 1,
	// 	DefenseType:              incinerateRank.DefenseTypeCore(),
	// 	ThreatMultiplier:         1,
	// 	BonusCoefficient:         incinerateCoeff,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := incinerateRank.DamageEffect().Average(core.CharacterLevel)
	// 		if warlock.Immolate.Dot(target).IsActive() {
	// 			baseDamage += sim.Roll(111, 128)
	// 		}
	// 		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	//
	// 		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
	// 			spell.DealDamage(sim, result)
	// 		})
	// 	},
	// })
}
