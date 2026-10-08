package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The PvP gloves' equip spell 23300: a flat 20 on Judgement of the Crusader's Holy damage bonus.
var crusaderJudgementIncrease = spelldata.MustFind(23300).EffectN(1).Average()

func init() {
	// Knight-Lieutenant's Lamellar Gauntlets
	// https://www.wowhead.com/forever/item=16410/knight-lieutenant-s-lamellar-gauntlets
	//
	// Increases the Holy damage bonus of your Judgement of the Crusader by 20.
	core.NewItemEffect(16410, raiseJudgementOfTheCrusader)

	// Marshal's Lamellar Gloves
	// https://www.wowhead.com/forever/item=16471/marshal-s-lamellar-gloves
	//
	// Increases the Holy damage bonus of your Judgement of the Crusader by 20.
	core.NewItemEffect(16471, raiseJudgementOfTheCrusader)
}

func raiseJudgementOfTheCrusader(agent core.Agent) {
	agent.(PaladinAgent).GetPaladin().judgementOfTheCrusaderBonus += crusaderJudgementIncrease
}
