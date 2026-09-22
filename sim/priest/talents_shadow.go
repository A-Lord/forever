package priest

import (
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerShadowTalents() {
	// Tier 1
	priest.applyShadowFocus()
	priest.applyBlackout()
	priest.applySpiritTap()

	// Tier 2
	priest.applyShadowAffinity()
	priest.applyImprovedShadowWordPain()
	priest.applyShadowReach()

	// Tier 3
	priest.applyImprovedMindBlast()
	priest.applyImprovedPsychicScream()
	priest.applyMindFlay()
	priest.applyImprovedMindFlay()

	// Tier 4
	priest.applyImprovedFade()
	priest.applyVampiricEmbrace()
	priest.applyShadowWeaving()

	// Tier 5
	priest.applySilence()
	priest.applyDevouringContagion()

	// Tier 6
	priest.applyEarlyDemise()
	priest.applyDarkness()

	// Tier 7
	priest.applyShadowform()
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (priest *Priest) applyShadowFocus() {
	if priest.Talents.ShadowFocus == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if priest.Talents.ShadowFocus == 0 {
	// 	return
	// }
	//
	// priest.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexShadow] += spellData.ShadowFocus.ValueAt(priest.Talents.ShadowFocus)
	//
}

// applyBlackout implements Blackout, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyBlackout() {
	if priest.Talents.Blackout == 0 {
		return
	}
}

// applySpiritTap implements Spirit Tap, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applySpiritTap() {
	if priest.Talents.SpiritTap == 0 {
		return
	}
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (priest *Priest) applyShadowAffinity() {
	if priest.Talents.ShadowAffinity == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if priest.Talents.ShadowAffinity == 0 {
	// 	return
	// }
	//
	// threatReduction := []float64{0, -0.08, -0.16, -0.25}[priest.Talents.ShadowAffinity]
	//
	// priest.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_ThreatMultiplier_Pct,
	// 	FloatValue: threatReduction,
	// 	ClassMask:  PriestShadowSpells,
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (priest *Priest) applyImprovedShadowWordPain() {
	if priest.Talents.ImprovedShadowWordPain == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if priest.Talents.ImprovedShadowWordPain == 0 {
	// 	return
	// }
	//
	// priest.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_DotNumberOfTicks_Flat,
	// 	IntValue:  int32(priest.Talents.ImprovedShadowWordPain),
	// 	ClassMask: PriestSpellShadowWordPain,
	// })
}

// applyShadowReach implements Shadow Reach, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyShadowReach() {
	if priest.Talents.ShadowReach == 0 {
		return
	}
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (priest *Priest) applyImprovedMindBlast() {
	if priest.Talents.ImprovedMindBlast == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if priest.Talents.ImprovedMindBlast == 0 {
	// 	return
	// }
	//
	// priest.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_Cooldown_Flat,
	// 	TimeValue: time.Millisecond * time.Duration(-500*priest.Talents.ImprovedMindBlast),
	// 	ClassMask: PriestSpellMindBlast,
	// })
}

// applyImprovedPsychicScream implements Improved Psychic Scream, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyImprovedPsychicScream() {
	if priest.Talents.ImprovedPsychicScream == 0 {
		return
	}
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (priest *Priest) applyMindFlay() {
	if !priest.Talents.MindFlay {
		return
	}

	// The TBC implementation, kept for the port:
	// if !priest.Talents.MindFlay {
	// 	return
	// }
	// MindFlayRankMap.Each(func(_ int32, rank *spelldata.Spell) { priest.registerMindFlaySpell(rank) })
}

var MindFlayRankMap = spellData.MindFlay

// TODO: To be implemented. Mind Flay already has a full Forever rank ladder (spellData.MindFlay); the
// TBC body needs review before it's uncommented.
func (priest *Priest) registerMindFlaySpell(rank *spelldata.Spell) {
	// The TBC implementation, kept for the port:
	// tick := rank.PeriodicEffect()
	// tickLength := tick.Period()
	//
	// priest.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: rank.ID},
	// 	SpellSchool:    core.SpellSchoolShadow,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagChanneled | core.SpellFlagAPL,
	// 	ClassSpellMask: PriestSpellMindFlay,
	// 	Rank:           rank.RankNumber(),
	// 	MaxRange:       float64(rank.MaxRange),
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(rank.Cost()),
	// 	},
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: rank.GCD(),
	// 		},
	// 	},
	//
	// 	DamageMultiplier:         1,
	// 	DamageMultiplierAdditive: 1,
	// 	ThreatMultiplier:         1,
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label: fmt.Sprintf("MindFlay-%d", rank.RankNumber()),
	// 		},
	// 		NumberOfTicks:        int32(rank.Duration() / tickLength),
	// 		TickLength:           tickLength,
	// 		AffectedByCastSpeed:  true,
	// 		HasteReducesDuration: true,
	// 		BonusCoefficient:     tick.Coeff(),
	//
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, tick.Average(core.CharacterLevel), dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
	// 		if result.Landed() {
	// 			spell.Dot(target).Apply(sim)
	// 		}
	// 	},
	//
	// 	ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
	// 		return spell.CalcPeriodicDamage(sim, target, tick.Average(core.CharacterLevel), spell.OutcomeExpectedMagicHit)
	// 	},
	// })
}

// applyImprovedMindFlay implements Improved Mind Flay, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyImprovedMindFlay() {
	if priest.Talents.ImprovedMindFlay == 0 {
		return
	}
}

// applyImprovedFade implements Improved Fade, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyImprovedFade() {
	if priest.Talents.ImprovedFade == 0 {
		return
	}
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (priest *Priest) applyVampiricEmbrace() {
	if !priest.Talents.VampiricEmbrace {
		return
	}

	// The TBC implementation, kept for the port:
	// if !priest.Talents.VampiricEmbrace {
	// 	return
	// }
	//
	// // TODO: Forever drops Improved Vampiric Embrace; base heal percent only until we know
	// // whether the bonus moved onto another talent.
	// healPct := 0.15
	// healthMetrics := priest.NewHealthMetrics(core.ActionID{SpellID: 15286})
	//
	// veDebuffAuras := priest.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
	// 	aura := target.RegisterAura(core.Aura{
	// 		Label:    "Vampiric Embrace",
	// 		ActionID: core.ActionID{SpellID: 15286},
	// 		Duration: time.Second * 60,
	// 	})
	// 	aura.AttachProcTriggerCallback(target, core.ProcTrigger{
	// 		Name:               "Vampiric Embrace Proc",
	// 		Callback:           core.CallbackOnSpellHitTaken | core.CallbackOnPeriodicDamageTaken,
	// 		ClassSpellMask:     PriestShadowSpells,
	// 		RequireDamageDealt: true,
	// 		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 			priest.GainHealth(sim, result.Damage*healPct, healthMetrics)
	// 		},
	// 	})
	// 	return aura
	// })
	//
	// priest.VampiricEmbrace = priest.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 15286},
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskEmpty,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: PriestSpellVampiricEmbrace,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		BaseCostPercent: 2,
	// 	},
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: core.GCDDefault,
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    priest.NewTimer(),
	// 			Duration: time.Second * 10,
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		veDebuffAuras.Get(target).Activate(sim)
	// 	},
	//
	// 	RelatedAuraArrays: veDebuffAuras.ToMap(),
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (priest *Priest) applyShadowWeaving() {
	if priest.Talents.ShadowWeaving == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if priest.Talents.ShadowWeaving == 0 {
	// 	return
	// }
	//
	// swAuras := priest.NewEnemyAuraArray(core.ShadowWeavingAura)
	//
	// priest.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:             "Shadow Weaving Trigger",
	// 	CanProcFromProcs: true, // 15257, 15331-15334 carry the bit.
	// 	ClassSpellMask:   PriestShadowSpells,
	// 	Callback:         core.CallbackOnSpellHitDealt,
	// 	Outcome:          core.OutcomeLanded,
	// 	ProcChance:       spellData.ShadowWeaving.FractionAt(priest.Talents.ShadowWeaving),
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		swAuras.Get(result.Target).Activate(sim)
	// 		swAuras.Get(result.Target).AddStack(sim)
	// 	},
	// })
}

// applySilence implements Silence, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applySilence() {
	if !priest.Talents.Silence {
		return
	}
}

// applyDevouringContagion implements Devouring Contagion, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyDevouringContagion() {
	if priest.Talents.DevouringContagion == 0 {
		return
	}
}

// applyEarlyDemise implements Early Demise, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyEarlyDemise() {
	if priest.Talents.EarlyDemise == 0 {
		return
	}
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (priest *Priest) applyDarkness() {
	if priest.Talents.Darkness == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if priest.Talents.Darkness == 0 {
	// 	return
	// }
	//
	// priest.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: spellData.Darkness.Effect(dbcenums.A_MOD_DAMAGE_PERCENT_DONE, 32).FractionAt(priest.Talents.Darkness),
	// 	ClassMask:  PriestShadowSpells,
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (priest *Priest) applyShadowform() {
	if !priest.Talents.Shadowform {
		return
	}

	// The TBC implementation, kept for the port:
	// if !priest.Talents.Shadowform {
	// 	return
	// }
	//
	// shadowformAura := priest.RegisterAura(core.Aura{
	// 	Label:    "Shadowform",
	// 	ActionID: core.ActionID{SpellID: 15473},
	// 	Duration: core.NeverExpires,
	// 	OnReset: func(aura *core.Aura, sim *core.Simulation) {
	// 		if priest.SelfBuffs.PreShadowform {
	// 			aura.Activate(sim)
	// 		}
	// 	},
	// 	// Casting any holy-school spell breaks Shadowform.
	// 	OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
	// 		if spell.SpellSchool.Matches(core.SpellSchoolHoly) {
	// 			aura.Deactivate(sim)
	// 		}
	// 	},
	// }).AttachSpellMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: 0.15,
	// 	ClassMask:  PriestShadowSpells,
	// }).AttachMultiplicativePseudoStatBuff(
	// 	&priest.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical], 0.85,
	// )
	//
	// priest.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 15473},
	// 	SpellSchool:    core.SpellSchoolShadow,
	// 	ProcMask:       core.ProcMaskEmpty,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: PriestSpellShadowform,
	// 	ManaCost: core.ManaCostOptions{
	// 		BaseCostPercent: 32,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: core.GCDDefault,
	// 		},
	// 	},
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
	// 		shadowformAura.Activate(sim)
	// 	},
	// })
}
