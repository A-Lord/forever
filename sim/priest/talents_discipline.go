package priest

func (priest *Priest) registerDisciplineTalents() {
	// Tier 1
	priest.applyPowerInLight()
	priest.applyWandSpecialization()
	priest.applyTwinDisciplines()

	// Tier 2
	priest.applySilentResolve()
	priest.applyHolyPrecision()
	priest.applyImprovedPowerWordShield()
	priest.applyMartyrdom()

	// Tier 3
	priest.applyMentalAgility()
	priest.applyInnerFocus()
	priest.applyMeditation()

	// Tier 4
	priest.applyImprovedInnerFire()
	priest.applyMentalStrength()
	priest.applySoulWarding()
	priest.applyImprovedManaBurn()

	// Tier 5
	priest.applyPenance()
	priest.applyRenewedHope()

	// Tier 6
	priest.applyDivineAegis()

	// Tier 7
	priest.applyPowerInfusion()
}

// applyPowerInLight implements Power in Light, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyPowerInLight() {
	if priest.Talents.PowerInLight == 0 {
		return
	}
}

// applyWandSpecialization implements Wand Specialization, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyWandSpecialization() {
	if priest.Talents.WandSpecialization == 0 {
		return
	}
}

// applyTwinDisciplines implements Twin Disciplines, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyTwinDisciplines() {
	if priest.Talents.TwinDisciplines == 0 {
		return
	}
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's brought back.
func (priest *Priest) applySilentResolve() {
	if priest.Talents.SilentResolve == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if priest.Talents.SilentResolve == 0 {
	// 	return
	// }
	// // -4% threat per rank for discipline and holy spells
	// threatReduction := []float64{0, -0.04, -0.08, -0.12, -0.16, -0.20}[priest.Talents.SilentResolve]
	// priest.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_ThreatMultiplier_Pct,
	// 	FloatValue: threatReduction,
	// 	ClassMask:  PriestHolySpells,
	// })
}

// applyHolyPrecision implements Holy Precision, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyHolyPrecision() {
	if priest.Talents.HolyPrecision == 0 {
		return
	}
}

// applyImprovedPowerWordShield implements Improved Power Word: Shield, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyImprovedPowerWordShield() {
	if priest.Talents.ImprovedPowerWordShield == 0 {
		return
	}
}

