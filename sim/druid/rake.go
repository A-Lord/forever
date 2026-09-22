package druid

var rakeRank = spellData.Rake.Highest()
var rakeTick = rakeRank.PeriodicEffect()

// TODO: To be implemented.
func (druid *Druid) registerRakeSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// druid.Rake = druid.RegisterSpell(Cat, core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: rakeRank.ID},
	// 	SpellSchool:    rakeRank.SpellSchool(),
	// 	DefenseType:    rakeRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	ClassSpellMask: DruidSpellRake,
	// 	Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost:   rakeRank.Cost(),
	// 		Refund: 0.8,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: rakeRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	// 	MaxRange:         core.MaxMeleeRange,
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label:    "Rake",
	// 			Duration: time.Second * 9,
	// 		},
	// 		NumberOfTicks: int32(rakeRank.Duration() / rakeTick.Period()),
	// 		TickLength:    rakeTick.Period(),
	//
	// 		OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			druid.UpdateBleedPower(druid.Rake, sim, target, true, true)
	// 		},
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, rakeTick.Average(core.CharacterLevel)+0.02*dot.Spell.MeleeAttackPower(target), dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := rakeRank.DamageEffect().Average(core.CharacterLevel) + 0.01*spell.MeleeAttackPower(target)
	// 		if druid.MangleAuras != nil && druid.MangleAuras.Get(target).IsActive() {
	// 			baseDamage *= 1.3
	// 		}
	//
	// 		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
	//
	// 		if result.Landed() {
	// 			druid.AddComboPoints(sim, 1, spell.ComboPointMetrics())
	// 			spell.Dot(target).Apply(sim)
	// 		} else {
	// 			spell.IssueRefund(sim)
	// 		}
	// 	},
	//
	// 	ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
	// 		tickBase := rakeTick.Average(core.CharacterLevel) + 0.02*spell.MeleeAttackPower(target)
	// 		ticks := spell.CalcPeriodicDamage(sim, target, tickBase, spell.OutcomeExpectedMagicAlwaysHit)
	// 		attackTable := spell.Unit.AttackTables[target.UnitIndex]
	// 		critChance := spell.PhysicalCritChance(attackTable)
	// 		ticks.Damage *= 1 + critChance*(spell.CritDamageMultiplier(attackTable)-1)
	// 		return ticks
	// 	},
	// })
	//
	// druid.Rake.ShortName = "Rake"
}

func (druid *Druid) CurrentRakeCost() float64 {
	return druid.Rake.Cost.GetCurrentCost()
}
