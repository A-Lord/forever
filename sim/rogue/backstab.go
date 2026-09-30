package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 8 is the highest trainable. Rank 9 exists from a AQ book; acquisition TBD
var backstabRank = spellData.Backstab.Rank(8)

func (rogue *Rogue) registerBackstabSpell() {
	backstabBaseDamage := backstabRank.DamageEffect().Average()
	comboPointGen := backstabRank.EnergizeEffect().Average()
	config := spelldata.SpellConfig(&rogue.Unit, backstabRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags |= SpellFlagBuilder | SpellFlagColdBlooded
	config.ClassSpellMask = RogueSpellBackstab
	config.DamageMultiplier = backstabRank.EffectN(2).Percent()
	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return !rogue.PseudoStats.InFrontOfTarget && rogue.HasDagger(core.MainHand)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := backstabBaseDamage +
			spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))

		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialNoBlockDodgeParry)

		if result.Landed() {
			rogue.AddComboPoints(sim, int32(comboPointGen), spell.ComboPointMetrics())
			chanceForBonus := spellData.PuncturingWounds.EffectAt(2).FractionAt(rogue.Talents.PuncturingWounds)
			if sim.Proc(chanceForBonus, "Puncturing Wounds") {
				rogue.AddComboPoints(sim, 1, rogue.puncturingWoundsComboPointMetrics)
			}
		} else {
			spell.IssueRefund(sim)
		}
	}

	rogue.Backstab = rogue.RegisterSpell(config)

	// The TBC implementation, kept for the port:
	// baseDamage := backstabRank.DamageEffect().Average(core.CharacterLevel)
	// weaponDamage := 1.5
	//
	// rogue.Backstab = rogue.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: backstabRank.ID},
	// 	SpellSchool:    backstabRank.SpellSchool(),
	// 	DefenseType:    backstabRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	Flags:          core.SpellFlagMeleeMetrics | SpellFlagBuilder | core.SpellFlagAPL,
	// 	ClassSpellMask: RogueSpellBackstab,
	// 	ClassFlags:     backstabRank.ClassFlags,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost:   int32(backstabRank.Cost()),
	// 		Refund: 0.8,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: backstabRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	// 	ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
	// 		return !rogue.PseudoStats.InFrontOfTarget && rogue.HasDagger(core.MainHand)
	// 	},
	//
	// 	DamageMultiplierAdditive: weaponDamage,
	// 	DamageMultiplier:         1,
	// 	ThreatMultiplier:         1,
	//
	// 	BonusCoefficient: backstabRank.DamageEffect().Coeff(),
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		rogue.BreakStealth(sim)
	// 		baseDamage := baseDamage +
	// 			spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
	//
	// 		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
	//
	// 		if result.Landed() {
	// 			rogue.AddComboPoints(sim, 1, spell.ComboPointMetrics())
	// 		} else {
	// 			spell.IssueRefund(sim)
	// 		}
	// 	},
	// })
}
