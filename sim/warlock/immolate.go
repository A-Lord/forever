package warlock

var immolateRank = spellData.Immolate.Highest()
var immolateTick = immolateRank.PeriodicEffect()
var immolateCoeff = immolateRank.DamageEffect().Coeff()
var immolateDotCoeff = immolateTick.Coeff()

// TODO: To be implemented. Port the TBC Immolate implementation below; not yet verified against the Forever client.
func (warlock *Warlock) registerImmolate() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// actionID := core.ActionID{SpellID: immolateRank.ID}
	// tickLength := immolateTick.Period()
	// tickCount := int32(immolateRank.Duration() / tickLength)
	// warlock.ImmolateTickBaseDamage = immolateTick.Average(core.CharacterLevel)
	//
	// warlock.Immolate = warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       actionID,
	// 	SpellSchool:    core.SpellSchoolFire,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: WarlockSpellImmolate,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(immolateRank.Cost()),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD:      immolateRank.GCD(),
	// 			CastTime: immolateRank.CastTime(),
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	DefenseType:      core.DefenseTypeMagic,
	// 	ThreatMultiplier: 1,
	// 	BonusCoefficient: immolateCoeff,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcDamage(sim, target, immolateRank.DamageEffect().Average(core.CharacterLevel), spell.OutcomeMagicHitAndCrit)
	// 		if result.Landed() {
	// 			spell.RelatedDotSpell.Dot(target).Apply(sim)
	// 		}
	//
	// 		spell.DealDamage(sim, result)
	// 	},
	// })
	//
	// warlock.Immolate.RelatedDotSpell = warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       actionID.WithTag(1),
	// 	SpellSchool:    core.SpellSchoolFire,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	ClassSpellMask: WarlockSpellImmolateDot,
	// 	Flags:          core.SpellFlagPassiveSpell,
	//
	// 	DamageMultiplier: 1,
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label: "Immolate (DoT)",
	// 		},
	// 		NumberOfTicks:    tickCount,
	// 		TickLength:       tickLength,
	// 		BonusCoefficient: immolateDotCoeff,
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, warlock.ImmolateTickBaseDamage, dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
	// 		dot := spell.Dot(target)
	// 		result := spell.CalcPeriodicDamage(sim, target, warlock.ImmolateTickBaseDamage*float64(tickCount), spell.OutcomeExpectedMagicHit)
	// 		result.Damage /= dot.CalcTickPeriod().Round(time.Millisecond).Seconds()
	// 		return result
	// 	},
	// })
}
