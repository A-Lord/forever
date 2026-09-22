package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Devouring Plague - Undead Racial
// Shadow school DoT, 3 min cooldown, 24s duration

var DevouringPlagueRankMap = spellData.DevouringPlague

// TODO: To be implemented. Devouring Plague already has a full Forever rank ladder
// (spellData.DevouringPlague); the TBC body needs review before it's uncommented.
func (priest *Priest) registerDevouringPlagueSpell(rank *spelldata.Spell, cdTimer *core.Timer) {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// tick := rank.PeriodicEffect()
	// tickLength := tick.Period()
	//
	// healthMetrics := priest.NewHealthMetrics(core.ActionID{SpellID: rank.ID}.WithTag(1))
	//
	// priest.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: rank.ID},
	// 	SpellSchool:    core.SpellSchoolShadow,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: PriestSpellDevouringPlague,
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
	// 	ThreatMultiplier:         1,
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label: fmt.Sprintf("DevouringPlague-%d", rank.RankNumber()),
	// 			OnInit: func(aura *core.Aura, sim *core.Simulation) {
	// 				aura.AttachProcTrigger(core.ProcTrigger{
	// 					Name:               "DevouringPlague-Heal",
	// 					Callback:           core.CallbackOnPeriodicDamageTaken,
	// 					ClassSpellMask:     PriestSpellDevouringPlague,
	// 					RequireDamageDealt: true,
	// 					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 						priest.GainHealth(sim, result.Damage, healthMetrics)
	// 					},
	// 				})
	// 			},
	// 		},
	// 		NumberOfTicks:       int32(rank.Duration() / tickLength),
	// 		TickLength:          tickLength,
	// 		AffectedByCastSpeed: false,
	// 		BonusCoefficient:    tick.Coeff(),
	//
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, tick.Average(core.CharacterLevel), dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
	// 		if result.Landed() {
	// 			spell.Dot(target).Apply(sim)
	// 			spell.Dot(target).TickOnce(sim)
	// 		}
	// 		spell.DealOutcome(sim, result)
	// 	},
	//
	// 	ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
	// 		return spell.CalcPeriodicDamage(sim, target, tick.Average(core.CharacterLevel), spell.OutcomeExpectedMagicHit)
	// 	},
	// })
}
