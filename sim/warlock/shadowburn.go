package warlock

var shadowBurnRank = spellData.Shadowburn.Highest()
var shadowBurnCoeff = shadowBurnRank.DamageEffect().Coeff()

// TODO: To be implemented. Port the TBC Shadow Burn implementation below; not yet verified against the Forever client.
func (warlock *Warlock) registerShadowBurn() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	//
	// warlock.Shadowburn = warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: shadowBurnRank.ID},
	// 	SpellSchool:    shadowBurnRank.SpellSchool(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL | core.SpellFlagBinary,
	// 	ClassSpellMask: WarlockSpellShadowBurn,
	//
	// 	ManaCost: core.ManaCostOptions{FlatCost: int32(shadowBurnRank.Cost())},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: shadowBurnRank.GCD(),
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    warlock.NewTimer(),
	// 			Duration: max(shadowBurnRank.Cooldown(), shadowBurnRank.CategoryCooldown()),
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	DefenseType:      shadowBurnRank.DefenseTypeCore(),
	// 	ThreatMultiplier: 1,
	// 	BonusCoefficient: shadowBurnCoeff,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		dmgRoll := shadowBurnRank.DamageEffect().Average(core.CharacterLevel)
	// 		spell.CalcAndDealDamage(sim, target, dmgRoll, spell.OutcomeMagicHitAndCrit)
	//
	// 	},
	// })
}
