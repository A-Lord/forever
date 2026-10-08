package warlock

import (
	"github.com/wowsims/forever/sim/core"
)

// Soul Shards (item 6265) are a reagent in SpellReagents. Of the spells the sim casts, Soul Fire (row 20520,
// spell 17924) and Shadowburn (row 20511, spell 18871) take one each; the spell store carries no reagents, so
// the spells call hasSoulShard and spendSoulShard themselves. Each iteration starts with the shards carried
// into the fight. The summons and the Create *stone spells take one too, and are cast before the fight.
// No shard comes back from a kill, Drain Soul's or Shadowburn's within 8 sec, as the target does not die.
//
// The shards are the stacks of an aura on the warlock, so the timeline and the APL's aura stack values read
// them. With none left the aura falls off, and a refund brings it back.
func (warlock *Warlock) registerSoulShards() {
	carried := warlock.Options.SoulShards
	warlock.SoulShardsAura = warlock.RegisterAura(core.Aura{
		Label:     "Soul Shards",
		ActionID:  core.ActionID{ItemID: 6265},
		Duration:  core.NeverExpires,
		MaxStacks: carried,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			if carried > 0 {
				aura.Activate(sim)
				aura.SetStacks(sim, carried)
			}
		},
	})
}

// Casting needs a shard, checked where the sim checks mana: when the cast starts and when it completes.
func (warlock *Warlock) hasSoulShard(_ *core.Simulation, _ *core.Unit) bool {
	return warlock.SoulShardsAura.GetStacks() > 0
}

// Called first in ApplyEffects, which the sim runs when the cast completes, right after it spends the mana:
// the reagent goes then too.
func (warlock *Warlock) spendSoulShard(sim *core.Simulation) {
	warlock.SoulShardsAura.RemoveStack(sim)
}

func (warlock *Warlock) gainSoulShard(sim *core.Simulation) {
	if !warlock.SoulShardsAura.IsActive() {
		warlock.SoulShardsAura.Activate(sim)
	}
	warlock.SoulShardsAura.AddStack(sim)
}
