package priest

import (
	"github.com/wowsims/forever/sim/core/spelldata"
)

var SmiteRankMap = spellData.Smite

// TODO: To be implemented. Smite already has a full Forever rank ladder (spellData.Smite); the TBC body
// needs review before it's uncommented.
func (priest *Priest) registerSmiteSpell(rank *spelldata.Spell) {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// priest.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: rank.ID},
	// 	SpellSchool:    core.SpellSchoolHoly,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: PriestSpellSmite,
	// 	Rank:           rank.RankNumber(),
	// 	MaxRange:       float64(rank.MaxRange),
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(rank.Cost()),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD:      rank.GCD(),
	// 			CastTime: rank.CastTime(),
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	BonusCoefficient: rank.DamageEffect().Coeff(),
	// 	ThreatMultiplier: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := rank.DamageEffect().Average(core.CharacterLevel)
	// 		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	// 	},
	// })
}
