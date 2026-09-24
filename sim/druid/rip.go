package druid

var ripRank = spellData.Rip.Highest()
var ripTick = ripRank.PeriodicEffect()

// TODO: To be implemented.
func (druid *Druid) registerRipSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// var cp int32
	//
	// druid.Rip = druid.RegisterSpell(core.SpellConfig{
	// 	ActionID:        core.ActionID{SpellID: ripRank.ID},
	// 	CastRequirement: ripRank.CastRequirement(),
	// 	SpellSchool:     ripRank.SpellSchool(),
	// 	DefenseType:     ripRank.DefenseTypeCore(),
	// 	ProcMask:        core.ProcMaskMeleeMHSpecial,
	// 	ClassSpellMask:  DruidSpellRip,
	// 	Flags:           core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost: int32(ripRank.Cost()),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: ripRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	// 	ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
	// 		return druid.ComboPoints() > 0
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	// 	MaxRange:         core.MaxMeleeRange,
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label: "Rip",
	// 		},
	// 		NumberOfTicks: int32(ripRank.Duration() / ripTick.Period()),
	// 		TickLength:    ripTick.Period(),
	//
	// 		OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			druid.UpdateBleedPower(druid.Rip, sim, target, true, true)
	// 		},
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			ap := dot.Spell.MeleeAttackPower(target)
	//
	// 			var tickDamage float64
	// 			switch {
	// 			case cp <= 3:
	// 				tickDamage = 990 + 0.18*ap
	// 			case cp == 4:
	// 				tickDamage = 1272 + 0.24*ap
	// 			default: // 5
	// 				tickDamage = 1554 + 0.24*ap
	// 			}
	// 			tickDamage = tickDamage / 6
	//
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, tickDamage, dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
	// 		if result.Landed() {
	// 			cp = druid.ComboPoints()
	// 			spell.Dot(target).Apply(sim)
	// 			druid.SpendComboPoints(sim, spell.ComboPointMetrics())
	// 		}
	// 		spell.DealOutcome(sim, result)
	// 	},
	//
	// 	ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
	// 		// Assume 5 CP for projections.
	// 		ap := spell.MeleeAttackPower(target)
	// 		tickDamage := (1554 + 0.24*ap) / 6
	// 		result := spell.CalcPeriodicDamage(sim, target, tickDamage, spell.OutcomeExpectedMagicAlwaysHit)
	// 		attackTable := spell.Unit.AttackTables[target.UnitIndex]
	// 		critChance := spell.PhysicalCritChance(attackTable)
	// 		result.Damage *= 1 + critChance*(spell.CritDamageMultiplier(attackTable)-1)
	// 		return result
	// 	},
	// })
	//
	// druid.Rip.ShortName = "Rip"
}

func (druid *Druid) CurrentRipCost() float64 {
	return druid.Rip.Cost.GetCurrentCost()
}
