package spelldata

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// Attaches to a charged buff the listener that spends its charges, as the buff's row states it, and
// returns that listener. A cast spends one charge however many targets it hits. The spells that
// spend one are the ones the row's proc effect names; a row whose proc effect names none states them
// on its spell modifiers instead, and the charge goes to the spells the buff modifies.
//
// A channel is one cast: the charge stays up for every tick and every target, and is spent when the
// channel ends.
// TODO: A tick that lands after its channel ends, a missile still in flight, misses the charge.
func AttachChargeSpender(character *core.Character, buff *core.Aura, s *Spell, name string, opts ...ProcOpt) core.ProcTrigger {
	firstHit, resetFirstHit := core.OncePerCast()
	channelHit := false
	buff.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
		resetFirstHit()
		channelHit = false
	})

	hooked := map[*core.Dot]bool{}
	spendAtChannelEnd := func(channel *core.Dot) {
		channelHit = true
		if hooked[channel] {
			return
		}
		hooked[channel] = true
		channel.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
			if channelHit && buff.IsActive() {
				channelHit = false
				buff.RemoveStack(sim)
			}
		})
	}

	spender := ProcTrigger(character, s, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		buff.RemoveStack(sim)
	}, opts...)
	spender.Name = name
	spender.ClassFlags = chargeClassFlags(character, s)
	spender.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		// A channel's last tick can land from its own expiry, too late to hook; it spends like a cast.
		if channel := character.ChanneledDot; channel != nil && (channel.IsActive() || hooked[channel]) {
			spendAtChannelEnd(channel)
			return false
		}
		if spell.Flags.Matches(core.SpellFlagChanneled) {
			// The channel's own hit lands before its dot starts; its ticks hook the dot.
			channelHit = true
			return false
		}
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
