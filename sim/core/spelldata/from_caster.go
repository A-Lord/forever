package spelldata

import (
	"github.com/wowsims/forever/sim/core"
)

// The slot a parse holds on its caster's attack table against the aura's unit, for the effects that
// reach only the caster's attacks on that unit. Every damage effect the parse attaches multiplies
// into one DamageDoneByCasterExtraMultiplier handler, and every crit effect adds into one
// CritChanceByCasterBonus handler, each enabled while any of its effects is up.
type casterSlot struct {
	caster, target *core.Unit
	index, total   int

	damage, crit     []*casterPart
	damageOn, critOn bool
}

// One effect in the slot: the client's percentage per level, the level it is at and the spells it
// reaches.
type casterPart struct {
	value   float64
	level   float64
	reaches func(*core.Spell) bool
}

// An effect of a target-side aura that reaches only its caster's attacks. A damage effect is a
// percentage the client applies once per stack, so a level of 2 is worth 1 + 2*value/100; a crit
// effect is percentage points, once per stack. The static path has no aura on a target, and without
// FromCaster there is no caster, so both skip the row.
func (p *parser) fromCaster(kind string, damage bool, v float64, reaches func(*core.Spell) bool) *attachment {
	if p.casterSlot == nil || p.static {
		return nil
	}

	a := &attachment{kind: kind, value: v}
	if damage {
		a.value, a.multiplicative = percentMultiplier(v), true
	}
	if p.dry {
		return a
	}

	part := &casterPart{value: v, reaches: reaches}
	slot := p.casterSlot
	if damage {
		slot.damage = append(slot.damage, part)
	} else {
		slot.crit = append(slot.crit, part)
	}
	a.set = func(_ *core.Simulation, level float64) {
		part.level = level
		slot.update()
	}
	return a
}

// Enables or disables each handler to match whether any of its effects is up. The attack tables
// exist once the environment is finalized, which is before any aura can be gained.
func (s *casterSlot) update() {
	table := s.caster.AttackTables[s.target.UnitIndex]

	if on := anyUp(s.damage); on != s.damageOn {
		s.damageOn = on
		if on {
			core.EnableDamageDoneByCaster(s.index, s.total, table, s.damageMultiplier)
		} else {
			core.DisableDamageDoneByCaster(s.index, table)
		}
	}

	if on := anyUp(s.crit); on != s.critOn {
		s.critOn = on
		if on {
			core.EnableCritChanceByCaster(s.index, s.total, table, s.critChance)
		} else {
			core.DisableCritChanceByCaster(s.index, table)
		}
	}
}

func (s *casterSlot) damageMultiplier(_ *core.Simulation, spell *core.Spell, _ *core.AttackTable) float64 {
	mult := 1.0
	for _, part := range s.damage {
		if part.level > 0 && part.reaches(spell) {
			mult *= percentMultiplier(part.value * part.level)
		}
	}
	return mult
}

func (s *casterSlot) critChance(spell *core.Spell, _ *core.AttackTable) float64 {
	bonus := 0.0
	for _, part := range s.crit {
		if part.level > 0 && part.reaches(spell) {
			bonus += part.value * part.level
		}
	}
	return bonus
}

func anyUp(parts []*casterPart) bool {
	for _, part := range parts {
		if part.level > 0 {
			return true
		}
	}
	return false
}
