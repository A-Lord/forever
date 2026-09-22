package mage

var fireballRank = spellData.Fireball.Highest()

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerFireballSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// fireballTick := fireballRank.PeriodicEffect()
	// tickLength := fireballTick.Period()
	//
	// mage.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: fireballRank.ID},
	// 	SpellSchool:    fireballRank.SpellSchool(),
	// 	DefenseType:    fireballRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: MageSpellFireball,
	// 	MissileSpeed:   float64(fireballRank.Speed),
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: fireballRank.Cost(),
	// 	},
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD:      fireballRank.GCD(),
	// 			CastTime: fireballRank.CastTime(),
	// 		},
	// 	},
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label: "FireballDoT",
	// 		},
	// 		NumberOfTicks: int32(fireballRank.Duration() / tickLength),
	// 		TickLength:    tickLength,
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, fireballTick.Average(core.CharacterLevel), dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	BonusCoefficient: fireballRank.DamageEffect().Coeff(),
	// 	ThreatMultiplier: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := fireballRank.DamageEffect().Average(core.CharacterLevel)
	// 		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	//
	// 		spell.WaitTravelTime(sim, func(s *core.Simulation) {
	// 			spell.DealDamage(sim, result)
	// 			if result.Landed() {
	// 				spell.Dot(target).Apply(sim)
	// 			}
	// 		})
	//
	// 	},
	// })
}
