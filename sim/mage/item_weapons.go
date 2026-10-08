package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func init() {
	// Atiesh, Greatstaff of the Guardian
	// https://www.wowhead.com/forever/item=22589/atiesh-greatstaff-of-the-guardian
	//
	// Increases the spell critical chance of all party members within 30 yards by 2%.
	//
	// 28142 is a party aura, and the wearer is in its own party. The party buff's Atiesh - Mage count is
	// the staves the other party members carry, as for the other three staves, whose party auras reach
	// their wearer through the item's stats.
	core.NewItemEffect(22589, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(28142)
		aura := core.MakePermanent(character.RegisterAura(spelldata.AuraConfig(row)))
		spelldata.ParseEffects(character, aura, row)
		character.ItemSwap.RegisterProc(22589, aura)
	})
}
