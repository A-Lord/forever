package warlock

// Package-level state the commented-out implementations used:
// var agonyRank = spellData.CurseOfAgony.ByID(27218)
// var agonyTick = agonyRank.PeriodicEffect()
// var agonyCoeff = agonyTick.Coeff()

// TODO: To be implemented. Forever renamed this to **Bane of Agony**: spells 980, 1014, 6217, 11711 (and up)
// on the Affliction line, a full rank chain. The registrar needs re-pointing at that name,
// not implementing from nothing.
func (warlock *Warlock) registerCurseOfAgony() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	//
	// tickLength := agonyTick.Period()
	// calculateBaseDamage := func(sim *core.Simulation, dot *core.Dot) float64 {
	// 	damageMultiplier := core.TernaryFloat64(warlock.AmplifyCurseAura != nil && warlock.AmplifyCurseAura.IsActive(), 1.5, 1.0)
	// 	return agonyTick.Average(core.CharacterLevel) * damageMultiplier
	// }
	//
	// warlock.CurseOfAgony = warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: agonyRank.ID},
	// 	Flags:          core.SpellFlagAPL,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	SpellSchool:    agonyRank.SpellSchool(),
	// 	DefenseType:    agonyRank.DefenseTypeCore(),
	// 	ClassSpellMask: WarlockSpellCurseOfAgony,
	//
	// 	ThreatMultiplier: 1,
	// 	DamageMultiplier: 1,
	// 	BonusCoefficient: agonyCoeff,
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: agonyRank.GCD(),
	// 		},
	// 	},
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: agonyRank.Cost(),
	// 	},
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
	// 		if result.Landed() {
	// 			warlock.DeactivateOtherCurses(sim, spell, target)
	// 			spell.Dot(target).Apply(sim)
	// 		}
	// 		spell.DealOutcome(sim, result)
	// 	},
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label: "Agony",
	// 			Tag:   "Affliction",
	// 		},
	//
	// 		TickLength:               tickLength,
	// 		NumberOfTicks:            int32(agonyRank.Duration() / tickLength),
	// 		PeriodicDamageMultiplier: 1,
	//
	// 		BonusCoefficient: agonyCoeff,
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, calculateBaseDamage(sim, dot), dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
	// 		dot := spell.Dot(target)
	//
	// 		// Always compare fully stacked agony damage
	// 		result := spell.CalcPeriodicDamage(sim, target, calculateBaseDamage(sim, dot), spell.OutcomeExpectedMagicHit)
	// 		result.Damage *= 10
	// 		result.Damage /= dot.CalcTickPeriod().Round(time.Millisecond).Seconds()
	// 		return result
	// 	},
	// })
}
