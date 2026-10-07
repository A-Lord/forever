package warlock

import (
	"github.com/wowsims/forever/sim/core"
)

// Soul Shards (item 6265) are a reagent in SpellReagents. Of the spells the sim casts, Soul Fire (row 20520,
// spell 17924) and Shadowburn (row 20511, spell 18871) take one each; the spell store carries no reagents, so
// the spells call hasSoulShard and spendSoulShard themselves. Each iteration starts with the shards carried
// into the fight. The summons and the Create *stone spells take one too, and are cast before the fight.
// No shard comes back from a kill, Drain Soul's or Shadowburn's within 8 sec, as the target does not die.
func (warlock *Warlock) registerSoulShards() {
	warlock.RegisterResetEffect(func(_ *core.Simulation) {
		warlock.soulShards = warlock.Options.SoulShards
	})
}

// Casting needs a shard, checked where the sim checks mana: when the cast starts and when it completes.
func (warlock *Warlock) hasSoulShard(_ *core.Simulation, _ *core.Unit) bool {
	return warlock.soulShards > 0
}

// Called first in ApplyEffects, which the sim runs when the cast completes, right after it spends the mana:
// the reagent goes then too.
func (warlock *Warlock) spendSoulShard(sim *core.Simulation) {
	warlock.soulShards--
	if sim.Log != nil {
		warlock.Log(sim, "Spent a Soul Shard, %d left", warlock.soulShards)
	}
}

func (warlock *Warlock) gainSoulShard(sim *core.Simulation) {
	warlock.soulShards++
	if sim.Log != nil {
		warlock.Log(sim, "Gained a Soul Shard, %d left", warlock.soulShards)
	}
}
