package rogue

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (rogue *Rogue) registerAssassinationTalents() {
	// Tier 1
	// Gouge NYI
	// Remorseless Attacks NYI
	rogue.registerMalice()

	// Tier 2
	// Ruthlessness modeled in rogue.go ApplyFinisher
	rogue.registerMurder()
	// Improved Slice modeled in Slice and Dice directly

	// Tier 3
	// Relentless Strikes modeled in rogue.go ApplyFinisher
	// Improved Expose Armor modeled in Expose Armor directly
	rogue.registerLethality()

	// Tier 4
	rogue.registerVilePoisons()
	rogue.registerColdBlood()
	rogue.registerImprovedPoisons()

	// Tier 5
	// Vigor modeled in rogue.go during initialization
	rogue.registerMutilate()
	// Kidney Shot NYI

	// Tier 6
	rogue.registerSealFate()

	// Tier 7
	rogue.registerVenom()
}

func (rogue *Rogue) registerImprovedEviscerate() {
	if rogue.Talents.ImprovedEviscerate == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.ImprovedEviscerate.Rank(rogue.Talents.ImprovedEviscerate))
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerMalice() {
	if rogue.Talents.Malice == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.Malice.Rank(rogue.Talents.Malice))
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerMurder() {
	if rogue.Talents.Murder == 0 {
		return
	}

	multiplier := spellData.Murder.MultiplierAt(rogue.Talents.Murder)
	rogue.Env.RegisterPostFinalizeEffect(func() {
		rogue.MultiplyDamageVersus(multiplier, true, proto.MobType_MobTypeHumanoid, proto.MobType_MobTypeDragonkin)
	})
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerLethality() {
	if rogue.Talents.Lethality == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.Lethality.Rank(rogue.Talents.Lethality))
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerVilePoisons() {
	if rogue.Talents.VilePoisons == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.VilePoisons.Rank(rogue.Talents.VilePoisons))
}

var coldBloodRank = spellData.ColdBlood.Highest()

func (rogue *Rogue) registerColdBlood() {
	if !rogue.Talents.ColdBlood {
		return
	}

	coldBloodAura := rogue.RegisterAura(spelldata.AuraConfig(coldBloodRank))
	spelldata.ParseEffects(&rogue.Character, coldBloodAura, coldBloodRank)
	config := spelldata.SpellConfig(&rogue.Unit, coldBloodRank)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		coldBloodAura.Activate(sim)
	}

	rogue.ColdBlood = rogue.RegisterSpell(config)

	// The TBC implementation, kept for the port:
	// if !rogue.Talents.ColdBlood {
	// 	return
	// }
	//
	// cbAura := rogue.GetOrRegisterAura(core.Aura{
	// 	Label:    "Cold Blood",
	// 	ActionID: core.ActionID{SpellID: 14177},
	// 	Duration: core.NeverExpires,
	//
	// 	OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		if spell.Matches(RogueSpellActives) {
	// 			aura.Deactivate(sim)
	// 		}
	// 	},
	// }).AttachSpellMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_BonusCrit_Percent,
	// 	ClassMask:  RogueSpellActives,
	// 	FloatValue: 100.0,
	// })
	//
	// rogue.ColdBlood = rogue.GetOrRegisterSpell(core.SpellConfig{
	// 	ActionID: core.ActionID{SpellID: 14177},
	// 	ClassFlags: spellData.ColdBlood.Highest().ClassFlags,
	//
	// 	Cast: core.CastConfig{
	// 		CD: core.Cooldown{
	// 			Timer:    rogue.NewTimer(),
	// 			Duration: time.Minute * 3,
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		cbAura.Activate(sim)
	// 	},
	// })
	//
	// rogue.AddMajorCooldown(core.MajorCooldown{
	// 	Spell: rogue.ColdBlood,
	// 	Type:  core.CooldownTypeDPS,
	// })
}

func (rogue *Rogue) registerImprovedPoisons() {
	if rogue.Talents.ImprovedPoisons == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.ImprovedPoisons.Rank(rogue.Talents.ImprovedPoisons))
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerSealFate() {
	if rogue.Talents.SealFate == 0 {
		return
	}

	sfRank := spellData.SealFate.Rank(rogue.Talents.SealFate)
	sfMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: sfRank.ID})

	trigger := spelldata.ProcTrigger(&rogue.Character, sfRank,
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Outcome.Matches(core.OutcomeCrit) {
				rogue.AddComboPoints(sim, 1, sfMetrics)
			}
		})
	trigger.Name = "Seal Fate - Trigger"
	trigger.SpellFlags = SpellFlagBuilder
	trigger.TriggerImmediately = true

	rogue.MakeProcTriggerAura(trigger)

	// The TBC implementation, kept for the port:
	// if rogue.Talents.SealFate == 0 {
	// 	return
	// }
	//
	//
	//
	// rogue.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:     "Seal Fate Trigger",
	// 	ActionID: core.ActionID{SpellID: 14195},
	// 	// Forever puts the real per-rank chance on the effect; ProcChanceAt reads a flat 100%.
	// 	ProcChance: spellData.SealFate.FractionAt(rogue.Talents.SealFate),
	// 	Callback:   core.CallbackOnSpellHitDealt,
	// 	Outcome:    core.OutcomeCrit,
	// 	SpellFlags: SpellFlagBuilder,
	// 	ICD:        time.Millisecond * 500,
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		rogue.AddComboPoints(sim, 1, sfMetrics)
	// 	},
	// })
}

