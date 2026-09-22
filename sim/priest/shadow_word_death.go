package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var ShadowWordDeathRankMap = spellData.ShadowWordDeath

// TODO: To be implemented. Shadow Word: Death already has a full Forever rank ladder
// (spellData.ShadowWordDeath); the TBC body needs review before it's uncommented.
func (priest *Priest) registerShadowWordDeathSpell(rank *spelldata.Spell, cdTimer *core.Timer) {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// priest.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: rank.ID},
	// 	SpellSchool:    core.SpellSchoolShadow,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: PriestSpellShadowWordDeath,
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
	// 		CD: core.Cooldown{
	// 			Timer:    cdTimer,
	// 			Duration: max(rank.Cooldown(), rank.CategoryCooldown()),
	// 		},
	// 	},
	//
	// 	DamageMultiplier:         1,
	// 	DamageMultiplierAdditive: 1,
	// 	BonusCoefficient:         rank.DamageEffect().Coeff(),
	// 	ThreatMultiplier:         1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := rank.DamageEffect().Average(core.CharacterLevel)
	// 		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	// 		spell.DealDamage(sim, result)
	//
	// 		// Set to always remove for purpose of sim
	// 		priest.RemoveHealth(sim, result.Damage)
	// 	},
	// })
}
