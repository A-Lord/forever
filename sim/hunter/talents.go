package hunter

func (hunter *Hunter) ApplyTalents() {
	hunter.registerAimedShot()

	hunter.registerBeastMasteryTalents()
	hunter.registerMarksmanshipTalents()
	hunter.registerSurvivalTalents()

	if hunter.Pet != nil {
		hunter.Pet.ApplyTalents()
	}
}

var aimedShotRank = spellData.AimedShot.Highest()

// TODO: To be implemented.
func (hunter *Hunter) registerAimedShot() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// hunter.AimedShot = hunter.RegisterRangedSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: aimedShotRank.ID},
	// 	SpellSchool:    aimedShotRank.SpellSchool(),
	// 	DefenseType:    aimedShotRank.DefenseTypeCore(),
	// 	ClassSpellMask: HunterSpellAimedShot,
	// 	ProcMask:       core.ProcMaskRangedSpecial,
	// 	Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(aimedShotRank.Cost()),
	// 	},
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			CastTime: time.Millisecond * 3000,
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    hunter.NewTimer(),
	// 			Duration: max(aimedShotRank.Cooldown(), aimedShotRank.CategoryCooldown()),
	// 		},
	// 	},
	//
	// 	BonusCoefficient: aimedShotRank.DamageEffect().Coeff(),
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := 0.2*spell.RangedAttackPower(target) +
	// 			hunter.AutoAttacks.Ranged().BaseDamage(sim) +
	// 			hunter.talonOfAlarBonus() +
	// 			aimedShotRank.DamageEffect().Average(core.CharacterLevel)
	//
	// 		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)
	//
	// 		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
	// 			spell.DealDamage(sim, result)
	// 		})
	// 	},
	// })
}

// TODO: uncalled -- Forever drops the Readiness talent; re-gate before wiring back
// into ApplyTalents.
func (hunter *Hunter) registerReadiness() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// hunter.Readiness = hunter.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 23989},
	// 	SpellSchool:    core.SpellSchoolPhysical,
	// 	ClassSpellMask: HunterSpellReadiness,
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: time.Second,
	// 		},
	// 		IgnoreHaste: true,
	// 		CD: core.Cooldown{
	// 			Timer:    hunter.NewTimer(),
	// 			Duration: time.Minute * 5,
	// 		},
	// 	},
	//
	// 	ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
	// 		return !hunter.RapidFire.IsReady(sim) ||
	// 			!hunter.MultiShot.IsReady(sim) ||
	// 			!hunter.ArcaneShot.IsReady(sim) ||
	// 			!hunter.KillCommand.IsReady(sim) ||
	// 			!hunter.RaptorStrike.IsReady(sim)
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
	// 		hunter.RapidFire.CD.Reset()
	// 		hunter.MultiShot.CD.Reset()
	// 		hunter.ArcaneShot.CD.Reset()
	// 		hunter.KillCommand.CD.Reset()
	// 		hunter.RaptorStrike.CD.Reset()
	// 	},
	// })
	//
	// hunter.AddMajorCooldown(core.MajorCooldown{
	// 	Spell: hunter.Readiness,
	// 	Type:  core.CooldownTypeDPS,
	//
	// 	ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
	// 		return !hunter.RapidFire.IsReady(sim)
	// 	},
	// })
}
