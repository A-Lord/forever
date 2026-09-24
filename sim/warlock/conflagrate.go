package warlock

var conflagrateRank = spellData.Conflagrate.Highest()
var conflagrateCoeff = conflagrateRank.DamageEffect().Coeff()

// TODO: To be implemented. Port the TBC Conflagrate implementation below; not yet verified against the Forever client.
func (warlock *Warlock) registerConflagrate() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	//
	// if !warlock.Talents.Conflagrate {
	// 	return
	// }
	//
	// warlock.Conflagrate = warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: conflagrateRank.ID},
	// 	SpellSchool:    conflagrateRank.SpellSchool(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: WarlockSpellConflagrate,
	//
	// 	ManaCost: core.ManaCostOptions{FlatCost: int32(conflagrateRank.Cost())},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: conflagrateRank.GCD(),
	// 		},
	// 		CD: core.Cooldown{
	// 			Duration: max(conflagrateRank.Cooldown(), conflagrateRank.CategoryCooldown()),
	// 		},
	// 	},
	// 	ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
	// 		return warlock.Immolate.Dot(target).IsActive()
	// 	},
	//
	// 	DamageMultiplier: 1.0,
	// 	DefenseType:      conflagrateRank.DefenseTypeCore(),
	// 	ThreatMultiplier: 1,
	// 	BonusCoefficient: conflagrateCoeff,
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		//tie this to landed/hit
	// 		dmgRoll := conflagrateRank.DamageEffect().Average(core.CharacterLevel)
	// 		result := spell.CalcAndDealDamage(sim, target, dmgRoll, spell.OutcomeMagicHitAndCrit)
	//
	// 		if result.Landed() || result.DidResist() {
	// 			warlock.Immolate.Dot(target).Deactivate(sim)
	// 		}
	// 	},
	// })
}
