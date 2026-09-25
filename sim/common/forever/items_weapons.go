package forever

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func init() {
	// Graverobber's Shovel
	//
	// Equip 1319047: 12 attack power and 12 ranged attack power, which the item's own stats carry,
	// and effects 3 and 4, 12 of each against Undead.
	core.NewItemEffect(273817, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(1319047)
		aura := core.MakePermanent(character.RegisterAura(spelldata.AuraConfig(row)))
		spelldata.ParseEffects(character, aura, row, spelldata.Effects(3, 4))
		character.ItemSwap.RegisterProc(273817, aura)
	})
}
