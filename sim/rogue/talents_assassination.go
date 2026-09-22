package rogue

import (
	"github.com/wowsims/forever/sim/core"
)

func (rogue *Rogue) registerAssassinationTalents() {
	// Tier 1
	rogue.registerImprovedEviscerate()
	// Remorseless Attacks not implemented
	rogue.registerMalice()

	// Tier 2
	// Ruthless implemented in ApplyFinisher
	rogue.registerMurder()
	rogue.registerPuncturingWounds()

	// Tier 3
	// Relentless Strikes implemented in ApplyFinisher
	rogue.registerImprovedExposeArmor()
	rogue.registerLethality()

	// Tier 4
	rogue.registerVilePoisons()
	// Improved Poisons implemented in poisons.go

	// Tier 5
	// Fleet Footed NYI
	rogue.registerColdBlood()
	// Improved Kidney NYI
	// Quick Recovery implemented in individual finisher EnergyCostOptions

	// Tier 6
	rogue.registerSealFate()

	// Tier 7
	// Vigor implemented in rogue.go
	// Deadened Nerves NYI

	// Tier 9
	rogue.registerMutilate()

	// Forever additions, not yet implemented.
	rogue.registerImprovedKidneyShot()
	rogue.registerRemorselessAttacks()
	rogue.registerVenom()
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerImprovedEviscerate() {
	if rogue.Talents.ImprovedEviscerate == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.ImprovedEviscerate == 0 {
	// 	return
	// }
	//
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  RogueSpellEviscerate,
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: spellData.ImprovedEviscerate.FractionAt(rogue.Talents.ImprovedEviscerate),
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerMalice() {
	if rogue.Talents.Malice == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.Malice == 0 {
	// 	return
	// }
	//
	// rogue.AddStat(stats.PhysicalCritPercent, float64(rogue.Talents.Malice))
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerMurder() {
	if rogue.Talents.Murder == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.Murder == 0 {
	// 	return
	// }
	//
	// var multiplier float64 = spellData.Murder.MultiplierAt(rogue.Talents.Murder)
	// rogue.Env.RegisterPostFinalizeEffect(func() {
	// 	for _, at := range rogue.AttackTables {
	// 		if slices.Contains([]proto.MobType{proto.MobType_MobTypeHumanoid, proto.MobType_MobTypeGiant, proto.MobType_MobTypeBeast, proto.MobType_MobTypeDragonkin}, at.Defender.MobType) {
	// 			at.DamageDealtMultiplier *= multiplier
	// 			at.CritMultiplier *= multiplier
	// 		}
	// 	}
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerPuncturingWounds() {
	if rogue.Talents.PuncturingWounds == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.PuncturingWounds == 0 {
	// 	return
	// }
	//
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_BonusCrit_Percent,
	// 	ClassMask:  RogueSpellBackstab,
	// 	FloatValue: 10.0 * float64(rogue.Talents.PuncturingWounds),
	// })
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_BonusCrit_Percent,
	// 	ClassMask: RogueSpellMutilateHit,
	// 	// The proc trigger sits between the two crit modifiers, at effect position 2; the
	// 	// Mutilate crit bonus is the second of the two, at position 3, and they share an aura
	// 	// and misc so have to be indexed rather than named.
	// 	FloatValue: spellData.PuncturingWounds.EffectAt(3).ValueAt(rogue.Talents.PuncturingWounds),
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerImprovedExposeArmor() {
	if rogue.Talents.ImprovedExposeArmor == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.ImprovedExposeArmor == 0 {
	// 	return
	// }
	//
	// // The bonus of Imp EA is handled inside of Expose Armor in debuffs.go
	//
	// // Create a dummy aura for APL handling
	// core.MakePermanent(rogue.RegisterAura(core.Aura{
	// 	Label:    "Improved Expose Armor",
	// 	ActionID: core.ActionID{SpellID: 14168},
	// }))
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerLethality() {
	if rogue.Talents.Lethality == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.Lethality == 0 {
	// 	return
	// }
	//
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_CritMultiplier_Flat,
	// 	ClassMask:  RogueSpellLethality,
	// 	FloatValue: spellData.Lethality.FractionAt(rogue.Talents.Lethality),
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerVilePoisons() {
	if rogue.Talents.VilePoisons == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.VilePoisons == 0 {
	// 	return
	// }
	//
	// rogue.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	ClassMask:  RogueSpellPoisons,
	// 	FloatValue: spellData.VilePoisons.Effect(dbcenums.A_ADD_PCT_MODIFIER, spelldata.SPELLMOD_DAMAGE).FractionAt(rogue.Talents.VilePoisons),
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerColdBlood() {
	if !rogue.Talents.ColdBlood {
		return
	}

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

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerSealFate() {
	if rogue.Talents.SealFate == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if rogue.Talents.SealFate == 0 {
	// 	return
	// }
	//
	// sfMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: 14195})
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

// Was 34413; Forever reworked Mutilate onto an entirely new set of spell ids, so this
// follows the highest rank the client actually ships.
var MutilateSpellID int32 = spellData.Mutilate.Highest().ID

var mutilateRank = spellData.Mutilate.ByID(MutilateSpellID)

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) registerMutilate() {
	if !rogue.Talents.Mutilate {
		return
	}

	// The TBC implementation, kept for the port:
	// if !rogue.Talents.Mutilate {
	// 	return
	// }
	//
	// rogue.MutilateMH = rogue.newMutilateHitSpell(true)
	// rogue.MutilateOH = rogue.newMutilateHitSpell(false)
	//
	// rogue.Mutilate = rogue.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: MutilateSpellID, Tag: 0},
	// 	SpellSchool:    core.SpellSchoolPhysical,
	// 	DefenseType:    core.DefenseTypeMelee,
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
	// 	ClassSpellMask: RogueSpellMutilate,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost:   mutilateRank.Cost(),
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

func (rogue *Rogue) newMutilateHitSpell(isMH bool) *core.Spell {
	actionID := core.ActionID{SpellID: MutilateSpellID, Tag: 1}
	procMask := core.ProcMaskMeleeMHSpecial
	if !isMH {
		actionID = core.ActionID{SpellID: MutilateSpellID, Tag: 2}
		procMask = core.ProcMaskMeleeOHSpecial
	}
	mutBaseDamage := 101.0

	return rogue.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       procMask,
		Flags:          core.SpellFlagMeleeMetrics | SpellFlagBuilder,
		ClassSpellMask: RogueSpellMutilateHit,

		DamageMultiplier:         1,
		DamageMultiplierAdditive: 1,
		ThreatMultiplier:         1,

		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := mutBaseDamage
			if isMH {
				baseDamage += spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			} else {
				baseDamage += spell.Unit.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			}

			oldMultiplier := spell.DamageMultiplier
			if rogue.DeadlyPoison.Dot(target).IsActive() || rogue.WoundPoisonDebuffAuras.Get(target).IsActive() {
				spell.DamageMultiplier += 0.5
			}

			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialBlockAndCrit)
			spell.DamageMultiplier = oldMultiplier
		},
	})
}

// registerImprovedKidneyShot implements Improved Kidney Shot, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerImprovedKidneyShot() {
	if rogue.Talents.ImprovedKidneyShot == 0 {
		return
	}
}

// registerRemorselessAttacks implements Remorseless Attacks, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerRemorselessAttacks() {
	if rogue.Talents.RemorselessAttacks == 0 {
		return
	}
}

// registerVenom implements Venom, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (rogue *Rogue) registerVenom() {
	if !rogue.Talents.Venom {
		return
	}
}
