package paladin

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
)

var ConsecrationRankMap = spellData.Consecration

// Consecration
// https://www.wowhead.com/forever/spell=20924
//
// Consecrates the land beneath the Paladin, doing X Holy damage over 8 sec to enemies who enter the
// area. The first 4 enemies who enter the area will take an additional Y damage over 8 sec.
func (paladin *Paladin) registerConsecration(row shared.SpellData) {
	tick := row.Periodic.(shared.SpellDataPeriodic)

	// The extra damage the first few targets take, which is the only part of the spell the client
	// gives a spell power coefficient: the tick everyone takes has none.
	bonus := row.SecondaryPeriodic.(shared.SpellDataPeriodic)
	bonusTargets := int(row.Effect(shared.A_PERIODIC_DUMMY, 0).Value)

	// The bonus scales on its own coefficient, so it is added to the base damage here rather than
	// through the dot's, which is the base tick's. Consecrated Ground marks the same targets.
	dealTick := func(sim *core.Simulation, dot *core.Dot) {
		for i, target := range sim.Encounter.ActiveTargetUnits {
			damage := tick.Tick
			if i < bonusTargets {
				damage += bonus.Tick + bonus.Coef*dot.Spell.BonusDamage(dot.Spell.Unit.AttackTables[target.UnitIndex])
				if paladin.consecratedGroundAuras != nil {
					paladin.consecratedGroundAuras.Get(target).Activate(sim)
				}
			}
			dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, dot.OutcomeTickMagicHit)
		}
	}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: row.SpellID},
		SpellSchool:    row.SpellSchool,
		DefenseType:    row.DefenseType,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskConsecration,
		Rank:           row.Rank,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		MaxRange: 8,

		ManaCost: manaCost(row),
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: row.GCD,
			},
			CD: core.Cooldown{
				Timer:    paladin.sharedTimer(&paladin.consecrationTimer),
				Duration: row.Cooldown,
			},
		},

		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				ActionID: core.ActionID{SpellID: row.SpellID},
				Label:    "Consecration" + paladin.Label + " " + row.GetRankLabel(),
			},
			NumberOfTicks:    tick.NumberOfTicks - 1, // the sim adds an immediate tick below
			TickLength:       tick.TickLength,
			BonusCoefficient: tick.Coef,
			OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
				dealTick(sim, dot)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// Consecration does one hit check on cast but the ground effect will still be applied
			// meaning it's only needed to proc things like Eye of Magtheridon (procs on resist)
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)

			dot := spell.AOEDot()
			dot.Apply(sim)
			dealTick(sim, dot)
		},
	})
}
