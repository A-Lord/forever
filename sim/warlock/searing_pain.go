package warlock

var searingPainRank = spellData.SearingPain.Highest()
var searingPainCoeff = searingPainRank.DamageEffect().Coeff()

// TODO: To be implemented. Port the TBC Searing Pain implementation below; not yet verified against the Forever client.
func (warlock *Warlock) registerSearingPain() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	//
	// warlock.Shadowburn = warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: searingPainRank.ID},
	// 	SpellSchool:    searingPainRank.SpellSchool(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: WarlockSpellSearingPain,
	// 	MaxRange:       float64(searingPainRank.MaxRange),
	//
	// 	ManaCost: core.ManaCostOptions{FlatCost: searingPainRank.Cost()},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD:      searingPainRank.GCD(),
	// 			CastTime: searingPainRank.CastTime(),
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	DefenseType:      searingPainRank.DefenseTypeCore(),
	// 	ThreatMultiplier: 2,
	// 	BonusCoefficient: searingPainCoeff,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		dmgRoll := searingPainRank.DamageEffect().Average(core.CharacterLevel)
	// 		spell.CalcAndDealDamage(sim, target, dmgRoll, spell.OutcomeMagicHitAndCrit)
	// 	},
	// })
}
