package warlock

var siphonLifeRank = spellData.SiphonLife.Highest()
var siphonLifeTick = siphonLifeRank.PeriodicEffect()
var siphonLifeCoeff = siphonLifeTick.Coeff()

// TODO: To be implemented. Port the TBC Siphon Life Spell implementation below; not yet verified against the Forever client.
func (warlock *Warlock) registerSiphonLifeSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// actionID := core.ActionID{SpellID: siphonLifeRank.ID}
	// tickLength := siphonLifeTick.Period()
	// baseCost := siphonLifeRank.Cost()
	//
	// healthMetrics := warlock.NewHealthMetrics(actionID)
	//
	// warlock.SiphonLife = warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       actionID,
	// 	SpellSchool:    core.SpellSchoolShadow,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ClassSpellMask: WarlockSpellSiphonLife,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	BaseCost:       baseCost,
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			Cost: baseCost,
	// 			GCD:  siphonLifeRank.GCD(),
	// 		},
	// 	},
	// 	DamageMultiplier: 1,
	// 	BonusCoefficient: siphonLifeCoeff,
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
	//
	// 		if result.Landed() {
	// 			spell.Dot(target).Apply(sim)
	// 		}
	// 		spell.DealOutcome(sim, result)
	// 	},
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label: "SiphonLife",
	// 			Tag:   "Affliction",
	// 		},
	// 		NumberOfTicks:    int32(siphonLifeRank.Duration() / tickLength),
	// 		TickLength:       tickLength,
	// 		BonusCoefficient: siphonLifeCoeff,
	//
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			result := dot.Spell.CalcAndDealPeriodicDamage(sim, target, siphonLifeTick.Average(core.CharacterLevel), dot.OutcomeTick)
	//
	// 			healthToRegain := result.Damage * (1 * warlock.PseudoStats.BonusHealingTaken)
	// 			warlock.GainHealth(sim, healthToRegain, healthMetrics)
	// 			dot.Spell.ApplyAOEThreat(healthToRegain * 0.5)
	// 		},
	// 	},
	//
	// 	ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
	// 		dot := spell.Dot(target)
	// 		result := dot.Spell.CalcPeriodicDamage(sim, target, siphonLifeTick.Average(core.CharacterLevel)*float64(int32(siphonLifeRank.Duration()/tickLength)), spell.OutcomeExpectedMagicHit)
	// 		result.Damage /= dot.CalcTickPeriod().Round(time.Millisecond).Seconds()
	// 		return result
	// 	},
	// })
}
