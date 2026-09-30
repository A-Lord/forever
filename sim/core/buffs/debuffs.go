package buffs

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// The debuffs the raid puts on a target: the generated rows, then the one the game files state in a
// shape no manifest row can carry.
func applyDebuffs(target *core.Unit, debuffs *proto.Debuffs, raid *proto.Raid) {
	applyGeneratedDebuffs(target, debuffs, raid)

	if debuffs.JudgementOfTheCrusader {
		core.MakePermanent(NewSpellDataJudgementOfTheCrusaderAuraRank(target, JudgementOfTheCrusaderMaxRank))
	}
}

// A hunter's own Scorpid Sting, one per hunter, tagged with the tag its stings share.
func NewSpellDataScorpidStingAuraFor(target *core.Unit, hunter *core.Unit, stingTag string) *core.Aura {
	return newCasterDebuff(target, scorpidStingMeta, hunter, stingTag, 0)
}
