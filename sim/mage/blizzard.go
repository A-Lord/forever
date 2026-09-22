package mage

// The dbcenums and core imports belong with the commented implementation.

var blizzardRank = spellData.Blizzard.Highest()

// TODO: To be implemented.
func (mage *Mage) registerBlizzardSpell() {
	panic("To be implemented")

	// The ported implementation, kept until this class is done:
	// blizzardActionId := core.ActionID{SpellID: blizzardRank.ID}
	// tickLength := blizzardRank.Effect(dbcenums.A_PERIODIC_DUMMY, 0).Period()
	//
	// // Blizzard's periodic damage is the spell BlizzardTriggered casts each tick.
	// blizzardTick := spellData.BlizzardTriggered.Highest().DamageEffect()
	//
	// blizzardTickSpell := mage.RegisterSpell(core.SpellConfig{
	// 	ActionID:       blizzardActionId,
	// 	SpellSchool:    core.SpellSchoolFrost,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	ClassSpellMask: MageSpellBlizzard,
	//
	// 	DamageMultiplier: 1,
	// 	BonusCoefficient: blizzardTick.Coeff(),
	// 	ThreatMultiplier: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
	// 		spell.CalcAndDealAoeDamage(sim, blizzardTick.Average(core.CharacterLevel), spell.OutcomeMagicHit)
	// 	},
	// })
	//
	// mage.RegisterSpell(core.SpellConfig{
	// 	ActionID:       blizzardActionId,
	// 	SpellSchool:    core.SpellSchoolFrost,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagChanneled | core.SpellFlagAPL,
	// 	ClassSpellMask: MageSpellBlizzard,
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: blizzardRank.Cost(),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: blizzardRank.GCD(),
	// 		},
	// 	},
	// 	Dot: core.DotConfig{
	// 		IsAOE: true,
	// 		Aura: core.Aura{
	// 			Label:    "Blizzard",
	// 			ActionID: blizzardActionId,
	// 		},
	// 		NumberOfTicks:        int32(blizzardRank.Duration() / tickLength),
	// 		TickLength:           tickLength,
	// 		AffectedByCastSpeed:  true,
	// 		HasteReducesDuration: true,
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			blizzardTickSpell.Cast(sim, target)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		spell.AOEDot().Apply(sim)
	// 	},
	// })
}
