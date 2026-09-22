package warlock

// Package-level state the commented-out implementations used:
// var doomRank = spellData.CurseOfDoom.ByID(30910)
// var doomTick = doomRank.PeriodicEffect()
// var doomCoeff = doomTick.Coeff()

// TODO: To be implemented. Forever renamed this to **Bane of Doom**: spell 603 on the Affliction line.
// The registrar needs re-pointing at that name, not implementing from nothing.
func (warlock *Warlock) registerCurseOfDoom() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	//
	// tickLength := doomTick.Period()
	// calculateBaseDamage := func() float64 {
	// 	damageMultiplier := core.TernaryFloat64(warlock.AmplifyCurseAura != nil && warlock.AmplifyCurseAura.IsActive(), 1.5, 1.0)
	// 	return doomTick.Average(core.CharacterLevel) * damageMultiplier
	// }
	//
	// warlock.CurseOfDoom = warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: doomRank.ID},
	// 	SpellSchool:    doomRank.SpellSchool(),
	// 	DefenseType:    doomRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: WarlockSpellCurseOfDoom,
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: doomRank.GCD(),
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    warlock.NewTimer(),
	// 			Duration: max(doomRank.Cooldown(), doomRank.CategoryCooldown()),
	// 		},
	// 	},
	//
	// 	ThreatMultiplier: 1,
	// 	DamageMultiplier: 1,
	// 	BonusCoefficient: doomCoeff,
	//
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
	// 			Label: "Doom",
	// 			Tag:   "Affliction",
	// 		},
	// 		NumberOfTicks:            int32(doomRank.Duration() / tickLength),
	// 		TickLength:               tickLength,
	// 		BonusCoefficient:         doomCoeff,
	// 		PeriodicDamageMultiplier: 1,
	//
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, calculateBaseDamage(), dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
	// 		dot := spell.Dot(target)
	// 		return dot.Spell.CalcPeriodicDamage(sim, target, calculateBaseDamage(), spell.OutcomeExpectedMagicHit)
	// 	},
	// })
}
