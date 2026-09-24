package mage

var pyroblastRank = spellData.Pyroblast.Highest()

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerPyroblastSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// actionID := core.ActionID{SpellID: pyroblastRank.ID}
	//
	// pyroblastDotCoefficient := 0.05000000075
	// pyroblastTick := pyroblastRank.PeriodicEffect()
	// tickLength := pyroblastTick.Period()
	//
	// mage.Pyroblast = mage.RegisterSpell(core.SpellConfig{
	// 	ActionID:       actionID,
	// 	SpellSchool:    core.SpellSchoolFire,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: MageSpellPyroblast,
	// 	MissileSpeed:   float64(pyroblastRank.Speed),
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(pyroblastRank.Cost()),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD:      pyroblastRank.GCD(),
	// 			CastTime: pyroblastRank.CastTime(),
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	BonusCoefficient: pyroblastRank.DamageEffect().Coeff(),
	// 	ThreatMultiplier: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := pyroblastRank.DamageEffect().Average(core.CharacterLevel)
	// 		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	//
	// 		spell.WaitTravelTime(sim, func(s *core.Simulation) {
	// 			spell.DealDamage(sim, result)
	// 			if result.Landed() {
	// 				spell.RelatedDotSpell.Cast(sim, target)
	// 			}
	// 		})
	// 	},
	// })
	//
	// mage.Pyroblast.RelatedDotSpell = mage.RegisterSpell(core.SpellConfig{
	// 	ActionID:       actionID.WithTag(1),
	// 	SpellSchool:    core.SpellSchoolFire,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	ClassSpellMask: MageSpellPyroblastDot,
	// 	Flags:          core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,
	//
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label: "PyroblastDoT",
	// 		},
	// 		NumberOfTicks:    int32(pyroblastRank.Duration() / tickLength),
	// 		TickLength:       tickLength,
	// 		BonusCoefficient: pyroblastDotCoefficient,
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, pyroblastTick.Average(core.CharacterLevel), dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		spell.Dot(target).Apply(sim)
	// 	},
	// })
}
