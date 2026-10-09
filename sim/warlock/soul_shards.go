package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

const soulShardItemID = 6265

// The reagents the sim counts, each as the stacks of an aura on the warlock, so the timeline and the APL's
// aura stack values read them. Of the reagents the warlock's rows name, only the Soul Shard is counted: the
// Infernal Stone (5565, Inferno) and the Demonic Figurine (16583, Ritual of Doom) belong to spells the sim
// does not cast. Each iteration starts with the shards carried into the fight. No shard comes back from a
// kill, Drain Soul's or Shadowburn's within 8 sec, as the target does not die. With none left the aura
// falls off, and a refund brings it back.
func (warlock *Warlock) registerSoulShards() {
	carried := warlock.Options.SoulShards
	warlock.SoulShardsAura = warlock.RegisterAura(core.Aura{
		Label:     "Soul Shards",
		ActionID:  core.ActionID{ItemID: soulShardItemID},
		Duration:  core.NeverExpires,
		MaxStacks: carried,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			if carried > 0 {
				aura.Activate(sim)
				aura.SetStacks(sim, carried)
			}
		},
	})
	warlock.reagentAuras = map[int32]*core.Aura{soulShardItemID: warlock.SoulShardsAura}
}

// An effect that lets the spells its class flags name cast without their reagents while its aura is up.
type reagentWaiver struct {
	aura  *core.Aura
	flags core.ClassFlags
}

// The cast needs and spends the counted reagents its row states. They are checked where the sim checks
// mana, when the cast starts and when it completes, and spent first in ApplyEffects, which the sim runs
// when the cast completes, right after it spends the mana. Wraps the config's ExtraCastCondition and
// ApplyEffects, so it is called after both are set.
func (warlock *Warlock) useReagents(config *core.SpellConfig, row *spelldata.Spell) {
	if len(row.Reagents) == 0 {
		return
	}

	waived := func() bool {
		for _, waiver := range warlock.reagentWaivers {
			if waiver.aura.IsActive() && waiver.flags.Matches(row.ClassFlags) {
				return true
			}
		}
		return false
	}

	extraCastCondition := config.ExtraCastCondition
	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		if extraCastCondition != nil && !extraCastCondition(sim, target) {
			return false
		}
		if waived() {
			return true
		}
		for _, reagent := range row.Reagents {
			if aura := warlock.reagentAuras[reagent.Item]; aura != nil && aura.GetStacks() < int32(reagent.Count) {
				return false
			}
		}
		return true
	}

	applyEffects := config.ApplyEffects
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if !waived() {
			for _, reagent := range row.Reagents {
				if aura := warlock.reagentAuras[reagent.Item]; aura != nil {
					aura.RemoveStacks(sim, int32(reagent.Count))
				}
			}
		}
		applyEffects(sim, target, spell)
	}
}

func (warlock *Warlock) gainSoulShard(sim *core.Simulation) {
	if !warlock.SoulShardsAura.IsActive() {
		warlock.SoulShardsAura.Activate(sim)
	}
	warlock.SoulShardsAura.AddStack(sim)
}
