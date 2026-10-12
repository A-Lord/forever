package shaman

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The rows Lightning Overload casts, one for each rank of Lightning Bolt and of Chain Lightning, taught
// at the level of the rank it copies. No effect of the game files names them.
var lightningBoltOverloadRanks = spelldata.Ranked(408439, 408440, 408441, 408442, 408443, 408472, 408473, 408474, 408475, 408477)
var chainLightningOverloadRanks = spelldata.Ranked(408479, 408481, 408482, 408484)

// The tag the results label an overload by.
const CastTagLightningOverload int32 = 6

// The overload row taught at the level of rank.
func overloadRank(overloads spelldata.Ladder, rank *spelldata.Spell) *spelldata.Spell {
	var found *spelldata.Spell
	overloads.Each(func(_ int32, overload *spelldata.Spell) {
		if overload.SpellLevel == rank.SpellLevel {
			found = overload
		}
	})
	if found == nil {
		panic(fmt.Sprintf("shaman: no Lightning Overload row at level %d for spell %d", rank.SpellLevel, rank.ID))
	}
	return found
}

// Lightning Bolt and Chain Lightning, and the copy of either that Lightning Overload casts. The copy is
// a row of its own, with half the coefficient and damage of its own, cast with no cast time or cost. Its
// row states no threat and lacks Not a Proc, so only listeners that can proc from procs hear its hits.
// It reports under the spell id of the rank it copies, tagged CastTagLightningOverload.
func (shaman *Shaman) newElectricSpellConfig(rank *spelldata.Spell, isOverload bool) core.SpellConfig {
	opts := []spelldata.SpellOpt{
		spelldata.Magic(core.ProcMaskSpellDamage),
		spelldata.Flags(SpellFlagShamanSpell | SpellFlagFocusable),
	}
	if isOverload {
		opts = append(opts, spelldata.Proc(), spelldata.Flags(core.SpellFlagProc))
	}

	config := spelldata.SpellConfig(&shaman.Unit, rank, opts...)
	if rank.GeneratesNoThreat() {
		config.ThreatMultiplier = 0
	}

	return config
}
