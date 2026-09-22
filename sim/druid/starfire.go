package druid

import (
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: was Ranks(6, 8); Forever's Starfire tops out at rank 7, so the max-rank entry
// moves down rather than naming a rank the table does not hold.
var StarfireRankMap = spelldata.Ranked(spellData.Starfire.Rank(6).ID, spellData.Starfire.Rank(7).ID)

// TODO: To be implemented.
func (druid *Druid) registerStarfireSpell(rankConfig *spelldata.Spell) {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// spell := druid.RegisterSpell(Humanoid|Moonkin, core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: rankConfig.ID},
	// 	SpellSchool:    core.SpellSchoolArcane,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	ClassSpellMask: DruidSpellStarfire,
	// 	Flags:          core.SpellFlagAPL,
	// 	Rank:           rankConfig.RankNumber(),
	// 	MaxRange:       float64(rankConfig.MaxRange),
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: rankConfig.Cost(),
	// 	},
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD:      rankConfig.GCD(),
	// 			CastTime: rankConfig.CastTime(),
	// 		},
	// 	},
	//
	// 	BonusCoefficient: rankConfig.DamageEffect().Coeff(),
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := rankConfig.DamageEffect().Average(core.CharacterLevel)
	// 		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	// 	},
	// })
	//
	// druid.Starfire = append(druid.Starfire, spell)
}
