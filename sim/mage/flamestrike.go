package mage

import (
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The core and dbcenums imports belong with the commented implementation.

// TODO: was Ranks(7, 6); Forever's Flamestrike tops out at rank 6, so only that remains.
var FlameStrikeRankMap = spelldata.Ranked(spellData.Flamestrike.Rank(6).ID)

// Flamestrike
// https://www.wowhead.com/forever/spell=10216
//
// Calls down a pillar of fire, burning all enemies within the area for X Fire damage and an additional
// Y Fire damage over 8 sec. Only one Flamestrike can be active per Mage at a time.
// TODO: To be implemented.
func (mage *Mage) registerFlamestrike(rankConfig *spelldata.Spell) {
	panic("To be implemented")

	// The ported implementation, kept until this class is done:
	// actionID := core.ActionID{SpellID: rankConfig.ID}
	//
	// // Flamestrike's periodic damage is the spell FlamestrikeTriggered casts each tick, at the
	// // same rank: each rank of Flamestrike names a different rank of that family.
	// tick := spellData.FlamestrikeTriggered.Rank(rankConfig.RankNumber()).DamageEffect()
	// tickLength := rankConfig.Effect(dbcenums.A_PERIODIC_DUMMY, 0).Period()
	//
	// spell := mage.RegisterSpell(core.SpellConfig{
	// 	ActionID:       actionID,
	// 	SpellSchool:    core.SpellSchoolFire,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: MageSpellFlamestrike,
	// 	Rank:           rankConfig.RankNumber(),
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(rankConfig.Cost()),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD:      rankConfig.GCD(),
	// 			CastTime: rankConfig.CastTime(),
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	BonusCoefficient: rankConfig.DamageEffect().Coeff(),
	// 	ThreatMultiplier: 1,
	//
	// 	Dot: core.DotConfig{
	// 		IsAOE: true,
	// 		Aura: core.Aura{
	// 			ActionID: actionID,
	// 			Label:    "Flamestrike" + mage.Label + " " + rankConfig.Rank,
	// 		},
	// 		NumberOfTicks:    int32(rankConfig.Duration() / tickLength),
	// 		TickLength:       tickLength,
	// 		BonusCoefficient: tick.Coeff(),
	// 		OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
	// 			dot.Spell.CalcAndDealPeriodicAoeDamage(sim, tick.Average(core.CharacterLevel), dot.OutcomeTickMagicHit)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
	// 		baseDamage := rankConfig.DamageEffect().Average(core.CharacterLevel)
	// 		spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
	// 		spell.AOEDot().Apply(sim)
	// 	},
	// })
	//
	// mage.Flamestrike = append(mage.Flamestrike, spell)
}
