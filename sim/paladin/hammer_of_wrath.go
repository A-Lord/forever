package paladin

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
)

var HammerOfWrathRankMap = spellData.HammerOfWrath

// Hammer of Wrath
// https://www.wowhead.com/forever/spell=24239
//
// Hurls a hammer that strikes an enemy for 498 Holy damage. Only usable on enemies that have 20%
// or less health.
func (paladin *Paladin) registerHammerOfWrath(row shared.SpellData) {
	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: row.SpellID},
		SpellSchool:    row.SpellSchool,
		DefenseType:    row.DefenseType,
		ProcMask:       core.ProcMaskRangedSpecial,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHammerOfWrath,
		Rank:           row.Rank,
		MaxRange:       row.MaxRange,
		MissileSpeed:   row.MissileSpeed,

		ManaCost: manaCost(row),
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				// The client's 1s GCD sits at core's floor, so it is named as the floor too or
				// GCDTime clamps it.
				GCDMin:   row.GCD,
				GCD:      row.GCD,
				CastTime: row.CastTime,
			},
			CD: core.Cooldown{
				Timer:    paladin.sharedTimer(&paladin.hammerOfWrathTimer),
				Duration: row.Cooldown,
			},
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				castTime := paladin.ApplyCastSpeedForSpell(cast.CastTime, spell)
				paladin.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+castTime)
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: row.Direct.BonusCoefficient(),

		ExtraCastCondition: func(sim *core.Simulation, _ *core.Unit) bool {
			return sim.IsExecutePhase20()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcDamage(sim, target, row.Direct.Damage(sim), spell.OutcomeRangedHitAndCrit)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	})
}
