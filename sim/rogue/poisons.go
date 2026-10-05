package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (rogue *Rogue) applyPoisons() {
	if rogue.Options.GetMainHandPoison() == proto.RoguePoison_None &&
		rogue.Options.GetOffHandPoison() == proto.RoguePoison_None {
		return
	}
	rogue.applyDeadlyPoison()
	rogue.applyInstantPoison()
}

var deadlyPoisonRank = spellData.DeadlyPoison.Rank(4)

func (rogue *Rogue) registerDeadlyPoisonSpell() {
	config := spelldata.SpellConfig(&rogue.Unit, deadlyPoisonRank, spelldata.Magic(core.ProcMaskSpellDamageProc))
	config.ClassSpellMask = RogueSpellDeadlyPoison
	tick := deadlyPoisonRank.PeriodicEffect()
	numTicks := deadlyPoisonRank.Duration().Seconds() / tick.Period().Seconds()
	apScaling := 0.0045 / numTicks // Confirmed by Zirene

	config.Dot = spelldata.DotConfig(deadlyPoisonRank, tick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		stacks := dot.GetStacks()
		baseTick := tick.Average()
		apFactor := dot.Spell.MeleeAttackPower(target) * apScaling
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, (baseTick+apFactor)*float64(stacks), deadlyPoisonRank.TickOutcomeHitRolled(dot))
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
		if !result.Landed() {
			return
		}

		dot := spell.Dot(target)
		if dot.IsActive() {
			dot.Refresh(sim)
			dot.AddStack(sim)
		} else {
			dot.Apply(sim)
			dot.SetStacks(sim, 1)
		}
	}

	rogue.DeadlyPoison = rogue.RegisterSpell(config)

	// The TBC implementation, kept for the port:
	// procMask := rogue.getPoisonProcMask(deadlyImbueID)
	// if procMask == core.ProcMaskUnknown {
	// 	return
	// }
	// rogue.DeadlyPoison = rogue.GetOrRegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 27187},
	// 	SpellSchool:    core.SpellSchoolNature,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamageProc,
	// 	ClassSpellMask: RogueSpellDeadlyPoison,
	// 	ClassFlags:     spellData.DeadlyPoison.Highest().ClassFlags,
	// 	Flags:          core.SpellFlagPoison | core.SpellFlagPassiveSpell | core.SpellFlagProc,
	//
	// 	DamageMultiplier:         1,
	// 	DamageMultiplierAdditive: 1,
	// 	ThreatMultiplier:         1,
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label:     "Deadly Poison",
	// 			MaxStacks: 5,
	// 			Duration:  time.Second * 12,
	// 		},
	// 		NumberOfTicks: 4,
	// 		TickLength:    time.Second * 3,
	//
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			base := 45.0 * float64(dot.GetStacks())
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, base, dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
	// 		if !result.Landed() {
	// 			return
	// 		}
	//
	// 		dot := spell.Dot(target)
	// 		if dot.IsActive() {
	// 			dot.Refresh(sim)
	// 			dot.AddStack(sim)
	// 		} else {
	// 			dot.Apply(sim)
	// 			dot.SetStacks(sim, 1)
	// 		}
	// 	},
	// })
}

var instantPoisonRank = spellData.InstantPoison.Highest()

func (rogue *Rogue) registerInstantPoisonSpell() {
	ipBaseDamage := instantPoisonRank.DamageEffect().Average()
	config := spelldata.SpellConfig(&rogue.Unit, instantPoisonRank, spelldata.Magic(core.ProcMaskSpellDamageProc))
	config.ClassSpellMask = RogueSpellInstantPoison
	apScaling := 0.005 // Confirmed by Zirene

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, ipBaseDamage+apScaling*spell.MeleeAttackPower(target), spell.OutcomeMagicHitAndCrit)
	}

	rogue.InstantPoison = rogue.RegisterSpell(config)

	// The TBC implementation, kept for the port:
	// procMask := rogue.getPoisonProcMask(instantImbueID)
	// if procMask == core.ProcMaskUnknown {
	// 	return
	// }
	// ipBaseDamage := 146.0
	// ipRange := 48
	//
	// ipConfig := core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 26890},
	// 	SpellSchool:    core.SpellSchoolNature,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamageProc,
	// 	ClassSpellMask: RogueSpellInstantPoison,
	// 	ClassFlags:     spellData.InstantPoison.Highest().ClassFlags,
	// 	Flags:          core.SpellFlagPoison | core.SpellFlagPassiveSpell | core.SpellFlagProc,
	//
	// 	DamageMultiplier:         1,
	// 	DamageMultiplierAdditive: 1,
	// 	ThreatMultiplier:         1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		spell.CalcAndDealDamage(sim, target, ipBaseDamage+sim.RandomFloat("Instant Poison")*float64(ipRange), spell.OutcomeMagicHitAndCrit)
	// 	},
	// }
	//
	// rogue.InstantPoison = rogue.RegisterSpell(ipConfig)
}

func (rogue *Rogue) getPoisonProcMask(poisonId proto.RoguePoison) core.ProcMask {
	var mask core.ProcMask
	if rogue.Options.MainHandPoison == poisonId {
		mask |= core.ProcMaskMeleeMH
	}
	if rogue.Options.OffHandPoison == poisonId {
		mask |= core.ProcMaskMeleeOH
	}
	return mask
}

func (rogue *Rogue) applyDeadlyPoison() {
	procMask := rogue.getPoisonProcMask(proto.RoguePoison_DeadlyPoison)
	if procMask == core.ProcMaskUnknown {
		return
	}

	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:                "Deadly Poison",
		Outcome:             core.OutcomeLanded,
		Callback:            core.CallbackOnSpellHitDealt,
		TriggerImmediately:  true,
		ProcMask:            procMask,
		IsWeaponProc:        true,
		ProcSpellClassFlags: deadlyPoisonRank.ClassFlags,
		DPM:                 rogue.NewFixedProcChanceManager(0.3, procMask),
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			rogue.DeadlyPoison.Cast(sim, result.Target)
		},
	})
}

func (rogue *Rogue) applyInstantPoison() {
	procMask := rogue.getPoisonProcMask(proto.RoguePoison_InstantPoison)
	if procMask == core.ProcMaskUnknown {
		return
	}

	rogue.MakeProcTriggerAura(core.ProcTrigger{
		Name:                "Instant Poison",
		Outcome:             core.OutcomeLanded,
		Callback:            core.CallbackOnSpellHitDealt,
		TriggerImmediately:  true,
		ProcMask:            procMask,
		IsWeaponProc:        true,
		ProcSpellClassFlags: instantPoisonRank.ClassFlags,
		DPM:                 rogue.NewFixedProcChanceManager(0.2, procMask), //Confirmed in game 9/30/2026
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			rogue.InstantPoison.Cast(sim, result.Target)
		},
	})
}
