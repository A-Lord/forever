package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var sinisterStrikeRank = spellData.SinisterStrike.Highest()

func (rogue *Rogue) registerSinisterStrikeSpell() {
	sinisterStrikeBaseDamage := sinisterStrikeRank.DamageEffect().Average()
	config := spelldata.SpellConfig(&rogue.Unit, sinisterStrikeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags |= SpellFlagBuilder | SpellFlagColdBlooded
	config.ClassSpellMask = RogueSpellSinisterStrike

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := sinisterStrikeBaseDamage +
			spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))

		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if result.Landed() {
			rogue.AddComboPoints(sim, int32(sinisterStrikeRank.ProcEnergizeEffect().Roll(sim)), spell.ComboPointMetrics())

		} else {
			spell.IssueRefund(sim)
		}
	}

	rogue.SinisterStrike = rogue.RegisterSpell(config)

	//
	// rogue.SinisterStrike = rogue.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: sinisterStrikeRank.ID},
	// 	SpellSchool:    sinisterStrikeRank.SpellSchool(),
	// 	DefenseType:    sinisterStrikeRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskMeleeMHSpecial,
	// 	Flags:          core.SpellFlagMeleeMetrics | SpellFlagBuilder | core.SpellFlagAPL,
	// 	ClassSpellMask: RogueSpellSinisterStrike,
	// 	ClassFlags:     sinisterStrikeRank.ClassFlags,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost:   int32(sinisterStrikeRank.Cost()),
	// 		Refund: 0.8,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: sinisterStrikeRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	//
	// 	DamageMultiplier:         1,
	// 	DamageMultiplierAdditive: 1,
	// 	ThreatMultiplier:         1,
	//
	// 	BonusCoefficient: sinisterStrikeRank.DamageEffect().Coeff(),
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
