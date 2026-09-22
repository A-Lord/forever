package druid

// The dbcenums and core imports belong with the commented implementation.

var hurricaneRank = spellData.Hurricane.Highest()

// TODO: To be implemented.
func (druid *Druid) registerHurricaneSpell() {
	panic("To be implemented")

	// The ported implementation, kept until this class is done:
	// tickLength := hurricaneRank.Effect(dbcenums.A_PERIODIC_DUMMY, 0).Period()
	//
	// // Hurricane's periodic damage is the spell HurricaneTriggered casts each tick.
	// hurricaneTickSpell := spellData.HurricaneTriggered.Highest()
	// hurricaneTick := hurricaneTickSpell.DamageEffect()
	//
	// druid.Hurricane = druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: hurricaneRank.ID},
	// 	SpellSchool:    hurricaneRank.SpellSchool(),
	// 	DefenseType:    hurricaneRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagChanneled | core.SpellFlagAPL,
	// 	ClassSpellMask: DruidSpellHurricane,
	// 	MaxRange:       float64(hurricaneRank.MaxRange),
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: hurricaneRank.Cost(),
	// 	},
	// 	// TODO: Forever states no cooldown on Hurricane (the client rows carry none), so the
	// 	// spell is registered without one rather than with an invented duration.
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: hurricaneRank.GCD(),
	// 		},
	// 	},
	// 	Dot: core.DotConfig{
	// 		IsAOE: true,
	// 		Aura: core.Aura{
	// 			Label: "Hurricane (Aura)",
	// 		},
	// 		NumberOfTicks:       int32(hurricaneRank.Duration() / tickLength),
	// 		TickLength:          tickLength,
	// 		AffectedByCastSpeed: true,
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			druid.Hurricane.RelatedDotSpell.Cast(sim, target)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
	// 		spell.AOEDot().Apply(sim)
	// 	},
	// })
	//
	// druid.Hurricane.RelatedDotSpell = druid.Unit.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: hurricaneTickSpell.ID},
	// 	SpellSchool:    core.SpellSchoolNature,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	ClassSpellMask: DruidSpellHurricane,
	// 	// The tick is its own client row that the channel triggers, a proc rather than a cast.
	// 	Flags: core.SpellFlagProc,
	//
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	// 	BonusCoefficient: hurricaneTick.Coeff(),
	//
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
	// 		spell.CalcAndDealAoeDamage(sim, hurricaneTick.Average(core.CharacterLevel), spell.OutcomeMagicHit)
	// 	},
	// })
}