// applyMartyrdom implements Martyrdom, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyMartyrdom() {
	if priest.Talents.Martyrdom == 0 {
		return
	}
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's brought back.
func (priest *Priest) applyMentalAgility() {
	if priest.Talents.MentalAgility == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if priest.Talents.MentalAgility == 0 {
	// 	return
	// }
	//
	// priest.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_PowerCost_Pct_Add,
	// 	FloatValue: -0.02 * float64(priest.Talents.MentalAgility),
	// 	ClassMask:  PriestSpellInstant,
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's brought back.
func (priest *Priest) applyInnerFocus() {
	if !priest.Talents.InnerFocus {
		return
	}

	// The TBC implementation, kept for the port:
	// if !priest.Talents.InnerFocus {
	// 	return
	// }
	//
	// critMod := priest.AddDynamicMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_BonusCrit_Percent,
	// 	FloatValue: 25.0,
	// 	ClassMask:  PriestSpellsAll,
	// })
	//
	// var innerFocusSpell *core.Spell
	// priest.InnerFocusAura = priest.RegisterAura(core.Aura{
	// 	Label:    "Inner Focus",
	// 	ActionID: core.ActionID{SpellID: 14751},
	// 	Duration: time.Hour,
	// 	OnGain: func(aura *core.Aura, sim *core.Simulation) {
	// 		aura.Unit.PseudoStats.SpellCostPercentModifier -= 100
	// 		critMod.Activate()
	// 	},
	// 	OnExpire: func(aura *core.Aura, sim *core.Simulation) {
	// 		aura.Unit.PseudoStats.SpellCostPercentModifier += 100
	// 		critMod.Deactivate()
	// 		innerFocusSpell.CD.Use(sim)
	// 	},
	// 	OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
	// 		if !spell.Matches(PriestSpellsAll) {
	// 			return
	// 		}
	// 		aura.Deactivate(sim)
	// 	},
	// })
	//
	// innerFocusSpell = priest.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 14751},
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	Flags:          core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,
	// 	ClassSpellMask: PriestSpellFlagNone,
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			NonEmpty: true,
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    priest.NewTimer(),
	// 			Duration: time.Second * 180,
	// 		},
	// 	},
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
	// 		priest.InnerFocusAura.Activate(sim)
	// 	},
	// 	RelatedSelfBuff: priest.InnerFocusAura,
	// })
	//
	// priest.AddMajorCooldown(core.MajorCooldown{
	// 	Spell: innerFocusSpell,
	// 	Type:  core.CooldownTypeMana,
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's brought back.
func (priest *Priest) applyMeditation() {
	if priest.Talents.Meditation == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if priest.Talents.Meditation == 0 {
	// 	return
	// }
	//
	// priest.PseudoStats.SpiritRegenRateCasting += spellData.Meditation.FractionAt(priest.Talents.Meditation)
	// priest.UpdateManaRegenRates()
}

// applyImprovedInnerFire implements Improved Inner Fire, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyImprovedInnerFire() {
	if priest.Talents.ImprovedInnerFire == 0 {
		return
	}
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's brought back.
func (priest *Priest) applyMentalStrength() {
	if priest.Talents.MentalStrength == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if priest.Talents.MentalStrength == 0 {
	// 	return
	// }
	// // +2% mana per rank
	// priest.MultiplyStat(stats.Mana, spellData.MentalStrength.MultiplierAt(priest.Talents.MentalStrength))
}

// applySoulWarding implements Soul Warding, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applySoulWarding() {
	if !priest.Talents.SoulWarding {
		return
	}
}

// applyImprovedManaBurn implements Improved Mana Burn, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyImprovedManaBurn() {
	if priest.Talents.ImprovedManaBurn == 0 {
		return
	}
}

// applyPenance implements Penance, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyPenance() {
	if !priest.Talents.Penance {
		return
	}
}

// applyRenewedHope implements Renewed Hope, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyRenewedHope() {
	if priest.Talents.RenewedHope == 0 {
		return
	}
}

// applyDivineAegis implements Divine Aegis, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyDivineAegis() {
	if priest.Talents.DivineAegis == 0 {
		return
	}
}

// TODO: To be implemented. The priest's own cast still needs a spell around the
// generated aura, which is +20% damage and healing done for 15 seconds.
func (priest *Priest) applyPowerInfusion() {
	if !priest.Talents.PowerInfusion {
		return
	}

	// The body the port needs:
	// if !priest.Talents.PowerInfusion {
	// 	return
	// }
	//
	// piAura := buffs.PowerInfusionsAura(&priest.Unit, true, 0)
	//
	// piSpell := priest.RegisterSpell(core.SpellConfig{
	// 	ActionID:    core.ActionID{SpellID: 10060},
	// 	SpellSchool: core.SpellSchoolHoly,
	// 	Flags:       core.SpellFlagHelpful,
	// 	ManaCost: core.ManaCostOptions{
	// 		BaseCostPercent: 16,
	// 	},
	// 	Cast: core.CastConfig{
	// 		CD: core.Cooldown{
	// 			Timer:    priest.NewTimer(),
	// 			Duration: buffs.PowerInfusionsCooldown(),
	// 		},
	// 		DefaultCast: core.Cast{
	// 			NonEmpty: true,
	// 		},
	// 	},
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
	// 		piAura.Activate(sim)
	// 	},
	// })
	//
	// priest.AddMajorCooldown(core.MajorCooldown{
	// 	Spell:    piSpell,
	// 	Priority: core.CooldownPriorityBloodlust,
	// 	Type:     core.CooldownTypeMana,
	// })
}
