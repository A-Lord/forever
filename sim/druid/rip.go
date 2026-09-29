package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerRipSpell() {
	rank := spellData.Rip.Highest()
	tick := rank.PeriodicDamageEffect()
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = druid.requiresComboPoints(rank)
	config.Dot = spelldata.DotConfig(rank, tick, bleedAuraOpt())
	// Combo points are captured when the dot is applied. Attack power is read on the tick.
	// TODO: Ingame research needed -- whether Rip snapshots attack power.
	config.Dot.OnSnapshot = func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
		cp := float64(druid.ComboPoints())
		dot.SnapshotBaseDamage = tick.AverageAt(druid.level()) + float64(tick.PointsPerResource)*cp
		if druid.ripComboPoints == nil {
			druid.ripComboPoints = map[int32]float64{}
		}
		druid.ripComboPoints[dot.Unit.UnitIndex] = cp
		_ = sim
	}
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		cp := druid.ripComboPoints[target.UnitIndex]
		apScale := cp
		if apScale > 4 {
			apScale = 4
		}
		damage := dot.SnapshotBaseDamage + classicAPPerComboPointRip*apScale*dot.Spell.MeleeAttackPower(target)
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, rank.TickOutcome(dot))
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		druid.breakProwl(sim)
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
			druid.spendComboPoints(sim, spell)
		}
		druid.refundOnMiss(sim, spell, result, rank)
	}
	druid.Rip = druid.RegisterSpell(config)
	druid.Rip.ShortName = "Rip"
}

func (druid *Druid) CurrentRipCost() float64 {
	return druid.Rip.Cost.GetCurrentCost()
}
