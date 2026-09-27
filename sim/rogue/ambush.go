package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var ambushRank = spellData.Ambush.Highest()

func (rogue *Rogue) registerAmbushSpell() {
	ambushBaseDamage := ambushRank.DamageEffect().Average(core.CharacterLevel)
	comboPointGen := ambushRank.EnergizeEffect().Average(core.CharacterLevel)
	config := spelldata.SpellConfig(&rogue.Unit, ambushRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags |= SpellFlagBuilder

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := ambushBaseDamage +
			spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))

		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialNoBlockDodgeParry)

		if result.Landed() {
			rogue.AddComboPoints(sim, int32(comboPointGen), spell.ComboPointMetrics())
		} else {
			spell.IssueRefund(sim)
		}
	}

	rogue.Ambush = rogue.RegisterSpell(config)

	// The TBC implementation, kept for the port:
	// baseDamage := ambushRank.DamageEffect().Average(core.CharacterLevel)
	// weaponDamage := 2.75
	//
	// rogue.Ambush = rogue.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: ambushRank.ID},
	// 	SpellSchool:    ambushRank.SpellSchool(),
	// 	DefenseType:    ambushRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	Flags:          core.SpellFlagMeleeMetrics | SpellFlagBuilder | core.SpellFlagAPL,
	// 	ClassSpellMask: RogueSpellAmbush,
	// 	ClassFlags:     ambushRank.ClassFlags,
	// 	MaxRange:       core.MaxMeleeRange,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost:   int32(ambushRank.Cost()),
	// 		Refund: 0,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: ambushRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	// 	ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
	// 		return !rogue.PseudoStats.InFrontOfTarget && rogue.IsStealthed()
	// 	},
	//
	// 	DamageMultiplier:         weaponDamage,
	// 	DamageMultiplierAdditive: 1,
	// 	ThreatMultiplier:         1,
	//
	// 	BonusCoefficient: ambushRank.DamageEffect().Coeff(),
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		rogue.BreakStealth(sim)
	// 		baseDamage := baseDamage +
	// 			spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
	//
	// 		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialNoBlockDodgeParry)
	//
	// 		if result.Landed() {
	// 			rogue.AddComboPoints(sim, 1, spell.ComboPointMetrics())
	// 		} else {
	// 			spell.IssueRefund(sim)
	// 		}
	// 	},
	// })
}
