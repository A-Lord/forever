package buffs

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The buff row states the five-rank amount, which the talent's ranks scale the way the priest's own
// copy does.
var darkInspirationTalent = spelldata.Talent(1326159, 5)

// The party's copy is the providing priest's talent at the rank the party states, up for the fight.
func driveDarkInspiration(char *core.Character, party *proto.PartyBuffs) {
	m := darkInspirationMeta
	ranks := min(party.DarkInspiration, darkInspirationTalent.Len())
	opts := append(m.Options(0), spelldata.ScaledByRanks(ranks, darkInspirationTalent.Len()))
	core.MakePermanent(m.buff(&char.Unit, false, m.label(false), 0, opts))
}
