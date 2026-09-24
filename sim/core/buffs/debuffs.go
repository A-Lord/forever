package buffs

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// The debuffs the raid puts on a target: the generated rows, then the two the client states in a
// shape no manifest row can carry.
func applyDebuffs(target *core.Unit, debuffs *proto.Debuffs, raid *proto.Raid) {
	applyGeneratedDebuffs(target, debuffs, raid)

	if debuffs.JudgementOfTheCrusader {
		core.MakePermanent(JudgementOfTheCrusaderAura(target, JudgementOfTheCrusaderMaxRank))
	}

	if debuffs.Mangle {
		core.MakePermanent(MangleAura(target))
	}
}

func MangleAura(target *core.Unit) *core.Aura {
	multiplier := 1.3

	aura := target.GetOrRegisterAura(core.Aura{
		Label:    "Mangle",
		ActionID: core.ActionID{SpellID: 33876},
		Duration: time.Second * 12,
	})

	aura.NewExclusiveEffect("Mangle", true, core.ExclusiveEffect{
		Priority: multiplier,
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.PseudoStats.PeriodicPhysicalDamageTakenMultiplier *= ee.Priority
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.PseudoStats.PeriodicPhysicalDamageTakenMultiplier /= ee.Priority
		},
	})

	return aura
}
