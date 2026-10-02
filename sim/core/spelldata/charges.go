package spelldata

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// Attaches to a charged buff the listener that spends its charges, as the buff's row states it, and
// returns that listener. A cast spends one charge however many targets it hits. The spells that
// spend one are the ones the row's proc effect names; a row whose proc effect names none states them
// on its spell modifiers instead, and the charge goes to the spells the buff modifies.
func AttachChargeSpender(character *core.Character, buff *core.Aura, s *Spell, name string, opts ...ProcOpt) core.ProcTrigger {
	firstHit, resetFirstHit := core.OncePerCast()
	buff.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
		resetFirstHit()
	})

	spender := ProcTrigger(character, s, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		buff.RemoveStack(sim)
	}, opts...)
	spender.Name = name
	spender.ClassFlags = chargeClassFlags(character, s)
	spender.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return firstHit(sim, spell)
	}

	buff.AttachProcTriggerCallback(&character.Unit, spender)
	return spender
}

func chargeClassFlags(character *core.Character, s *Spell) core.ClassFlags {
	flags := rowClassFlags(s)
	if flags.IsZero() {
		flags = spellModClassFlags(s)
	}
	if othersFamily(character, flags) {
		return core.ClassFlags{}
	}
	return flags
}

func spellModClassFlags(s *Spell) core.ClassFlags {
	for i := range s.Effects {
		e := &s.Effects[i]
		if (e.Aura == dbcenums.A_ADD_FLAT_MODIFIER || e.Aura == dbcenums.A_ADD_PCT_MODIFIER) && !e.ClassFlags.IsZero() {
			return e.ClassFlags
		}
	}
	return core.ClassFlags{}
}
