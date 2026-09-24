package druid

var lacerateRank = spellData.Lacerate.Highest()
var lacerateTick = lacerateRank.PeriodicEffect()

// TODO: To be implemented.
func (druid *Druid) registerLacerateSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// tickDamageBase := lacerateTick.Average(core.CharacterLevel)
	//
	// druid.Lacerate = druid.RegisterSpell(core.SpellConfig{
	// 	ActionID:        core.ActionID{SpellID: lacerateRank.ID},
	// 	CastRequirement: lacerateRank.CastRequirement(),
	// 	SpellSchool:     lacerateRank.SpellSchool(),
	// 	DefenseType:     lacerateRank.DefenseTypeCore(),
	// 	ProcMask:        core.ProcMaskMeleeMHSpecial,
	// 	ClassSpellMask:  DruidSpellLacerate,
	// 	Flags:           core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
	//
	// 	RageCost: core.RageCostOptions{
	// 		Cost:   int32(lacerateRank.Cost()),
	// 		Refund: 0.8,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: lacerateRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 0.5,
	// 	FlatThreatBonus:  267,
	// 	MaxRange:         core.MaxMeleeRange,
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label:     "Lacerate",
	// 			MaxStacks: 5,
	// 			Duration:  time.Second * 15,
	// 		},
	// 		NumberOfTicks: int32(lacerateRank.Duration() / lacerateTick.Period()),
	// 		TickLength:    lacerateTick.Period(),
	//
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			perStack := tickDamageBase + 0.01*dot.Spell.MeleeAttackPower(target)
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, perStack*float64(dot.Aura.GetStacks()), dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := tickDamageBase + 0.01*spell.MeleeAttackPower(target)
	// 		if druid.MangleAuras != nil && druid.MangleAuras.Get(target).IsActive() {
	// 			baseDamage *= 1.3
	// 		}
	// 		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
	//
	// 		if result.Landed() {
	// 			dot := spell.Dot(target)
	// 			if dot.IsActive() {
	// 				dot.Refresh(sim)
	// 				dot.AddStack(sim)
	// 			} else {
	// 				dot.Apply(sim)
	// 				dot.SetStacks(sim, 1)
	// 			}
	// 		} else {
	// 			spell.IssueRefund(sim)
	// 		}
	// 	},
	// })
	//
	// druid.Lacerate.ShortName = "Lacerate"
}
