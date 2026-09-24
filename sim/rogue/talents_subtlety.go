package rogue

// Package-level state the commented-out implementations used:
// var hemorrhageRank = spellData.Hemorrhage.ByID(26864)

func (rogue *Rogue) registerSubtletyTalents() {
	// Tier 1
	// Master of Deception NYI
	rogue.registerOpportunity()

	// Tier 2
	// None in this tier implemented

	// Tier 3
	rogue.registerInitiative()
	rogue.registerGhostlyStrike()
	rogue.registerImprovedAmbush()

	// Tier 4
	// Setup NYI
	rogue.registerElusiveness()
	rogue.registerSerratedBlades()

	// Tier 5
	// Heightened Senses NYI
	rogue.registerPreparation()
	rogue.registerDirtyDeeds()
	rogue.registerHemorrhage()

	// Tier 7
	// Enveloping Shadows NYI
	rogue.registerPremeditation()
	// Cheat Death NYI

	// Forever additions, not yet implemented.
	rogue.registerCamouflage()
	rogue.registerCutthroat()
	rogue.registerDirtyTricks()
	rogue.registerHeightenedSenses()
	rogue.registerImprovedDistract()
	rogue.registerMasterOfDeception()
	rogue.registerQuietus()
	rogue.registerSetup()
	rogue.registerThousandCuts()
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerOpportunity() {
	if rogue.Talents.Opportunity == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.Opportunity == 0 {
	// 	return
	// }
	//
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	ClassMask:  RogueSpellBackstab | RogueSpellMutilate | RogueSpellAmbush,
	// 	FloatValue: spellData.Opportunity.Effect(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_DAMAGE)).FractionAt(rogue.Talents.Opportunity),
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerInitiative() {
	if rogue.Talents.Initiative == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.Initiative == 0 {
	// 	return
	// }
	//
	// initMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 13980})
	//
	// rogue.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:           "Initiative Trigger",
	// 	ActionID:       core.ActionID{SpellID: 13980},
	// 	ProcChance:     0.25 * float64(rogue.Talents.Initiative),
	// 	Callback:       core.CallbackOnSpellHitDealt,
	// 	Outcome:        core.OutcomeLanded,
	// 	ClassSpellMask: RogueSpellGarrote | RogueSpellAmbush,
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		rogue.AddComboPoints(sim, 1, initMetrics)
	// 	},
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerGhostlyStrike() {
	if !rogue.Talents.GhostlyStrike {
		return
	}

	// The TBC implementation, kept for the port:
	// if !rogue.Talents.GhostlyStrike {
	// 	return
	// }
	//
	// pointMetric := rogue.NewComboPointMetrics(core.ActionID{SpellID: 14278})
	// rogue.GhostlyStrike = rogue.GetOrRegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 14278},
	// 	ClassSpellMask: RogueSpellGhostlyStrike,
	// 	SpellSchool:    core.SpellSchoolPhysical,
	// 	DefenseType:    core.DefenseTypeMelee,
	// 	Flags:          core.SpellFlagAPL | core.SpellFlagMeleeMetrics,
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	MaxRange:       core.MaxMeleeRange,
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: time.Second,
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    rogue.NewTimer(),
	// 			Duration: time.Second * 20,
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost:   40,
	// 		Refund: 0.8,
	// 	},
	//
	// 	DamageMultiplier: 1.25,
	// 	ThreatMultiplier: 1,
	//
	// 	BonusCoefficient: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		rogue.BreakStealth(sim)
	//
	// 		// Dodge Aura NYI
	//
	// 		baseDamage := rogue.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
	// 		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
	// 		if result.Landed() {
	// 			rogue.AddComboPoints(sim, 1, pointMetric)
	// 		}
	// 	},
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerImprovedAmbush() {
	if rogue.Talents.ImprovedAmbush == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.ImprovedAmbush == 0 {
	// 	return
	// }
	//
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_BonusCrit_Percent,
	// 	ClassMask:  RogueSpellAmbush,
	// 	FloatValue: spellData.ImprovedAmbush.ValueAt(rogue.Talents.ImprovedAmbush),
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerElusiveness() {
	if rogue.Talents.Elusiveness == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.Elusiveness == 0 {
	// 	return
	// }
	//
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_Cooldown_Flat,
	// 	ClassMask: RogueSpellVanish,
	// 	TimeValue: time.Second * 45 * time.Duration(rogue.Talents.Elusiveness),
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerSerratedBlades() {
	if rogue.Talents.SerratedBlades == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.SerratedBlades == 0 {
	// 	return
	// }
	//
	// rogue.AddStat(stats.ArmorPenetration, 186*float64(rogue.Talents.SerratedBlades))
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	ClassMask:  RogueSpellRupture,
	// 	FloatValue: spellData.SerratedBlades.Effect(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_DOT)).FractionAt(rogue.Talents.SerratedBlades),
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerPreparation() {
	if !rogue.Talents.Preparation {
		return
	}

	// The TBC implementation, kept for the port:
	// if !rogue.Talents.Preparation {
	// 	return
	// }
	//
	// rogue.Preparation = rogue.GetOrRegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 14185},
	// 	ClassSpellMask: RogueSpellPreparation,
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: time.Second,
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    rogue.NewTimer(),
	// 			Duration: time.Minute * 10,
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		if rogue.ColdBlood != nil {
	// 			rogue.ColdBlood.CD.Set(0)
	// 		}
	// 		if rogue.Shadowstep != nil {
	// 			rogue.Shadowstep.CD.Set(0)
	// 		}
	// 		if rogue.Premeditation != nil {
	// 			rogue.Premeditation.CD.Set(0)
	// 		}
	// 		if rogue.Vanish != nil {
	// 			rogue.Vanish.CD.Set(0)
	// 		}
	// 	},
	// })
	//
	// rogue.AddMajorCooldown(core.MajorCooldown{
	// 	Spell: rogue.Preparation,
	// 	Type:  core.CooldownTypeDPS,
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerDirtyDeeds() {
	if rogue.Talents.DirtyDeeds == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.DirtyDeeds == 0 {
	// 	return
	// }
	//
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_PowerCost_Flat,
	// 	ClassMask: RogueSpellGarrote,
	// 	IntValue:  -10 * rogue.Talents.DirtyDeeds,
	// })
	//
	// ddAura := rogue.GetOrRegisterAura(core.Aura{
	// 	Label:    "Dirty Deeds",
	// 	ActionID: core.ActionID{SpellID: 14083},
	// 	Duration: core.NeverExpires,
	// }).AttachSpellMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	ClassMask:  RogueSpellsAll,
	// 	FloatValue: 0.1 * float64(rogue.Talents.DirtyDeeds),
	// })
	//
	// rogue.RegisterResetEffect(func(sim *core.Simulation) {
	// 	ddAura.Deactivate(sim)
	// 	sim.RegisterExecutePhaseCallback(func(sim *core.Simulation, isExecute int32) {
	// 		if isExecute == 35 {
	// 			ddAura.Activate(sim)
	// 		}
	// 	})
	// })
}

// TODO: To be implemented. The ability exists: spell 16511 on the Subtlety line. No rank subtext, so no
// generated table -- pin the id directly.
func (rogue *Rogue) registerHemorrhage() {
	if !rogue.Talents.Hemorrhage {
		return
	}

	// The TBC implementation, kept for the port:
	// if !rogue.Talents.Hemorrhage {
	// 	return
	// }
	//
	// pointMetric := rogue.NewComboPointMetrics(core.ActionID{SpellID: hemorrhageRank.ID})
	// rogue.Hemorrhage = rogue.GetOrRegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: hemorrhageRank.ID},
	// 	ClassSpellMask: RogueSpellHemorrhage,
	// 	SpellSchool:    hemorrhageRank.SpellSchool(),
	// 	DefenseType:    hemorrhageRank.DefenseTypeCore(),
	// 	Flags:          core.SpellFlagAPL | core.SpellFlagMeleeMetrics | SpellFlagBuilder,
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	MaxRange:       core.MaxMeleeRange,
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: hemorrhageRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost:   int32(hemorrhageRank.Cost()),
	// 		Refund: 0.8,
	// 	},
	//
	// 	DamageMultiplier: 1.1,
	// 	ThreatMultiplier: 1,
	//
	// 	BonusCoefficient: hemorrhageRank.DamageEffect().Coeff(),
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		rogue.BreakStealth(sim)
	//
	// 		baseDamage := rogue.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
	// 		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
	// 		if result.Landed() {
	// 			rogue.AddComboPoints(sim, 1, pointMetric)
	// 		}
	// 	},
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerPremeditation() {
	if !rogue.Talents.Premeditation {
		return
	}

	// The TBC implementation, kept for the port:
	// if !rogue.Talents.Premeditation {
	// 	return
	// }
	//
	// comboMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 14183})
	// shouldTimeout := false
	//
	// premedAura := rogue.RegisterAura(core.Aura{
	// 	Label:    "Premed Timeout Aura",
	// 	Duration: time.Second * 10,
	//
	// 	OnGain: func(aura *core.Aura, sim *core.Simulation) {
	// 		shouldTimeout = true
	// 	},
	// 	OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
	// 		if spell.Flags.Matches(SpellFlagFinisher) && spell.ClassSpellMask == RogueSpellSliceAndDice {
	// 			shouldTimeout = false
	// 		}
	// 	},
	// 	OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		if spell.Flags.Matches(SpellFlagFinisher) && result.Landed() {
	// 			shouldTimeout = false
	// 		}
	// 	},
	// 	OnExpire: func(aura *core.Aura, sim *core.Simulation) {
	// 		// Remove 2 points because no finisher was casted
	// 		if shouldTimeout {
	// 			rogue.AddComboPoints(sim, -2, comboMetrics)
	// 			shouldTimeout = false
	// 		}
	// 	},
	// })
	//
	// rogue.Premeditation = rogue.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 14183},
	// 	Flags:          core.SpellFlagAPL | core.SpellFlagNoOnCastComplete,
	// 	ClassSpellMask: RogueSpellPremeditation,
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			Cost: 0,
	// 			GCD:  0,
	// 		},
	// 		IgnoreHaste: true,
	// 		CD: core.Cooldown{
	// 			Timer:    rogue.NewTimer(),
	// 			Duration: time.Minute * 2,
	// 		},
	// 	},
	// 	ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
	// 		return rogue.IsStealthed()
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
	// 		rogue.AddComboPoints(sim, 2, comboMetrics)
	// 		premedAura.Activate(sim)
	// 	},
	// })
	//
	// rogue.AddMajorCooldown(core.MajorCooldown{
	// 	Spell:              rogue.Premeditation,
	// 	Type:               core.CooldownTypeDPS,
	// 	Priority:           core.CooldownPriorityLow,
	// 	AllowSpellQueueing: true,
	// })
}

// registerCamouflage implements Camouflage, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerCamouflage() {
	if rogue.Talents.Camouflage == 0 {
		return
	}
}

// registerCutthroat implements Cutthroat, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerCutthroat() {
	if rogue.Talents.Cutthroat == 0 {
		return
	}
}

// registerDirtyTricks implements Dirty Tricks, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerDirtyTricks() {
	if rogue.Talents.DirtyTricks == 0 {
		return
	}
}

// registerHeightenedSenses implements Heightened Senses, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerHeightenedSenses() {
	if rogue.Talents.HeightenedSenses == 0 {
		return
	}
}

// registerImprovedDistract implements Improved Distract, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerImprovedDistract() {
	if rogue.Talents.ImprovedDistract == 0 {
		return
	}
}

// registerMasterOfDeception implements Master of Deception, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerMasterOfDeception() {
	if rogue.Talents.MasterOfDeception == 0 {
		return
	}
}

// registerQuietus implements Quietus, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerQuietus() {
	if rogue.Talents.Quietus == 0 {
		return
	}
}

// registerSetup implements Setup, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerSetup() {
	if rogue.Talents.Setup == 0 {
		return
	}
}

// registerThousandCuts implements Thousand Cuts, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerThousandCuts() {
	if !rogue.Talents.ThousandCuts {
		return
	}
}
