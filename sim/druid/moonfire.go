package druid

var moonfireRank = spellData.Moonfire.Highest()
var moonfireTick = moonfireRank.PeriodicEffect()

func (druid *Druid) registerMoonfireSpell() {
	druid.registerMoonfireImpactSpell()
	druid.registerMoonfireDoTSpell()
}

// TODO: To be implemented.
func (druid *Druid) registerMoonfireDoTSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// druid.Moonfire.RelatedDotSpell = druid.Unit.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: moonfireRank.ID}.WithTag(1),
	// 	SpellSchool:    moonfireRank.SpellSchool(),
	// 	DefenseType:    moonfireRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	ClassSpellMask: DruidSpellMoonfireDoT,
	// 	Flags:          core.SpellFlagPassiveSpell,
	//
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label: "Moonfire",
	// 		},
	// 		NumberOfTicks:       int32(moonfireRank.Duration() / moonfireTick.Period()),
	// 		TickLength:          moonfireTick.Period(),
	// 		AffectedByCastSpeed: false,
	// 		BonusCoefficient:    moonfireTick.Coeff(),
	//
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, moonfireTick.Average(core.CharacterLevel), dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeAlwaysHitNoHitCounter)
	//
	// 		spell.Dot(target).Apply(sim)
	// 		spell.DealOutcome(sim, result)
	// 	},
	// })
}

// TODO: To be implemented.
func (druid *Druid) registerMoonfireImpactSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// druid.Moonfire = druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: moonfireRank.ID},
	// 	SpellSchool:    moonfireRank.SpellSchool(),
	// 	DefenseType:    moonfireRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	ClassSpellMask: DruidSpellMoonfire,
	// 	Flags:          core.SpellFlagAPL,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: moonfireRank.Cost(),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: moonfireRank.GCD(),
	// 		},
	// 	},
	//
	// 	BonusCoefficient: moonfireRank.DamageEffect().Coeff(),
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	// 	MaxRange:         float64(moonfireRank.MaxRange),
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := moonfireRank.DamageEffect().Average(core.CharacterLevel)
	// 		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	//
	// 		if result.Landed() {
	// 			druid.Moonfire.RelatedDotSpell.Cast(sim, target)
	// 		}
	//
	// 		spell.DealDamage(sim, result)
	// 	},
	// })
}