var mutilateCastRank = spellData.Mutilate.Highest()

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerMutilate() {
	if !rogue.Talents.Mutilate {
		return
	}

	rogue.MutilateMH = rogue.newMutilateHitSpell(true)
	rogue.MutilateOH = rogue.newMutilateHitSpell(false)

	// TODO: Verify if Mutilate (Cast) can proc MH poisons itself (detected as a double MH poison proc)
	config := spelldata.SpellConfig(&rogue.Unit, mutilateCastRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	pointGen := mutilateCastRank.EnergizeEffect().Average(core.CharacterLevel)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHit) // Miss/Dodge/Parry/Hit
		if result.Landed() {
			rogue.AddComboPoints(sim, int32(pointGen), spell.ComboPointMetrics())
			rogue.MutilateOH.Cast(sim, target)
			rogue.MutilateMH.Cast(sim, target)
		} else {
			spell.IssueRefund(sim)
		}
		spell.DealOutcome(sim, result)
	}

	rogue.Mutilate = rogue.RegisterSpell(config)
	//
	// rogue.Mutilate = rogue.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: MutilateSpellID, Tag: 0},
	// 	SpellSchool:    core.SpellSchoolPhysical,
	// 	DefenseType:    core.DefenseTypeMelee,
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
	// 	ClassSpellMask: RogueSpellMutilate,
	// 	ClassFlags:     mutilateRank.ClassFlags,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost:   int32(mutilateRank.Cost()),
	// 		Refund: 0.8,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: mutilateRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	// 	ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
	// 		if rogue.HasDagger(core.MainHand) && rogue.HasDagger(core.OffHand) {
	// 			return true
	// 		}
	// 		return false
	// 	},
	//
	// 	ThreatMultiplier: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		rogue.BreakStealth(sim)
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHit) // Miss/Dodge/Parry/Hit
	// 		if result.Landed() {
	// 			rogue.AddComboPoints(sim, 2, spell.ComboPointMetrics())
	// 			rogue.MutilateOH.Cast(sim, target)
	// 			rogue.MutilateMH.Cast(sim, target)
	// 		} else {
	// 			spell.IssueRefund(sim)
	// 		}
	// 		spell.DealOutcome(sim, result)
	// 	},
	// })
}

var mutilateHitRank = spellData.MutilateTriggered.Highest()

func (rogue *Rogue) newMutilateHitSpell(isMH bool) *core.Spell {
	mutilateHitBaseDamage := mutilateHitRank.DamageEffect().Average(core.CharacterLevel)
	spellOpt := core.Ternary(isMH, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Melee(core.ProcMaskMeleeOHSpecial))
	config := spelldata.SpellConfig(&rogue.Unit, mutilateHitRank, spellOpt, spelldata.Proc())
	if isMH {
		config.Tag = 1
	} else {
		config.Tag = 2
	}
	config.Flags |= SpellFlagBuilder

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := mutilateHitBaseDamage
		if isMH {
			baseDamage += rogue.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
		} else {
			baseDamage += rogue.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
		}

		oldMultiplier := spell.DamageMultiplier
		// TODO: Detect any poison on the target, not just rogue poisons
		if rogue.DeadlyPoison.Dot(target).IsActive() || rogue.WoundPoisonDebuffAuras.Get(target).IsActive() {
			spell.DamageMultiplier *= 1.2
		}

		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialBlockAndCrit)
		spell.DamageMultiplier = oldMultiplier
	}

	return rogue.RegisterSpell(config)
}

var venomRank = spellData.Venom.Highest()

func (rogue *Rogue) registerVenom() {
	if !rogue.Talents.Venom {
		return
	}

	config := spelldata.SpellConfig(&rogue.Unit, venomRank, spelldata.Melee(core.ProcMaskEmpty))
	venomAura := rogue.RegisterAura(spelldata.AuraConfig(venomRank))
	spelldata.ParseEffects(&rogue.Character, venomAura, venomRank)
	venomAura.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
		rogue.additivePoisonChance += 0.1
		rogue.RefreshPoisonProcChance()
	})
	venomAura.ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
		rogue.additivePoisonChance -= 0.1
		rogue.RefreshPoisonProcChance()
	})
	config.ExtraCastCondition = rogue.FinisherExtraCastCondition

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		venomAura.Duration = venomRank.Duration() + time.Second*3*time.Duration(rogue.ComboPoints())
		rogue.ApplyFinisher(sim, spell)
		venomAura.Activate(sim)
	}

	rogue.RegisterSpell(config)
}
