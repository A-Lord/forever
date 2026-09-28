package buffs

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The armor share is on the talent's curve, one value per rank; the buff row states only the
// duration and the rank 1 share.
var inspirationTalent = spelldata.Talent(14892, 3)
var inspirationBuff = spelldata.MustFind(14893)

func InspirationAura(unit *core.Unit, points int32) *core.Aura {
	multiplier := 1 + inspirationTalent.Rank(points).EffectN(1).Percent()

	armorDep := unit.NewDynamicMultiplyStat(stats.Armor, multiplier)

	return unit.GetOrRegisterAura(spelldata.AuraConfig(inspirationBuff)).AttachStatDependency(armorDep)
}

func ApplyInspiration(character *core.Character, uptime float64) {
	if uptime <= 0 {
		return
	}
	uptime = min(1, uptime)

	inspirationAura := InspirationAura(&character.Unit, 3)

	ApplyFixedUptimeAura(inspirationAura, uptime, time.Millisecond*2500, 1)
}
