package druid

var demoralizingRoarRank = spellData.DemoralizingRoar.Highest()

// TODO: To be implemented.
func (druid *Druid) registerDemoralizingRoarSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// druid.registerDemoralizingRoarAura()
	//
	// druid.DemoralizingRoar = druid.RegisterSpell(Bear, core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: demoralizingRoarRank.ID},
	// 	SpellSchool:    demoralizingRoarRank.SpellSchool(),
	// 	DefenseType:    demoralizingRoarRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskEmpty,
	// 	ClassSpellMask: DruidSpellDemoralizingRoar,
	// 	Flags:          core.SpellFlagAPL,
	//
	// 	RageCost: core.RageCostOptions{
	// 		Cost: demoralizingRoarRank.Cost(),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: demoralizingRoarRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	//
	// 	ThreatMultiplier: 1,
	// 	FlatThreatBonus:  62 * 2,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		for _, aoeTarget := range druid.Env.Encounter.AllTargetUnits {
	// 			result := spell.CalcOutcome(sim, aoeTarget, spell.OutcomeMeleeSpecialHit)
	// 			if result.Landed() {
	// 				druid.DemoralizingRoarAuras.Get(aoeTarget).Activate(sim)
	// 			}
	// 		}
	// 	},
	// })
}

// TODO: To be implemented.
func (druid *Druid) registerDemoralizingRoarAura() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// druid.DemoralizingRoarAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
	// 	// TODO: Forever drops Feral Aggression; untalented (0 points) until we know
	// 	// whether the effect moved onto another talent.
	// 	return core.DemoralizingRoarAura(target, 0)
	// })
}
