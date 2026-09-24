package druid

var insectSwarmRank = spellData.InsectSwarm.Highest()
var insectSwarmTick = insectSwarmRank.PeriodicEffect()

// TODO: To be implemented.
func (druid *Druid) registerInsectSwarmSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// druid.InsectSwarm = druid.RegisterSpell(core.SpellConfig{
	// 	ActionID:        core.ActionID{SpellID: insectSwarmRank.ID},
	// 	CastRequirement: insectSwarmRank.CastRequirement(),
	// 	SpellSchool:     insectSwarmRank.SpellSchool(),
	// 	DefenseType:     insectSwarmRank.DefenseTypeCore(),
	// 	ProcMask:        core.ProcMaskSpellDamage,
	// 	ClassSpellMask:  DruidSpellInsectSwarm,
	// 	Flags:           core.SpellFlagAPL | core.SpellFlagBinary,
	//
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	// 	MaxRange:         float64(insectSwarmRank.MaxRange),
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(insectSwarmRank.Cost()),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: insectSwarmRank.GCD(),
	// 		},
	// 	},
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label: "Insect Swarm",
	// 		},
	//
	// 		NumberOfTicks:       int32(insectSwarmRank.Duration() / insectSwarmTick.Period()),
	// 		TickLength:          insectSwarmTick.Period(),
	// 		AffectedByCastSpeed: false,
	// 		BonusCoefficient:    insectSwarmTick.Coeff(),
	//
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, insectSwarmTick.Average(core.CharacterLevel), dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
	//
	// 		if result.Landed() {
	// 			spell.Dot(target).Apply(sim)
	// 		}
	//
	// 		spell.DealOutcome(sim, result)
	// 	},
	// })
}
