package warlock

var soulfireRank = spellData.SoulFire.Highest()
var soulfireCoeff = soulfireRank.DamageEffect().Coeff()

// TODO: To be implemented. Port the TBC Soulfire implementation below; not yet verified against the Forever client.
func (warlock *Warlock) registerSoulfire() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// warlock.Soulfire = warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: soulfireRank.ID},
	// 	SpellSchool:    soulfireRank.SpellSchool(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: WarlockSpellSoulFire,
	// 	MissileSpeed:   float64(soulfireRank.Speed),
	//
	// 	ManaCost: core.ManaCostOptions{FlatCost: int32(soulfireRank.Cost())},
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD:      soulfireRank.GCD(),
	// 			CastTime: soulfireRank.CastTime() - time.Duration(400*warlock.Talents.Bane),
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    warlock.NewTimer(),
	// 			Duration: max(soulfireRank.Cooldown(), soulfireRank.CategoryCooldown()),
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	DefenseType:      soulfireRank.DefenseTypeCore(),
	// 	ThreatMultiplier: 1,
	// 	BonusCoefficient: soulfireCoeff,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		dmgRoll := soulfireRank.DamageEffect().Average(core.CharacterLevel)
	// 		result := spell.CalcDamage(sim, target, dmgRoll, spell.OutcomeMagicHitAndCrit)
	// 		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
	// 			spell.DealDamage(sim, result)
	// 		})
	// 	},
	// })
	//
}
